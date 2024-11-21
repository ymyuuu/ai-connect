package util

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
)

var defaultProxyClient = http.Client{}

type HttpProxy struct {
	ProxyClient *http.Client
	EchoContext *echo.Context
}

func NewHttpProxy(echoContext *echo.Context) *HttpProxy {
	return &HttpProxy{ProxyClient: &defaultProxyClient, EchoContext: echoContext}
}

func (h *HttpProxy) Do() (*http.Response, error) {
	c := *h.EchoContext
	u := c.Request().URL
	sourceHost := u.Host
	// 构建目标url
	targetUrl := BuildTargetUrl(u)

	// 构建目标headers
	targetHeaders := make(http.Header)
	for k, v := range c.Request().Header {
		if FilterHeader(k) {
			continue
		}
		newV := strings.ReplaceAll(strings.Join(v, ","), sourceHost, targetUrl.Host)
		targetHeaders.Add(k, newV)
	}

	targetHeaders.Set("Referer", targetUrl.String())
	targetHeaders.Set("Origin", targetUrl.Scheme+"://"+targetUrl.Host)

	reqBs, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	// 重置body
	reqBody := strings.ReplaceAll(string(reqBs), sourceHost, targetUrl.Host)
	reqReader := strings.NewReader(reqBody)
	// 转发请求到目标url
	req, err := http.NewRequest(c.Request().Method, targetUrl.String(), reqReader)
	if err != nil {
		return nil, err
	}
	req.Header = targetHeaders
	if NeedAuth(u.Path) {
		token, err := c.Cookie("token")

		if err == nil && token.Value != "" {
			req.Header.Set("Authorization", "Bearer "+DealToken(token.Value))
		}
	}
	return h.ProxyClient.Do(req)
}
