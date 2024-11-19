package chatgpt

import (
	"github.com/dhbin/ai-connect/internal/common/web"
	"github.com/dhbin/ai-connect/internal/config"
	"github.com/dhbin/ai-connect/internal/delivery/rest/chatgpt"
	middleware2 "github.com/dhbin/ai-connect/internal/delivery/rest/middleware"
	"github.com/dhbin/ai-connect/internal/util"
	"github.com/dhbin/ai-connect/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"html/template"
	"net/http"
)

// RunMirror 运行chatgpt镜像
func RunMirror() {
	mirrorConfig := config.ChatGptMirror()
	tls := mirrorConfig.Tls

	e := echo.New()

	// 创建并加载模板
	tmpl := &templates.Template{
		Templates: template.Must(template.ParseFS(templates.TemplateFs, "chatgpt/*.html")),
	}
	e.Renderer = tmpl

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware2.RebuildRequestURL)

	e.GET("/", chatgpt.HandleIndex)
	e.GET("/chatgpt/hook.js", chatgpt.ReturnHookJs)
	e.GET("/c/*", chatgpt.HandleIndex)
	e.POST("/backend-api/accounts/logout_all", func(c echo.Context) error {
		return c.JSON(http.StatusForbidden, nil)
	})
	e.GET("/gpts", chatgpt.HandleGpts)
	e.Any("/webrtc/*", web.ProxyWebSocket(util.BuildTargetUrl))
	e.Any("/*", chatgpt.Handle)

	if tls.Enabled {
		err := e.StartTLS(mirrorConfig.Address, tls.Cert, tls.Key)
		cobra.CheckErr(err)
	} else {
		err := e.Start(mirrorConfig.Address)
		cobra.CheckErr(err)

	}
}
