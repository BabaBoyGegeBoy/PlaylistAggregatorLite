package models

// AggregateSource 聚合时单个源歌单的解析明细
type AggregateSource struct {
	URL          string `json:"url"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platform_name"`
	Count        int    `json:"count"`
	OK           bool   `json:"ok"`
	Error        string `json:"error,omitempty"`
}

// AggregateResult 多歌单聚合结果：跨平台解析、去重、排序后合并为一个歌单
type AggregateResult struct {
	Name              string            `json:"name"`
	Songs             []string          `json:"songs"`
	SongsDetail       []SongItem        `json:"songs_detail,omitempty"` // 结构化歌曲信息，供自定义 JSON 模板使用
	SongsCount        int               `json:"songs_count"`
	Sources           []AggregateSource `json:"sources"`
	DuplicatesRemoved int               `json:"duplicates_removed"`
}
