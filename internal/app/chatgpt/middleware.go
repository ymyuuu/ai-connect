package chatgpt

import (
	"github.com/dhbin/ai-connect/internal/common/web"
	"github.com/labstack/echo/v4"
	"strings"
)

func RebuildRequestURL(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := c.Request().URL

		xForwardProto := c.Request().Header.Get("x-forwarded-proto")
		if xForwardProto != "" {
			u.Scheme = xForwardProto
		} else {
			if u.Scheme == "" {
				if c.Request().TLS == nil {
					u.Scheme = "http"
				} else {
					u.Scheme = "https"
				}
			}
		}
		isWebsocket := web.IsWebsocket(c)
		if isWebsocket {
			u.Scheme = strings.ReplaceAll(u.Scheme, "http", "ws")
		}

		if u.Host == "" {
			u.Host = c.Request().Host
		}
		if u.Path == "" {
			u.Path = c.Request().RequestURI
		}

		return next(c)
	}
}
