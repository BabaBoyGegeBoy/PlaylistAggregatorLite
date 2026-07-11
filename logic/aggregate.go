package logic

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"

	"PlaylistAggregator/misc/models"

	"github.com/mozillazg/go-pinyin"
)

const aggregateMaxConcurrency = 5

// Aggregate 解析多个歌单链接，跨平台去重、排序后聚合成一个歌单。
// urls: 每行一个链接；detailed/format/order 透传给 Discover 与格式/排序处理。
// 排序：先按歌曲在几个源中出现的次数降序，次数相同再按歌名拼音/字母升序。
func Aggregate(rawURLs []string, detailed bool, format string, order string) (*models.AggregateResult, error) {
	urls := splitURLs(rawURLs)
	if len(urls) == 0 {
		return nil, errors.New("请提供至少一个歌单链接")
	}

	sem := make(chan struct{}, aggregateMaxConcurrency)
	var wg sync.WaitGroup
	type taskResult struct {
		url string
		sl  *models.SongList
		err error
	}
	results := make([]taskResult, len(urls))

	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sl, err := Discover(u, detailed)
			results[i] = taskResult{url: u, sl: sl, err: err}
		}(i, u)
	}
	wg.Wait()

	sources := make([]models.AggregateSource, 0, len(urls))
	type songEntry struct {
		original string
		nameKey  string
		count    int
		sig      songSig
		item     models.SongItem
	}
	// 按核心歌名分组；组内再按歌手兼容性（子集/别名/占位通配）匹配
	normMap := make(map[string][]*songEntry)
	totalRaw := 0

	for _, r := range results {
		src := models.AggregateSource{URL: r.url, OK: r.err == nil}
		if r.err != nil {
			src.Error = r.err.Error()
			sources = append(sources, src)
			continue
		}
		if r.sl != nil {
			src.Platform = r.sl.Platform
			src.PlatformName = r.sl.PlatformName
			src.Count = r.sl.SongsCount
		}
		sources = append(sources, src)

		// 套用统一格式，保证所有源比较口径一致
		applyFormat(r.sl, format)
		for i, s := range r.sl.Songs {
			totalRaw++
			sig := songSignature(s, format)
			item := models.SongItem{}
			if i < len(r.sl.SongsDetail) {
				item = r.sl.SongsDetail[i]
			}
			merged := false
			for _, e := range normMap[sig.core] {
				if artistCompatible(e.sig, sig) {
					e.count++
					if sig.priority > e.sig.priority {
						// 用优先级更高的那条作为代表（保留更完整/更规范的写法）
						e.original = s
						e.nameKey = pinyinKey(sig.sortName)
						e.sig = sig
						e.item = item
					}
					merged = true
					break
				}
			}
			if !merged {
				normMap[sig.core] = append(normMap[sig.core], &songEntry{
					original: s,
					nameKey:  pinyinKey(sig.sortName),
					count:    1,
					sig:      sig,
					item:     item,
				})
			}
		}
	}

	entries := make([]*songEntry, 0)
	totalUnique := 0
	for _, list := range normMap {
		for _, e := range list {
			entries = append(entries, e)
			totalUnique++
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].nameKey < entries[j].nameKey
	})

	songs := make([]string, 0, len(entries))
	detail := make([]models.SongItem, 0, len(entries))
	for _, e := range entries {
		songs = append(songs, e.original)
		detail = append(detail, e.item)
	}

	if order == "reverse" {
		for i, j := 0, len(songs)-1; i < j; i, j = i+1, j-1 {
			songs[i], songs[j] = songs[j], songs[i]
			detail[i], detail[j] = detail[j], detail[i]
		}
	}

	return &models.AggregateResult{
		Name:              "聚合歌单",
		Songs:             songs,
		SongsDetail:       detail,
		SongsCount:        len(songs),
		Sources:           sources,
		DuplicatesRemoved: totalRaw - totalUnique,
	}, nil
}

