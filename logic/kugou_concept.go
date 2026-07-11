package logic

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"PlaylistAggregator/misc/httputil"
	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"
)

// kugouH5Salt 酷狗 H5 接口签名盐值（逆向自 @kg_interface-signature）
const kugouH5Salt = "NVPh5oo715z5DIWAeQlhMDsWXXQV4hwt"

// kugouConceptClient 酷狗概念版接口专用客户端（带超时）
var kugouConceptClient = &http.Client{Timeout: 15 * time.Second}

// kugouConceptResp get_list_info 响应（实测结构：data[0] 含歌单元数据 name/count，
// 歌曲列表在 get_other_list_file（GET，参数在 query，公开歌单无需登录即可直取））
type kugouConceptResp struct {
	Status    int    `json:"status"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Data      []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"data"`
}

// KugouConceptDiscover 解析酷狗概念版歌单。
// 链接形如 https://t1.kugou.com/1v5Bd17G3V3（短链，302 跳转分享页），
// 或直接分享页 https://activity.kugou.com/share/...?global_specialid=collection_...&specialid=...
// 歌单元数据接口为 POST https://pubsongs.kugou.com/v1/get_list_info，需经 H5 签名（已逆向并实测通过）。
// 歌曲列表接口为 GET https://pubsongscdn.kugou.com/v2/get_other_list_file，参数在 query，
// 经同一 H5 签名（空 token 即可公开直取，无需登录 cookie），按 page/pagesize 分页拉取。
func KugouConceptDiscover(link string, detailed bool) (*models.SongList, error) {
	specialID, globalSpecialID, err := resolveKugouConceptParams(link)
	if err != nil {
		return nil, err
	}
	if specialID == "" || globalSpecialID == "" {
		return nil, fmt.Errorf("无法从链接解析酷狗概念版歌单参数（specialid=%q, global_specialid=%q）", specialID, globalSpecialID)
	}

	// 1. 歌单元数据（名称 + 总数），H5 签名已校准通过
	meta, err := kugouConceptFetchMeta(specialID, globalSpecialID)
	if err != nil {
		return nil, err
	}

	// 2. 歌曲列表：GET get_other_list_file，空 token 公开直取，无需登录 cookie
	songs, err := kugouConceptFetchSongs(globalSpecialID)
	if err != nil {
		return nil, fmt.Errorf("酷狗概念版《%s》获取歌曲失败: %w", meta.Name, err)
	}

	// 歌曲列表接口返回的是 "歌名 - 歌手"（已被 reverseSongName 翻转），据此拆分出结构化信息
	detail := make([]models.SongItem, 0, len(songs))
	for _, s := range songs {
		parts := strings.SplitN(s, " - ", 2)
		name := s
		artistStr := ""
		if len(parts) == 2 {
			name = parts[0]
			artistStr = parts[1]
		}
		detail = append(detail, models.SongItem{
			Name:    name,
			Artists: splitArtists(artistStr),
			Album:   "",
			Id:      "",
		})
	}

	log.Infof("酷狗概念版歌单[%s]解析完成，共 %d 首", meta.Name, len(songs))
	return &models.SongList{
		Name:       meta.Name,
		Songs:      songs,
		SongsDetail: detail,
		SongsCount: len(songs),
	}, nil
}

// kugouConceptMeta 歌单元数据
type kugouConceptMeta struct {
	Name  string
	Count int
}

// kugouConceptFetchMeta 调 get_list_info 获取歌单名称与歌曲总数（H5 签名）
func kugouConceptFetchMeta(specialID, globalSpecialID string) (*kugouConceptMeta, error) {
	body := fmt.Sprintf(`{"data":[{"specialid":%s,"global_collection_id":"%s"}]}`,
		specialID, globalSpecialID)

	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	params := map[string]string{
		"clientver": "20000",
		"srcappid":  "2919",
		"dfid":      "-",
		"clienttime": ts,
		"mid":       ts,
		"uuid":      ts,
	}
	sig := kugouH5Sign(params, body)
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("signature", sig)
	apiURL := "https://pubsongs.kugou.com/v1/get_list_info?" + q.Encode()

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", mobileUA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://activity.kugou.com/")
	req.Header.Set("Origin", "https://activity.kugou.com")

	resp, err := kugouConceptClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求酷狗概念版接口失败: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取酷狗概念版响应失败: %w", err)
	}

	var out kugouConceptResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("解析酷狗概念版响应失败: %w (raw=%s)", err, truncate(string(raw), 300))
	}
	if out.ErrorCode != 0 || out.Status != 1 {
		return nil, fmt.Errorf("酷狗概念版接口返回错误 status=%d error_code=%d msg=%s (raw=%s)",
			out.Status, out.ErrorCode, out.ErrorMsg, truncate(string(raw), 300))
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("酷狗概念版接口未返回歌单数据 (raw=%s)", truncate(string(raw), 300))
	}
	return &kugouConceptMeta{Name: out.Data[0].Name, Count: out.Data[0].Count}, nil
}

