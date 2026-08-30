// Package server 组装 Echo 应用并启动 HTTP 服务。
package server

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"nav/cmd"
	"nav/config/data"
	"nav/config/define"
	"nav/config/model"
	"nav/internal/pages/home"
	"nav/internal/resources"
)

// NewRouter 构建路由并返回 http.Handler，便于在测试里直接用 httptest 驱动。
func NewRouter() (http.Handler, error) {
	e := echo.New()
	e.Use(middleware.Recover())

	if err := resources.Register(e); err != nil {
		return nil, err
	}
	resources.RegisterAssets(e)
	home.RegisterRouting(e)
	return e, nil
}

// StartDaemon 注入启动参数并阻塞运行，与 flare 的入口保持一致。
func StartDaemon(flags *model.Flags) {
	if flags != nil {
		define.AppFlags = *flags
	}

	handler, err := NewRouter()
	if err != nil {
		log.Fatalf("启动失败：%v", err)
	}

	addr := cmd.Address(define.AppFlags.Port)

	// 启动时预加载数据（如有异常会自动回退至内置示例，不阻断服务启动）
	if _, err := data.Load(); err != nil {
		log.Printf("[警告] 数据文件预加载提示：%v", err)
	}

	log.Printf("书签数据：%s", define.ConfigPath())
	log.Printf("nav 已启动 → http://localhost%s", addr)

	server := &http.Server{Addr: addr, Handler: handler}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("监听 %s 失败：%v", addr, err)
	}
}
