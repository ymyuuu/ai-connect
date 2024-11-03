package chatgpt

import (
	"github.com/dhbin/ai-connect/internal/config"
	"net/http"
	"net/url"
	"strings"
)

func buildTargetUrl(sourceUrl *url.URL) *url.URL {
	targetUrl, _ := url.Parse(sourceUrl.String())
	// 判断url前缀是否/static，是的话替换host为cdn.oaistatic.com
	if strings.HasPrefix(targetUrl.Path, "/assets") {
		targetUrl.Host = "cdn.oaistatic.com"
	} else if strings.HasPrefix(targetUrl.Path, "/ab") {
		// 判断url前缀是否/ab，是的话替换host为ab.chatgpt.com
		targetUrl.Host = "ab.chatgpt.com"
		// 去除前缀
		targetUrl.Path = strings.TrimPrefix(targetUrl.Path, "/ab")
	} else if strings.HasPrefix(targetUrl.Path, "/webrtc") {
		targetUrl.Host = "webrtc.chatgpt.com"
		// 去除前缀
		targetUrl.Path = strings.TrimPrefix(targetUrl.Path, "/webrtc")
	} else {
		// 其他情况，替换host为chatgpt.com
		targetUrl.Host = "chatgpt.com"
	}
	if targetUrl.Scheme == "ws" {
		targetUrl.Scheme = "wss"
	} else if targetUrl.Scheme == "http" {
		targetUrl.Scheme = "https"
	}
	return targetUrl
}

func dealToken(token string) string {
	if strings.HasPrefix(token, "eyJhbGci") {
		return token
	}
	return config.ChatGptMirror().Tokens[token]
}

func needAuth(path string) bool {
	return !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, ".webp")

}

func bodyNeedHandle(u *url.URL) bool {
	return strings.HasSuffix(u.Path, ".js") || strings.HasSuffix(u.Path, ".css") || u.Path == "/backend-api/me"
}

func setIfNotEmpty(target, header http.Header, key string) {
	v := header.Get(key)
	if v != "" {
		target.Set(key, v)
	}
}
