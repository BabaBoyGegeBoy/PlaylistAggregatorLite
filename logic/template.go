package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"PlaylistAggregator/misc/models"
)

const (
	// DefaultTopTemplate 歌单（顶层）模板默认值，对应示例：
	// {"name":"歌单名称","tracks":[ ...单曲对象... ]}
	DefaultTopTemplate = `{"name":"{{name}}","tracks":[{{songs}}]}`
	// DefaultSongTemplate 单曲模板默认值，对应示例：
	// {"name":"晴天","artist":["周杰伦"]}
	DefaultSongTemplate = `{"name":"{{song.name}}","artist":[{{song.artists}}]}`
)

// jsonContent 返回 JSON 转义后的字符串内容（不含两侧引号），便于直接替换到模板的 "..." 中。
// 例：jsonContent(`晴"天`) == `晴\"天`，模板写成 "name":"{{song.name}}" 即得到合法 JSON "name":"晴\"天"。
func jsonContent(s string) string {
	b, _ := json.Marshal(s)
	return strings.TrimSuffix(strings.TrimPrefix(string(b), `"`), `"`)
}

// renderArtists 将歌手数组渲染为 JSON 数组元素序列，如 "周杰伦","陈奕迅"。
// 用于放进模板的 [{{song.artists}}] 中，构成一个合法 JSON 数组。
func renderArtists(artists []string) string {
	parts := make([]string, 0, len(artists))
	for _, a := range artists {
		b, _ := json.Marshal(a)
		parts = append(parts, string(b))
	}
	return strings.Join(parts, ",")
}

// RenderCustomJSON 按双模板渲染自定义 JSON。
//
// 歌单模板 topTmpl 支持占位符：{{name}}（歌单名）、{{count}}（歌曲数）、{{songs}}（自动替换为单曲数组拼接）。
// 单曲模板 songTmpl 支持占位符：{{song.name}}（歌名）、{{song.artist}}（歌手字符串，多歌手用 "、" 连接）、
// {{song.artists}}（歌手数组元素，可直接放进 [ ... ]）、{{song.album}}（专辑）、{{song.id}}（歌曲 ID）。
//
// 渲染结果会做 json.Valid 校验，校验通过后美化（2 空格缩进）返回；不合法时返回原始串与错误。
func RenderCustomJSON(name string, songs []models.SongItem, topTmpl, songTmpl string) (string, error) {
	if topTmpl == "" {
		topTmpl = DefaultTopTemplate
	}
	if songTmpl == "" {
		songTmpl = DefaultSongTemplate
	}

	renderedSongs := make([]string, 0, len(songs))
	for _, s := range songs {
		rs := songTmpl
		rs = strings.ReplaceAll(rs, "{{song.name}}", jsonContent(s.Name))
		rs = strings.ReplaceAll(rs, "{{song.artists}}", renderArtists(s.Artists))
		rs = strings.ReplaceAll(rs, "{{song.artist}}", jsonContent(strings.Join(s.Artists, "、")))
		rs = strings.ReplaceAll(rs, "{{song.album}}", jsonContent(s.Album))
		rs = strings.ReplaceAll(rs, "{{song.id}}", jsonContent(s.Id))
		renderedSongs = append(renderedSongs, rs)
	}
	songsStr := strings.Join(renderedSongs, ",\n")

	out := topTmpl
	out = strings.ReplaceAll(out, "{{name}}", jsonContent(name))
	out = strings.ReplaceAll(out, "{{count}}", strconv.Itoa(len(songs)))
	out = strings.ReplaceAll(out, "{{songs}}", songsStr)

	if !json.Valid([]byte(out)) {
		return out, fmt.Errorf("自定义模板渲染结果不是合法 JSON，请检查模板语法（例如歌曲数组需用 [{{songs}}] 包裹）")
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(out), "", "  "); err != nil {
		return out, nil
	}
	return buf.String(), nil
}
