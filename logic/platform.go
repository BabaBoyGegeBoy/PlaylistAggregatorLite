package logic

import (
	"errors"
	"fmt"
	"strings"

	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"
	"PlaylistAggregator/misc/utils"

	mlmodel "github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/apple"
	"github.com/guohuiyuan/music-lib/bilibili"
	"github.com/guohuiyuan/music-lib/fivesing"
	"github.com/guohuiyuan/music-lib/jamendo"
	"github.com/guohuiyuan/music-lib/joox"
	"github.com/guohuiyuan/music-lib/kuwo"
	"github.com/guohuiyuan/music-lib/migu"
	"github.com/guohuiyuan/music-lib/qianqian"
	"github.com/guohuiyuan/music-lib/soda"
)

// musicLibParseFunc 是 music-lib 各平台歌单解析函数的统一签名
type musicLibParseFunc func(link string) (*mlmodel.Playlist, []mlmodel.Song, error)

// newMusicLibParsers 构建"平台 -> 解析函数"映射（仅公开歌单，cookie 为空）
// 以下平台在 PlaylistAggregator 中此前缺失，使用 go-music-dl 所依赖的 music-lib 补齐。
var newMusicLibParsers = map[string]musicLibParseFunc{
	"kuwo":     kuwo.New("").ParsePlaylist,
	"migu":     migu.New("").ParsePlaylist,
	"qianqian": qianqian.New("").ParsePlaylist,
	"joox":     joox.New("").ParsePlaylist,
	"bilibili": bilibili.New("").ParsePlaylist,
	"fivesing": fivesing.New("").ParsePlaylist,
	"apple":    apple.New("").ParsePlaylist,
	"jamendo":  jamendo.New("").ParsePlaylist,
	// 汽水音乐：music-lib/soda 通过抓取分享页重定向 + 汽水 PC 端 API 解析，
	// 空 cookie 即可解析公开歌单（PlaylistAggregator 原手写 goquery 解析对 JS-SPA 失效）。
	"qishui": soda.New("").ParsePlaylist,
}

// DetectPlatform 根据歌单链接识别平台。
// 返回与 PlaylistAggregator 内部一致的平台标识：netease / qq / qishui / kugou / kuwo /
// migu / qianqian / joox / bilibili / fivesing / apple / jamendo。
// 移植自 go-music-dl 的 DetectSource，并保留 PlaylistAggregator 原有的 163cn 短链识别。
func DetectPlatform(link string) string {
	if strings.Contains(link, "163.com") || strings.Contains(link, "163cn") {
		return "netease"
	}
	if strings.Contains(link, "qq.com") {
		return "qq"
	}
	if strings.Contains(link, "5sing") {
		return "fivesing"
	}
	// 酷狗概念版：短链 t1.kugou.com 或 activity.kugou.com/share 分享页（含 collection 全局收藏ID）
	// 必须在普通 kugou.com 判断之前，否则会被识别成 kugou。
	if strings.Contains(link, "t1.kugou.com") ||
		(strings.Contains(link, "activity.kugou.com/share") && strings.Contains(link, "collection")) {
		return "kugouconcept"
	}
	if strings.Contains(link, "kugou.com") {
		return "kugou"
	}
	// 波点音乐：歌单链接在 kuwo 域名但走独立 bd-api.kuwo.cn 接口，不能复用 kuwo 解析器。
	// 必须在普通 kuwo.cn 判断之前。
	if strings.Contains(link, "h5app.kuwo.cn/m/bodian/") {
		return "bodian"
	}
	if strings.Contains(link, "kuwo.cn") {
		return "kuwo"
	}
	if strings.Contains(link, "migu.cn") {
		return "migu"
	}
	if strings.Contains(link, "joox.com") {
		return "joox"
	}
	if strings.Contains(link, "douyin.com") || strings.Contains(link, "qishui") {
		return "qishui"
	}
	if strings.Contains(link, "91q.com") {
		return "qianqian"
	}
	if strings.Contains(link, "jamendo.com") {
		return "jamendo"
	}
	if strings.Contains(link, "music.apple.com") || strings.Contains(link, "itunes.apple.com") {
		return "apple"
	}
	if strings.Contains(link, "bilibili.com") || strings.Contains(link, "b23.tv") {
		return "bilibili"
	}
	return ""
}

