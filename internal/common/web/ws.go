package web

import (
	"github.com/dhbin/ai-connect/internal/common"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ProxyWebSocket(convertUrlFunc func(*url.URL) *url.URL) func(c echo.Context) error {
	return func(c echo.Context) error {
		// Upgrade client connection
		clientConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(clientConn.Close)

		targetURL := convertUrlFunc(c.Request().URL).String()
		targetConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(targetConn.Close)

		// Proxy messages between client and target server
		forward := func(src, dest *websocket.Conn) {
			for {
				messageType, message, err := src.ReadMessage()
				if err != nil {
					break
				}
				err = dest.WriteMessage(messageType, message)
				if err != nil {
					break
				}
			}
		}

		go forward(clientConn, targetConn)
		forward(targetConn, clientConn)

		return nil
	}
}

func IsWebsocket(c echo.Context) bool {
	return c.Request().Header.Get("Upgrade") == "websocket"
}
