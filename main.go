package main

import (
	"embed"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"PlaylistAggregator/handler"
	"PlaylistAggregator/misc/log"
	"PlaylistAggregator/misc/models"
)

//go:embed static/dist
var distFS embed.FS

// openBrowser 尝试用系统默认浏览器打开指定 URL。
// 本项目主要发布 Windows 二进制，Windows 下使用 `cmd /c start`；
// 同时为 macOS / Linux 预留对应命令，便于跨平台编译。
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// start 的第二个参数是窗口标题，需用空串占位；URL 作为第三个参数
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func main() {
	r := handler.NewRouter(distFS)

	// 启动 HTTP 服务后，自动调起默认浏览器打开网页
	go func() {
		// 稍等片刻，待端口真正开始监听再打开，避免白屏
		time.Sleep(600 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://127.0.0.1:%d/", models.Port))
	}()

	if err := r.Run(fmt.Sprintf(models.PortFormat, models.Port)); err != nil {
		log.Errorf("fail to run server: %v", err)
		panic(err)
	}
}
