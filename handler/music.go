package handler

import (
	"net/http"
	"strings"
	"sync/atomic"

	"PlaylistAggregator/logic"
	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"

	"github.com/gin-gonic/gin"
)

const (
	SUCCESS = "success"
)

var (
	counter atomic.Int64 // request counter
)

// MusicHandler 处理音乐请求的入口函数
func MusicHandler(c *gin.Context) {
	link := c.PostForm("url")
	detailed := c.Query("detailed") == "true"
	format := c.Query("format")
	order := c.Query("order")
	currentCount := counter.Add(1)

	log.Infof("第 %v 次歌单请求：%v，详细歌曲名：%v，歌曲格式：%v，歌曲顺序：%v", currentCount, link, detailed, format, order)

	// 统一入口：自动识别平台并解析歌单（netease/qq/qishui 走原手写解析，其余走 music-lib）
	songList, err := logic.Discover(link, detailed)
	if err != nil {
		log.Errorf("获取歌单失败: %v", err)
		c.JSON(http.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: err.Error(), Data: nil})
		return
	}

	// 自定义 JSON 模板模式：按双模板渲染结构化歌曲信息并返回 JSON 字符串
	if format == "custom" {
		tmplTop := c.PostForm("template_top")
		tmplSong := c.PostForm("template_song")
		// 倒序处理（与 Songs 字段保持一致）
		for i, j := 0, len(songList.SongsDetail)-1; i < j; i, j = i+1, j-1 {
			songList.SongsDetail[i], songList.SongsDetail[j] = songList.SongsDetail[j], songList.SongsDetail[i]
		}
		out, renderErr := logic.RenderCustomJSON(songList.Name, songList.SongsDetail, tmplTop, tmplSong)
		if renderErr != nil {
			c.JSON(http.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: renderErr.Error(), Data: out})
			return
		}
		c.JSON(http.StatusOK, &models.Result{Code: models.SuccessCode, Msg: SUCCESS, Data: gin.H{
			"json":  out,
			"name":  songList.Name,
			"count": len(songList.SongsDetail),
		}})
		return
	}

	// 根据格式选项处理歌曲列表
	formatSongList(songList, format)

	// 根据顺序选项处理歌曲列表
	processSongOrder(songList, order)

	c.JSON(http.StatusOK, &models.Result{Code: models.SuccessCode, Msg: SUCCESS, Data: songList})
}

// processSongOrder 根据指定的顺序处理歌曲列表
func processSongOrder(songList *models.SongList, order string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	// 如果是倒序，则反转歌曲列表
	if order == "reverse" {
		songs := songList.Songs
		for i, j := 0, len(songs)-1; i < j; i, j = i+1, j-1 {
			songs[i], songs[j] = songs[j], songs[i]
		}
	}
}

// formatSongList 根据指定的格式处理歌曲列表
func formatSongList(songList *models.SongList, format string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	// 如果没有指定格式或格式为默认的"歌名-歌手"，则不做处理
	if format == "" || format == "song-singer" {
		return
	}

	formattedSongs := make([]string, 0, len(songList.Songs))

	for _, song := range songList.Songs {
		switch format {
		case "singer-song":
			// 将"歌名 - 歌手"转换为"歌手 - 歌名"
			parts := strings.Split(song, " - ")
			if len(parts) == 2 {
				formattedSongs = append(formattedSongs, parts[1]+" - "+parts[0])
			} else {
				// 如果格式不符合预期，保持原样
				formattedSongs = append(formattedSongs, song)
			}
		case "song":
			// 只保留歌名
			parts := strings.Split(song, " - ")
			if len(parts) > 0 {
				formattedSongs = append(formattedSongs, parts[0])
			} else {
				formattedSongs = append(formattedSongs, song)
			}
		default:
			// 未知格式，保持原样
			formattedSongs = append(formattedSongs, song)
		}
	}

	// 更新歌曲列表
	songList.Songs = formattedSongs
}