// splitURLs 将多行文本拆分为非空链接切片
func splitURLs(raw []string) []string {
	var out []string
	for _, line := range raw {
		u := strings.TrimSpace(line)
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

// 去重签名相关正则
var (
	// featRegex 匹配标题里的 (feat./ft./with X)，捕获合作艺人
	featRegex = regexp.MustCompile(`(?i)[（(](?:feat\.?|ft\.?|with)\s+([^()（）]*)[）)]`)
	// artistAliasRegex 匹配艺人字段里的括号别名，如 에일리 (AILEE)
	artistAliasRegex = regexp.MustCompile(`[（(]([^()（）]*)[）)]`)
	// sicExactRegex 匹配独立的 (sic.) 编辑注（用于保留规范写法的优先级判定）
	sicExactRegex = regexp.MustCompile(`\(sic\.\)`)
)

// artistAliasMap 艺人跨语言/错字别名归一表（小写查表）。可持续扩展。
var artistAliasMap = map[string]string{
	"瑞克和莫蒂":    "rick and morty",
	"ace specturm": "ace spectrum",
	"줄라이":      "july",
	"山岡晃":      "山冈晃",
}

// placeholderArtists 视为"未知/无约束"的占位歌手（参与通配合并，自身不保留）
var placeholderArtists = map[string]bool{
	"未知艺人": true, "未知": true, "不明": true, "不详": true,
	"unknown": true, "暂无": true, "无": true, "n/a": true, "na": true,
	"—": true, "-": true, "--": true,
}

// songSig 去重签名：核心歌名 + 归一化歌手信息 + 合并优先级。
type songSig struct {
	core          string   // 归一化核心歌名
	sortName      string   // 用于拼音排序的歌名
	artistKey     string   // 归一化歌手集（占位通配时为 ""）
	artists       []string // 归一化歌手集（有序）
	wildcard      bool     // 歌手为空或仅占位符（通配）
	canonical     bool     // 歌手无需别名映射即为规范写法
	hasSic        bool     // 标题含 (sic 编辑注
	standaloneSic bool     // 标题含独立的 (sic.) 标注
	priority      int      // 合并时保留优先级（越大越优先）
}

// songSignature 计算去重签名。核心歌名相同且歌手兼容才判为重复。
func songSignature(s string, format string) songSig {
	name, artist := splitNameArtist(s, format)
	core, featArtists, hasSic, standaloneSic := normalizeCoreName(name)
	ai := buildArtistSet(artist, featArtists)
	p := computePriority(ai.wildcard, ai.canonical, len(ai.tokens), hasSic, standaloneSic)
	return songSig{
		core:          core,
		sortName:      core,
		artistKey:     ai.key,
		artists:       ai.tokens,
		wildcard:      ai.wildcard,
		canonical:     ai.canonical,
		hasSic:        hasSic,
		standaloneSic: standaloneSic,
		priority:      p,
	}
}

// computePriority 计算合并时保留优先级：
// 占位歌手最低；其余偏好歌手更多、无需别名映射（更规范）、含 (sic.) 标注的写法。
func computePriority(wildcard, canonical bool, artistCount int, hasSic, standaloneSic bool) int {
	if wildcard {
		return 0
	}
	p := 100
	p += artistCount * 10
	if canonical {
		p += 5
	}
	if hasSic {
		p += 3
	}
	if standaloneSic {
		p += 2
	}
	return p
}

// artistCompatible 判断两组歌手是否兼容（满足其一即合并）：
// 任一方为占位通配 / 严格相等 / 一方为另一方子集。
func artistCompatible(a, b songSig) bool {
	if a.wildcard || b.wildcard {
		return true
	}
	if a.artistKey == b.artistKey {
		return true
	}
	return setSubset(a.artists, b.artists) || setSubset(b.artists, a.artists)
}

// setSubset 判断 a 是否是 b 的子集（元素集合意义）。
func setSubset(a, b []string) bool {
	if len(a) == 0 || len(a) > len(b) {
		return false
	}
	bset := make(map[string]bool, len(b))
	for _, x := range b {
		bset[x] = true
	}
	for _, x := range a {
		if !bset[x] {
			return false
		}
	}
	return true
}

// splitNameArtist 按格式把 "歌名 - 歌手" / "歌手 - 歌名" / "仅歌名" 拆开。
// 歌名内可能含 " - "，故以最后一个 " - " 为界。
func splitNameArtist(s string, format string) (name, artist string) {
	if format == "song" {
		return s, ""
	}
	parts := strings.Split(s, " - ")
	if len(parts) <= 1 {
		if format == "singer-song" {
			return "", s
		}
		return s, ""
	}
	if format == "singer-song" {
		name = parts[len(parts)-1]
		artist = strings.Join(parts[:len(parts)-1], " ")
	} else {
		name = strings.Join(parts[:len(parts)-1], " ")
		artist = parts[len(parts)-1]
	}
	return name, strings.TrimSpace(artist)
}

// normalizeCoreName 归一化歌名为核心歌名，并从标题抽取 feat/with/ft 艺人。
func normalizeCoreName(name string) (core string, featArtists []string, hasSic bool, standaloneSic bool) {
	name = toHalfWidth(strings.ToLower(name))
	name = normalizeSeparators(name) // A: 〜/～/／ 统一为 /
	name, featArtists = extractFeat(name)
	hasSic = strings.Contains(name, "(sic")
	standaloneSic = sicExactRegex.MatchString(name)
	core = stripAllBrackets(name)
	core = normalizeInstrumental(core) // E: instrumentals -> instrumental
	core = collapseSpaces(core)
	return core, featArtists, hasSic, standaloneSic
}

// normalizeSeparators 把歌名里的波浪号/全角斜杠统一为半角斜杠（同一首歌的不同标题分隔符）
func normalizeSeparators(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '〜', '～', '／':
			return '/'
		default:
			return r
		}
	}, s)
}