// kugouConceptFetchSongs 调 get_other_list_file 获取歌曲列表。
// 该接口为 GET，参数全部置于 query（含 global_collection_id），token 留空即可公开直取，无需登录 cookie。
// 经与 get_list_info 相同的 H5 签名（空 body）。按 page/pagesize 分页拉取直到取满 data.count。
// 每首歌直接采用响应中的 name 字段（已是 "歌手 - 歌名" 格式）。
func kugouConceptFetchSongs(globalSpecialID string) ([]string, error) {
	const (
		srcappid  = "2919"
		clientver = "20000"
		appid     = "1058"
		pagesize  = "100"
	)
	var songs []string
	seen := make(map[string]bool)
	page := 1
	for {
		ts := fmt.Sprintf("%d", time.Now().UnixMilli())
		params := map[string]string{
			"srcappid":             srcappid,
			"clientver":            clientver,
			"clienttime":           ts,
			"uuid":                 ts,
			"dfid":                 "-",
			"uid":                  "0",
			"appid":                appid,
			"token":                "",
			"type":                 "0",
			"module":               "playlist",
			"page":                 fmt.Sprintf("%d", page),
			"pagesize":             pagesize,
			"global_collection_id": globalSpecialID,
		}
		sig := kugouH5Sign(params, "")
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		q.Set("signature", sig)
		apiURL := "https://pubsongscdn.kugou.com/v2/get_other_list_file?" + q.Encode()

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", mobileUA)
		req.Header.Set("Referer", "https://activity.kugou.com/")
		req.Header.Set("Origin", "https://activity.kugou.com")

		resp, err := kugouConceptClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求酷狗概念版歌曲接口失败: %w", err)
		}
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取酷狗概念版歌曲响应失败: %w", err)
		}

		var out struct {
			ErrorCode int    `json:"error_code"`
			ErrorMsg  string `json:"errmsg"`
			Data      struct {
				Count int `json:"count"`
				Info  []struct {
					Name string `json:"name"`
				} `json:"info"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("解析酷狗概念版歌曲响应失败: %w", err)
		}
		if out.ErrorCode != 0 {
			return nil, fmt.Errorf("酷狗概念版歌曲接口返回错误 error_code=%d msg=%s", out.ErrorCode, out.ErrorMsg)
		}
		if len(out.Data.Info) == 0 {
			break
		}
		for _, s := range out.Data.Info {
			raw := strings.TrimSpace(s.Name)
			if raw == "" || seen[raw] {
				continue
			}
			seen[raw] = true
			// 接口返回的 name 为"歌手 - 歌名"，翻转为"歌名 - 歌手"以与其他平台统一
			songs = append(songs, reverseSongName(raw))
		}
		if len(songs) >= out.Data.Count || len(out.Data.Info) < 100 {
			break
		}
		page++
	}
	if len(songs) == 0 {
		return nil, fmt.Errorf("酷狗概念版歌曲接口未返回歌曲")
	}
	return songs, nil
}

// resolveKugouConceptParams 从短链或分享页解析 specialid 与 global_specialid。
func resolveKugouConceptParams(link string) (specialID, globalSpecialID string, err error) {
	target := link
	// 短链 t1.kugou.com/xxx → 302 跳转到 activity.kugou.com/share/...?...
	if strings.Contains(link, "t1.kugou.com") {
		loc, e := httputil.GetRedirectLocation(link)
		if e != nil {
			return "", "", fmt.Errorf("跟随酷狗概念版短链失败: %w", e)
		}
		if loc == "" {
			return "", "", fmt.Errorf("酷狗概念版短链未返回跳转地址")
		}
		target = loc
	}

	u, e := url.Parse(target)
	if e != nil {
		return "", "", fmt.Errorf("解析酷狗概念版链接失败: %w", e)
	}
	q := u.Query()
	return q.Get("specialid"), q.Get("global_specialid"), nil
}

// kugouH5Sign 计算酷狗 H5 接口签名。
// 规则（逆向自 @kg_interface-signature，并经实测校准通过）：
//  1. 取除 signature 外的所有参数，按 key 字典序排序；
//  2. 拼接为 key1=value1&key2=value2... 形式（每对用 "=" 连接，无需 "&"，
//     JS 源码为 i.push(e+"="+m[e])）；
//  3. 若带 JSON body 一并拼在参数串之后；
//  4. 整体前后各加盐值 SALT；
//  5. 取 MD5(32 位小写)。
func kugouH5Sign(params map[string]string, body string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	raw := kugouH5Salt + sb.String() + body + kugouH5Salt

	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// truncate 截断字符串用于错误日志
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
