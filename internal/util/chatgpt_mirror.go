package util

import (
	"github.com/dhbin/ai-connect/internal/config"
	"net/http"
	"net/url"
	"strings"
)

var ignoreHeadersMap = map[string]interface{}{
	"cf-warp-tag-id":                nil,
	"cf-visitor":                    nil,
	"cf-ray":                        nil,
	"cf-request-id":                 nil,
	"cf-worker":                     nil,
	"cf-access-client-id":           nil,
	"cf-access-client-device-type":  nil,
	"cf-access-client-device-model": nil,
	"cf-access-client-device-name":  nil,
	"cf-access-client-device-brand": nil,
	"cf-connecting-ip":              nil,
	"cf-ipcountry":                  nil,
	"x-real-ip":                     nil,
	"x-forwarded-for":               nil,
	"x-forwarded-proto":             nil,
	"x-forwarded-port":              nil,
	"x-forwarded-host":              nil,
	"x-forwarded-server":            nil,
	"cdn-loop":                      nil,
	"remote-host":                   nil,
	"x-frame-options":               nil,
	"x-xss-protection":              nil,
	"x-content-type-options":        nil,
	"content-security-policy":       nil,
	"host":                          nil,
	"cookie":                        nil,
	"connection":                    nil,
	"content-length":                nil,
	"content-encoding":              nil,
	"x-middleware-prefetch":         nil,
	"x-nextjs-data":                 nil,
	"x-forwarded-uri":               nil,
	"x-forwarded-path":              nil,
	"x-forwarded-method":            nil,
	"x-forwarded-protocol":          nil,
	"x-forwarded-scheme":            nil,
	"authorization":                 nil,
	"referer":                       nil,
	"origin":                        nil,
}

func BuildTargetUrl(sourceUrl *url.URL) *url.URL {
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

func DealToken(token string) string {
	if strings.HasPrefix(token, "eyJhbGci") {
		return token
	}
	return config.ChatGptMirror().Tokens[token]
}

func NeedAuth(path string) bool {
	return !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, ".webp")

}

func BodyNeedHandle(u *url.URL) bool {
	return strings.HasSuffix(u.Path, ".js") || strings.HasSuffix(u.Path, ".css") || u.Path == "/backend-api/me"
}

func SetIfNotEmpty(target, header http.Header, key string) {
	v := header.Get(key)
	if v != "" {
		target.Set(key, v)
	}
}

func FilterHeader(header string) bool {
	_, exists := ignoreHeadersMap[strings.ToLower(header)]
	return exists
}