// platformNameCN 平台标识 -> 中文名映射（用于返回给前端的 platform_name）
var platformNameCN = map[string]string{
	"netease":       "网易云音乐",
	"qq":           "QQ音乐",
	"qishui":       "汽水音乐",
	"kugou":        "酷狗音乐",
	"kugouconcept": "酷狗概念版",
	"kuwo":         "酷我音乐",
	"bodian":       "波点音乐",
	"migu":     "咪咕音乐",
	"qianqian": "千千音乐",
	"joox":     "JOOX",
	"bilibili": "哔哩哔哩",
	"fivesing": "5Sing原创音乐",
	"apple":    "Apple Music",
	"jamendo":  "Jamendo",
}

// platformName 返回平台中文名，未知时回退为原标识
func platformName(p string) string {
	if v, ok := platformNameCN[p]; ok {
		return v
	}
	return p
}

// Discover 统一的歌单解析入口：自动识别平台并返回 models.SongList。
// - netease / qq 沿用 PlaylistAggregator 原生解析（含缓存/签名/分页）
// - qishui(汽水) 及其他平台通过 music-lib 补齐
// 返回的 SongList 会附带 Platform(平台标识) 与 PlatformName(中文名)，供前端展示来源标签。
func Discover(link string, detailed bool) (*models.SongList, error) {
	if strings.TrimSpace(link) == "" {
		return nil, errors.New("歌单链接不能为空")
	}

	platform := DetectPlatform(link)
	var (
		sl  *models.SongList
		err error
	)

	switch platform {
	case "netease":
		sl, err = NetEasyDiscover(link, detailed)
	case "qq":
		sl, err = QQMusicDiscover(link, detailed)
	case "kugou":
		sl, err = KugouDiscover(link, detailed)
	case "bodian":
		sl, err = BodianDiscover(link, detailed)
	case "kugouconcept":
		sl, err = KugouConceptDiscover(link, detailed)
	default:
		parseFn, ok := newMusicLibParsers[platform]
		if !ok {
			log.Warnf("不支持的音乐链接格式: %s", link)
			return nil, errors.New("不支持的音乐链接格式")
		}
		var playlist *mlmodel.Playlist
		var songs []mlmodel.Song
		playlist, songs, err = parseFn(link)
		if err != nil {
			log.Errorf("解析%s歌单失败: %v", platform, err)
			return nil, fmt.Errorf("解析歌单失败: %w", err)
		}
		sl = convertMusicLibPlaylist(playlist, songs, detailed)
	}

	if err != nil {
		log.Errorf("获取歌单失败: %v", err)
		return nil, err
	}

	if sl != nil {
		sl.Platform = platform
		sl.PlatformName = platformName(platform)
	}
	return sl, nil
}

// splitArtists 将合并的歌手字符串拆分为歌手数组。
// music-lib 的 Song.Artist 多为已合并字符串，按 "、" / "/" / "|" 拆分兜底；
// 若某平台只返回单一合并串，则退化为单元素数组。
func splitArtists(artist string) []string {
	artist = strings.TrimSpace(artist)
	if artist == "" {
		return []string{}
	}
	parts := strings.FieldsFunc(artist, func(r rune) bool {
		return r == '、' || r == '/' || r == '|'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{artist}
	}
	return out
}

// convertMusicLibPlaylist 将 music-lib 的 Playlist/Song 转换为 PlaylistAggregator 统一的 SongList。
// 歌曲名沿用 PlaylistAggregator 的 StandardSongName 做规范化（非 detailed 模式）。
// 同时填充结构化 SongsDetail（歌手按 splitArtists 拆分；Id/Album 取 music-lib 原值）。
func convertMusicLibPlaylist(playlist *mlmodel.Playlist, songs []mlmodel.Song, detailed bool) *models.SongList {
	name := ""
	if playlist != nil {
		name = playlist.Name
	}

	out := make([]string, 0, len(songs))
	detail := make([]models.SongItem, 0, len(songs))
	for _, s := range songs {
		songName := s.Name
		if !detailed {
			songName = utils.StandardSongName(s.Name)
		}
		out = append(out, fmt.Sprintf("%s - %s", songName, s.Artist))
		detail = append(detail, models.SongItem{
			Name:    songName,
			Artists: splitArtists(s.Artist),
			Album:   s.Album,
			Id:      s.ID,
		})
	}

	return &models.SongList{
		Name:       name,
		Songs:      out,
		SongsDetail: detail,
		SongsCount: len(out),
	}
}

// reverseSongName 将"歌手 - 歌名"翻转为"歌名 - 歌手"。
// 仅按首个" - "切分一次，即使歌名内含" - "也能完整保留（如"七里香 - live"）。
func reverseSongName(s string) string {
	parts := strings.SplitN(s, " - ", 2)
	if len(parts) == 2 {
		return parts[1] + " - " + parts[0]
	}
	return s
}
