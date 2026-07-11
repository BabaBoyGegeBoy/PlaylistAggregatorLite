package handler

import (
	"net/http"
	"strings"

	"PlaylistAggregator/logic"
	"PlaylistAggregator/misc/models"

	"github.com/gin-gonic/gin"
)

// AggregateHandler 处理多歌单聚合请求：入参 urls（多行，每行一个链接）
func AggregateHandler(c *gin.Context) {
	raw := c.PostForm("urls")
	detailed := c.Query("detailed") == "true"
	format := c.Query("format")
	order := c.Query("order")

	if strings.TrimSpace(raw) == "" {
		c.JSON(http.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: "请提供至少一个歌单链接", Data: nil})
		return
	}
	urls := strings.Split(raw, "\n")

	agg, err := logic.Aggregate(urls, detailed, format, order)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: err.Error(), Data: nil})
		return
	}

	// 自定义 JSON 模板模式：按双模板渲染结构化歌曲信息并返回 JSON 字符串
	if format == "custom" {
		tmplTop := c.PostForm("template_top")
		tmplSong := c.PostForm("template_song")
		out, renderErr := logic.RenderCustomJSON(agg.Name, agg.SongsDetail, tmplTop, tmplSong)
		if renderErr != nil {
			c.JSON(http.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: renderErr.Error(), Data: out})
			return
		}
		c.JSON(http.StatusOK, &models.Result{Code: models.SuccessCode, Msg: SUCCESS, Data: gin.H{
			"json":  out,
			"name":  agg.Name,
			"count": len(agg.SongsDetail),
		}})
		return
	}

	c.JSON(http.StatusOK, &models.Result{Code: models.SuccessCode, Msg: SUCCESS, Data: agg})
}
