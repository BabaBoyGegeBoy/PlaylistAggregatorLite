package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"
	"PlaylistAggregator/misc/utils"
)

// mobileUA 手机端 UA，波点/酷狗概念版接口对 UA 与 Referer 有校验
const mobileUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"

// bodianClient 波点接口专用客户端（带超时）
var bodianClient = &http.Client{Timeout: 15 * time.Second}

// bodianReferer 波点接口必须携带的 Referer/Origin，否则返回 code:402
const bodianReferer = "https://h5app.kuwo.cn/"

// bodianInfoResp 歌单信息接口响应
type bodianInfoResp struct {
	Code int `json:"code"`
	Data struct {
		Name       string `json:"name"`
		MusicCount int    `json:"musicCount"`
	} `json:"data"`
}

// bodianSong 单首歌曲（musicList 列表项）
type bodianSong struct {
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	MusicRid string `json:"musicRid"`
	ID       int64  `json:"id"`
}

// bodianListResp 歌曲列表接口响应
type bodianListResp struct {
	Code int `json:"code"`
	Data struct {
		Total int          `json:"total"`
		List  []bodianSong `json:"list"`
	} `json:"data"`
}

// BodianDiscover 解析波点音乐歌单（链接形如
// https://h5app.kuwo.cn/m/bodian/collection.html?uid=xxx&playlistId=91813119&source=5）。
// 波点是 QQ音乐简洁版，但歌单链接在 kuwo 域名且走独立 bd-api.kuwo.cn 接口，
// 不能复用 music-lib 的 kuwo 解析器。
func BodianDiscover(link string, detailed bool) (*models.SongList, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("解析波点音乐链接失败: %w", err)
	}
	q := u.Query()

	playlistID := q.Get("playlistId")
	if playlistID == "" {
		return nil, errors.New("波点音乐链接缺少 playlistId 参数")
	}
	if _, err := strconv.Atoi(playlistID); err != nil {
		return nil, fmt.Errorf("波点音乐 playlistId 非法: %s", playlistID)
	}

	// source 实测不可省（波点为 QQ音乐源），缺失默认 5 并记录告警
	source := q.Get("source")
	if source == "" {
		source = "5"
		log.Warnf("波点音乐链接缺少 source 参数，默认按 source=5 请求: %s", link)
	}

	// 1. 歌单信息（歌单名 + 总数）
	infoURL := fmt.Sprintf("https://bd-api.kuwo.cn/api/service/playlist/info/%s?source=%s", playlistID, source)
	infoBody, err := bodianHTTPGet(infoURL)
	if err != nil {
		return nil, fmt.Errorf("获取波点音乐歌单信息失败: %w", err)
	}
	var info bodianInfoResp
	if err := json.Unmarshal(infoBody, &info); err != nil {
		return nil, fmt.Errorf("解析波点音乐歌单信息失败: %w", err)
	}
	if info.Code != 200 {
		return nil, fmt.Errorf("波点音乐歌单信息返回错误 code=%d", info.Code)
	}

	total := info.Data.MusicCount
	if total <= 0 {
		return &models.SongList{
			Name:       info.Data.Name,
			Songs:      []string{},
			SongsDetail: []models.SongItem{},
			SongsCount: 0,
		}, nil
	}

	// 2. 歌曲列表（分页循环，pn 为 1-based，每页 rn=100）。
	//    带 plat:h5 头后单页可取满 ≤100 的歌单；>100 首按 total 累加分页取满，空页即终止。
	const rn = 100
	songs := make([]string, 0, total)
	detail := make([]models.SongItem, 0, total)
	collected := 0
	for pn := 1; collected < total && pn < 200; pn++ {
		listURL := fmt.Sprintf("https://bd-api.kuwo.cn/api/service/playlist/%s/musicList?source=%s&pn=%d&rn=%d",
			playlistID, source, pn, rn)
		listBody, err := bodianHTTPGet(listURL)
		if err != nil {
			return nil, fmt.Errorf("获取波点音乐歌曲列表失败: %w", err)
		}
		var list bodianListResp
		if err := json.Unmarshal(listBody, &list); err != nil {
			return nil, fmt.Errorf("解析波点音乐歌曲列表失败: %w", err)
		}
		if list.Code != 200 {
			return nil, fmt.Errorf("波点音乐歌曲列表返回错误 code=%d", list.Code)
		}
		if len(list.Data.List) == 0 {
			break
		}
		for _, s := range list.Data.List {
			name := s.Name
			if !detailed {
				name = utils.StandardSongName(s.Name)
			}
			songs = append(songs, fmt.Sprintf("%s - %s", name, s.Artist))
			detail = append(detail, models.SongItem{
				Name:    name,
				Artists: splitArtists(s.Artist),
				Album:   "",
				Id:      strconv.FormatInt(s.ID, 10),
			})
		}
		collected += len(list.Data.List)
	}

	log.Infof("波点音乐歌单[%s]解析完成，共 %d 首", info.Data.Name, len(songs))
	return &models.SongList{
		Name:       info.Data.Name,
		Songs:      songs,
		SongsDetail: detail,
		SongsCount: len(songs),
	}, nil
}

// bodianHTTPGet 带 Referer/Origin/UA 的 GET 请求，返回响应体字节
func bodianHTTPGet(apiURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", mobileUA)
	req.Header.Set("Referer", bodianReferer)
	req.Header.Set("Origin", bodianReferer)
	req.Header.Set("Accept", "application/json")
	// plat: h5 是解除匿名请求歌曲数量上限的关键头（缺失时接口只返回约 20/40 首）；
	// ver 为空值一并携带以对齐网页真实请求。
	req.Header.Set("plat", "h5")
	req.Header.Set("ver", "")

	resp, err := bodianClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
