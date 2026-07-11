package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"PlaylistAggregator/misc/httputil"
	"PlaylistAggregator/misc/models"
	"PlaylistAggregator/misc/utils"
)

// neteasySongEntry 解析分块时缓存的单曲结构化信息（与 "歌名 - 歌手" 字符串并存）
type neteasySongEntry struct {
	Info string
	Item models.SongItem
}

const (
	netEasyUrlV6 = "https://music.163.com/api/v6/playlist/detail"
	netEasyUrlV3 = "https://music.163.com/api/v3/song/detail"
	chunkSize    = 400
)

// NetEasyDiscover 获取网易云音乐歌单信息（不依赖 Redis/MySQL，始终直连网易云 API）
// link: 歌单链接
// detailed: 是否使用详细歌曲名（原始歌曲名，不去除括号等内容）
func NetEasyDiscover(link string, detailed bool) (*models.SongList, error) {
	// 1. 获取歌单基本信息
	songIdsResp, err := getSongsInfo(link)
	if err != nil {
		return nil, fmt.Errorf("获取歌单信息失败: %w", err)
	}

	playlistName := songIdsResp.Playlist.Name     // 歌单名
	trackIds := songIdsResp.Playlist.TrackIds      // 歌曲ID列表
	tracksCount := songIdsResp.Playlist.TrackCount // 歌曲总数

	// 如果歌单为空，直接返回
	if len(trackIds) == 0 {
		return &models.SongList{
			Name:       playlistName,
			Songs:      []string{},
			SongsDetail: []models.SongItem{},
			SongsCount: 0,
		}, nil
	}

	// 收集所有歌曲ID，直接从 API 获取歌曲信息（无需缓存/数据库）
	allSongIds := make([]uint, len(trackIds))
	for i, track := range trackIds {
		allSongIds[i] = track.Id
	}

	resultMap := sync.Map{}
	if _, err := batchGetSongs(allSongIds, &resultMap, detailed); err != nil {
		return nil, fmt.Errorf("获取歌曲详情失败: %w", err)
	}

	// 返回最终结果
	return createSongList(playlistName, trackIds, resultMap, tracksCount), nil
}

// createSongList 创建歌单结果（同时填充 Songs 与结构化 SongsDetail）
func createSongList(name string, trackIds []*models.TrackId, resultMap sync.Map, count int) *models.SongList {
	songs := make([]string, 0, len(trackIds))
	detail := make([]models.SongItem, 0, len(trackIds))
	for _, t := range trackIds {
		v, ok := resultMap.Load(t.Id)
		if !ok {
			continue
		}
		e := v.(neteasySongEntry)
		songs = append(songs, e.Info)
		detail = append(detail, e.Item)
	}
	return &models.SongList{
		Name:       name,
		Songs:      songs,
		SongsDetail: detail,
		SongsCount: count,
	}
}

// getSongsInfo 获取歌单基本信息
func getSongsInfo(link string) (*models.NetEasySongId, error) {
	songListId, err := utils.GetNetEasyParam(link)
	if err != nil {
		return nil, fmt.Errorf("解析歌单链接失败: %w", err)
	}

	resp, err := httputil.Post(netEasyUrlV6, strings.NewReader("id="+songListId))
	if err != nil {
		return nil, fmt.Errorf("请求网易云API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %w", err)
	}

	songIdsResp := &models.NetEasySongId{}
	if err = json.Unmarshal(body, songIdsResp); err != nil {
		return nil, fmt.Errorf("解析响应内容失败: %w", err)
	}

	if songIdsResp.Code == 401 {
		return nil, errors.New("无权限访问该歌单")
	}

	return songIdsResp, nil
}

// batchGetSongs 批量获取歌曲详情，结果写入 resultMap
func batchGetSongs(missKeys []uint, resultMap *sync.Map, detailed bool) (sync.Map, error) {
	if len(missKeys) == 0 {
		return sync.Map{}, nil
	}

	// 1. 构建请求参数
	missSongIds := make([]*models.SongId, len(missKeys))
	for i, id := range missKeys {
		missSongIds[i] = &models.SongId{Id: id}
	}

	// 2. 分块处理，避免请求过大
	missSize := len(missSongIds)
	chunkCount := (missSize + chunkSize - 1) / chunkSize
	chunks := make([][]*models.SongId, chunkCount)

	for i := 0; i < missSize; i += chunkSize {
		end := i + chunkSize
		if end > missSize {
			end = missSize
		}
		chunks[i/chunkSize] = missSongIds[i:end]
	}

	// 3. 并发请求处理
	var eg errgroup.Group

	for _, chunk := range chunks {
		chunk := chunk // 创建副本避免闭包问题
		eg.Go(func() error {
			return processChunk(chunk, resultMap, detailed)
		})
	}

	// 4. 等待所有请求完成
	if err := eg.Wait(); err != nil {
		return sync.Map{}, err
	}

	return sync.Map{}, nil
}

// processChunk 处理一个分块的歌曲ID
func processChunk(chunk []*models.SongId, resultMap *sync.Map, detailed bool) error {
	// 1. 序列化请求参数
	marshal, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("序列化请求参数失败: %w", err)
	}

	// 2. 发送请求
	resp, err := httputil.Post(netEasyUrlV3, strings.NewReader("c="+string(marshal)))
	if err != nil {
		return fmt.Errorf("请求歌曲详情失败: %w", err)
	}
	defer resp.Body.Close()

	// 3. 读取响应内容
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 4. 解析响应内容
	songs := &models.Songs{}
	if err = json.Unmarshal(bytes, songs); err != nil {
		return fmt.Errorf("解析响应内容失败: %w", err)
	}

	// 5. 处理歌曲信息
	for _, song := range songs.Songs {
		// 根据detailed参数决定是否使用原始歌曲名
		var songName string
		if detailed {
			songName = song.Name // 使用原始歌曲名
		} else {
			songName = utils.StandardSongName(song.Name) // 使用标准化的歌曲名
		}

		// 构建作者信息
		authors := make([]string, len(song.Ar))
		for i, ar := range song.Ar {
			authors[i] = ar.Name
		}

		// 格式化歌曲信息（"歌名 - 歌手"）
		songInfo := fmt.Sprintf("%s - %s", songName, strings.Join(authors, " / "))

		// 同时缓存结构化信息，供自定义 JSON 模板使用
		entry := neteasySongEntry{
			Info: songInfo,
			Item: models.SongItem{
				Name:    songName,
				Artists: authors,
				Album:   "",
				Id:      strconv.FormatUint(uint64(song.Id), 10),
			},
		}

		// 存储结果
		resultMap.Store(song.Id, entry)
	}

	return nil
}
