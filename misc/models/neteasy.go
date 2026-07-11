package models

import "fmt"

// SongItem 单首歌的结构化信息，供「自定义 JSON 模板」渲染使用。
// 与原 Songs []string（"歌名 - 歌手"）并存，互不影响。
type SongItem struct {
	Name    string   `json:"name"`             // 歌名
	Artists []string `json:"artists"`          // 歌手数组，如 ["周杰伦","杨瑞代"]
	Album   string   `json:"album,omitempty"`  // 专辑（部分平台可能为空）
	Id      string   `json:"id,omitempty"`     // 歌曲 ID（部分平台可能为空）
}

// SongList represents a song list entity.
type SongList struct {
	Name         string     `json:"name"`              // song list name
	Songs        []string   `json:"songs"`             // songs list ("歌名 - 歌手")
	SongsDetail  []SongItem `json:"songs_detail,omitempty"` // 结构化歌曲信息，供自定义 JSON 模板使用
	SongsCount   int        `json:"songs_count"`       // total number of songs
	Platform     string     `json:"platform"`          // 来源平台标识，如 qq / kugou / netease
	PlatformName string     `json:"platform_name"`     // 来源平台中文名，如 QQ音乐 / 酷狗音乐
}

// SongId represents a song ID entity.
type SongId struct {
	Id uint `json:"id"`
}

// String returns the string representation of the SongId.
func (r *SongId) String() string {
	if r == nil {
		return "nil"
	}
	return fmt.Sprintf("{\"id\":%v}", r.Id)
}

// NetEasySongId represents a NetEasy song ID entity.
type NetEasySongId struct {
	Code     int `json:"code"`
	Playlist struct {
		Id         int64      `json:"id"`
		Name       string     `json:"name"`
		TrackIds   []*TrackId `json:"trackIds"`
		TrackCount int        `json:"trackCount"`
	} `json:"playlist"`
}

// TrackId represents a track ID entity.
type TrackId struct {
	Id uint `json:"id"`
}

// Songs represents a songs entity.
type Songs struct {
	Songs []struct {
		Id   uint   `json:"id"`
		Name string `json:"name"`
		Ar   []struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"ar"`
	} `json:"songs"`
}
