package chatgpt

import (
	"github.com/dhbin/ai-connect/internal/config"
	"github.com/dhbin/ai-connect/internal/delivery/rest/chatgpt"
	middleware2 "github.com/dhbin/ai-connect/internal/delivery/rest/middleware"
	"github.com/dhbin/ai-connect/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"html/template"
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

	// chatgpt镜像路由
	chatgpt.NewMirrorHandler(e)

	if tls.Enabled {
		err := e.StartTLS(mirrorConfig.Address, tls.Cert, tls.Key)
		cobra.CheckErr(err)
	} else {
		err := e.Start(mirrorConfig.Address)
		cobra.CheckErr(err)

	}
}
