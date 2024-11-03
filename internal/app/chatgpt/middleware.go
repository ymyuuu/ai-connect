package chatgpt

import (
	"github.com/dhbin/ai-connect/internal/common/web"
	"github.com/labstack/echo/v4"
)

func RebuildRequestURL(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := c.Request().URL

		xForwardProto := c.Request().Header.Get("x-forwarded-proto")
		if xForwardProto != "" {
			u.Scheme = xForwardProto
		} else {
			if u.Scheme == "" {
				isWebsocket := web.IsWebsocket(c)
				if c.Request().TLS == nil {
					if isWebsocket {
						u.Scheme = "ws"
					} else {
						u.Scheme = "http"
					}
				} else {
					if isWebsocket {
						u.Scheme = "wss"
					} else {
						u.Scheme = "https"
					}
				}
			}
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