// normalizeInstrumental 把复数 instrumentals 归一为 instrumental（保留词内部归一）
func normalizeInstrumental(s string) string {
	return strings.ReplaceAll(s, "instrumentals", "instrumental")
}

// extractFeat 从标题抽取 (feat./ft./with X) 并删除该括号，X 作为合作艺人返回。
func extractFeat(name string) (string, []string) {
	var arts []string
	name = featRegex.ReplaceAllStringFunc(name, func(m string) string {
		if sub := featRegex.FindStringSubmatch(m); len(sub) > 1 {
			arts = append(arts, sub[1])
		}
		return " "
	})
	return name, arts
}

// stripAllBrackets 依次处理各类括号，按类型去留。
func stripAllBrackets(s string) string {
	s = stripBrackets(s, '(', ')')
	s = stripBrackets(s, '（', '）')
	s = stripBrackets(s, '[', ']')
	s = stripBrackets(s, '【', '】')
	s = stripBrackets(s, '《', '》')
	return s
}

// stripBrackets 处理指定配对括号：保留重度改编标注，去除其余。
func stripBrackets(s string, open, close rune) string {
	runes := []rune(s)
	var out strings.Builder
	i := 0
	for i < len(runes) {
		if runes[i] == open {
			depth := 1
			j := i + 1
			for j < len(runes) && depth > 0 {
				switch runes[j] {
				case open:
					depth++
				case close:
					depth--
				}
				if depth == 0 {
					break
				}
				j++
			}
			if j < len(runes) {
				out.WriteString(classifyBracket(string(runes[i+1:j]), open, close))
				i = j + 1
				continue
			}
		}
		out.WriteRune(runes[i])
		i++
	}
	return out.String()
}

// classifyBracket 把括号内容按 | _ / , 拆词判定，仅保留重度改编标注。
func classifyBracket(inner string, open, close rune) string {
	tokens := []string{inner}
	for _, sep := range []string{"|", "_", "/", ","} {
		var next []string
		for _, t := range tokens {
			next = append(next, strings.Split(t, sep)...)
		}
		tokens = next
	}
	var kept []string
	for _, t := range tokens {
		if k := matchKeep(t); k != "" {
			kept = append(kept, k)
		}
	}
	if len(kept) == 0 {
		return " "
	}
	return " " + string(open) + strings.Join(kept, " ") + string(close) + " "
}

// keepTokens 重度改编/效果标注（保留词）。匹配时取其子串，便于从混合括号中提取。
var keepTokens = []string{"remix", "rmx", "instrumental", "inst", "karaoke",
	"live", "nightcore", "sped up", "speed", "slowed", "reverbed", "bass boosted",
	"piano&guitar", "summernoise"}

// isKeepToken 判断是否为"重度改编/效果"标注（保留），否则视为信息性标注（去除）。
func isKeepToken(tok string) bool {
	return matchKeep(tok) != ""
}

