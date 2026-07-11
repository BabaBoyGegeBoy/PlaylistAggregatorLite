package logic

import (
	"os"
	"regexp"
	"strings"

	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"

	mlmodel "github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/kugou"
)

// reKugouGcid 匹配酷狗分享/收藏歌单的全局收藏 ID（形如 gcid_3z19of348z5z0aa）。
// 这类链接在 music-lib 中走 fetchSonglistDetail，仅返回首屏 3 首，
// 需改用 cloudlist 接口按 gcid 拉全量。
var reKugouGcid = regexp.MustCompile(`gcid_[A-Za-z0-9_]+`)

func isKugouGcidLink(link string) bool {
	return strings.Contains(link, "kugou.com") && reKugouGcid.MatchString(link)
}

func extractKugouGcid(link string) string {
	return reKugouGcid.FindString(link)
}

// loadKugouCookie 读取酷狗 App 端 cookie（须含 userid/token/KUGOU_API_MID）。
// 优先环境变量 KUGOU_COOKIE，其次项目根目录的 .kugou_cookie 文件。
// 不硬编码、不提交 git。
func loadKugouCookie() string {
	if v := strings.TrimSpace(os.Getenv("KUGOU_COOKIE")); v != "" {
		return v
	}
	for _, p := range []string{".kugou_cookie", "kugou_cookie.txt"} {
		if b, err := os.ReadFile(p); err == nil {
			if s := strings.TrimSpace(string(b)); s != "" {
				return s
			}
		}
	}
	return ""
}

// KugouDiscover 酷狗歌单解析入口。
// - 非 gcid_ 链接（普通 special 歌单）走 music-lib 原逻辑；
// - gcid_ 分享/收藏歌单：用 App 端 cookie 调 cloudlist 接口（get_list_all_file）
//   补全全量歌曲，绕过 fetchSonglistDetail 仅返回首屏 3 首的限制；
//   cookie 缺失或 cloudlist 调用失败时自动回退原逻辑（仅首屏歌曲）。
func KugouDiscover(link string, detailed bool) (*models.SongList, error) {
	if !isKugouGcidLink(link) {
		playlist, songs, err := kugou.New("").ParsePlaylist(link)
		if err != nil {
			return nil, err
		}
		return convertMusicLibPlaylist(playlist, songs, detailed), nil
	}

	gcid := extractKugouGcid(link)

	// 元信息（歌单名/封面/count）用默认解析即可，结果准确
	metaPlaylist, metaSongs, metaErr := kugou.New("").ParsePlaylist(link)
	if metaErr != nil || metaPlaylist == nil {
		metaPlaylist = &mlmodel.Playlist{}
	}

	cookie := loadKugouCookie()
	if strings.TrimSpace(cookie) == "" {
		log.Warnf("酷狗 gcid 歌单[%s]补全需要 App 端 cookie，回退默认解析（仅首屏歌曲）", gcid)
		return convertMusicLibPlaylist(metaPlaylist, metaSongs, detailed), nil
	}

	songs, err := kugou.New(cookie).GetPlaylistSongs("cloudlist:" + gcid)
	if err != nil {
		log.Warnf("酷狗 gcid 歌单[%s]cloudlist 补全失败: %v，回退默认解析", gcid, err)
		return convertMusicLibPlaylist(metaPlaylist, metaSongs, detailed), nil
	}

	log.Infof("酷狗 gcid 歌单[%s]cloudlist 补全成功，共 %d 首", gcid, len(songs))
	return convertMusicLibPlaylist(metaPlaylist, songs, detailed), nil
}
