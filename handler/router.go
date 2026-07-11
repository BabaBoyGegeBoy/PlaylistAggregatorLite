package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(distFS embed.FS) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())                  // allow all origins
	router.POST("/songlist", MusicHandler)      // 单歌单解析
	router.POST("/aggregate", AggregateHandler) // 多歌单聚合

	// 托管前端 build 产物（已通过 //go:embed 编译进二进制，无需外部 static/dist 目录）
	sub, err := fs.Sub(distFS, "static/dist")
	if err != nil {
		panic(err)
	}
	router.StaticFS("/", http.FS(sub))
	return router
}