// matchKeep 返回 tok 中包含的保留词短语；无则返回 ""。
// 例："(sic. instrumental)" 中的 token "sic. instrumental" -> "instrumental"。
func matchKeep(tok string) string {
	t := strings.ToLower(strings.TrimSpace(tok))
	if t == "" {
		return ""
	}
	for _, k := range keepTokens {
		if strings.Contains(t, k) {
			return k
		}
	}
	return ""
}

// artistInfo 归一化后的歌手信息。
type artistInfo struct {
	key      string   // 有序拼接的规范歌手集（占位通配时为 ""）
	tokens   []string // 规范歌手集（有序）
	wildcard bool     // 仅占位符/空 -> 通配
	canonical bool    // 无需别名映射即为规范写法
}

// buildArtistSet 归一化歌手字段：跨语言别名映射、占位符通配、统计是否规范。
func buildArtistSet(artist string, featArtists []string) artistInfo {
	set := map[string]bool{}
	remapped := false
	add := func(raw string) {
		raw = toHalfWidth(strings.ToLower(strings.TrimSpace(raw)))
		raw = collapseSpaces(raw)
		if raw == "" || isPlaceholderArtist(raw) {
			return
		}
		canon := artistAlias(raw)
		if canon != raw {
			remapped = true
		}
		set[canon] = true
	}
	for _, p := range splitArtistTokens(artist) {
		base, aliases := extractArtistAliases(p)
		if len(aliases) > 0 {
			// 有括号别名时以别名为准（如 에일리 (AILEE) -> ailee），便于与纯别名写法合并
			for _, a := range aliases {
				add(a)
			}
		} else {
			add(base)
		}
	}
	for _, f := range featArtists {
		for _, p := range splitArtistTokens(f) {
			add(p)
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return artistInfo{
		key:       strings.Join(keys, "/"),
		tokens:    keys,
		wildcard:  len(keys) == 0,
		canonical: !remapped,
	}
}

// artistAlias 查表把已知跨语言/错字艺人映射到规范写法。
func artistAlias(s string) string {
	if v, ok := artistAliasMap[s]; ok {
		return v
	}
	return s
}

// isPlaceholderArtist 是否为"未知/无约束"占位歌手。
func isPlaceholderArtist(s string) bool {
	return placeholderArtists[s]
}

// splitArtistTokens 按常见分隔符拆分歌手字段为多个艺人（去重专用，区别于 platform.go 的 splitArtists）。
func splitArtistTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '&' || r == '/' || r == '、' || r == '×' || r == ';'
	})
}

// extractArtistAliases 提取艺人括号内的别名，如 "에일리 (AILEE)" -> 基名+别名 AILEE。
func extractArtistAliases(token string) (base string, aliases []string) {
	base = artistAliasRegex.ReplaceAllStringFunc(token, func(m string) string {
		if sub := artistAliasRegex.FindStringSubmatch(m); len(sub) > 1 {
			aliases = append(aliases, sub[1])
		}
		return " "
	})
	base = collapseSpaces(base)
	return base, aliases
}

// toHalfWidth 把全角字符（FF01~FF5E）转半角，并统一全角空格为普通空格。
func toHalfWidth(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 0xFF01 && r <= 0xFF5E:
			b.WriteRune(r - 0xFEE0)
		case r == 0x3000:
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapseSpaces 合并多余空格并去首尾。
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// pinyinKey 将字符串转为拼音排序键：汉字转拼音、其余字符保留
func pinyinKey(s string) string {
	args := pinyin.NewArgs()
	return strings.Join(pinyin.LazyConvert(s, &args), "")
}

// applyFormat 按指定格式处理歌单（与 handler.formatSongList 逻辑一致，供聚合复用）
func applyFormat(songList *models.SongList, format string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}
	if format == "" || format == "song-singer" {
		return
	}
	formatted := make([]string, 0, len(songList.Songs))
	for _, song := range songList.Songs {
		switch format {
		case "singer-song":
			parts := strings.Split(song, " - ")
			if len(parts) == 2 {
				formatted = append(formatted, parts[1]+" - "+parts[0])
			} else {
				formatted = append(formatted, song)
			}
		case "song":
			parts := strings.Split(song, " - ")
			if len(parts) > 0 {
				formatted = append(formatted, parts[0])
			} else {
				formatted = append(formatted, song)
			}
		default:
			formatted = append(formatted, song)
		}
	}
	songList.Songs = formatted
}
