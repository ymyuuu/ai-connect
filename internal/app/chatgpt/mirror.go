package chatgpt

import (
	"compress/gzip"
	"encoding/json"
	"github.com/andybalholm/brotli"
	"github.com/dhbin/ai-connect/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Me struct {
	Amr                      []interface{} `json:"amr"`
	Created                  int           `json:"created"`
	Email                    string        `json:"email"`
	Groups                   []interface{} `json:"groups"`
	HasPaygProjectSpendLimit bool          `json:"has_payg_project_spend_limit"`
	Id                       string        `json:"id"`
	MfaFlagEnabled           bool          `json:"mfa_flag_enabled"`
	Name                     string        `json:"name"`
	Object                   string        `json:"object"`
	Orgs                     struct {
		Data []struct {
			Created                        int           `json:"created"`
			Description                    string        `json:"description"`
			Geography                      interface{}   `json:"geography"`
			Groups                         []interface{} `json:"groups"`
			Id                             string        `json:"id"`
			IsDefault                      bool          `json:"is_default"`
			IsScaleTierAuthorizedPurchaser interface{}   `json:"is_scale_tier_authorized_purchaser"`
			IsScimManaged                  bool          `json:"is_scim_managed"`
			Name                           string        `json:"name"`
			Object                         string        `json:"object"`
			ParentOrgId                    interface{}   `json:"parent_org_id"`
			Personal                       bool          `json:"personal"`
			Projects                       struct {
				Data   []interface{} `json:"data"`
				Object string        `json:"object"`
			} `json:"projects"`
			Role     string `json:"role"`
			Settings struct {
				DisableUserApiKeys       bool   `json:"disable_user_api_keys"`
				ThreadsUiVisibility      string `json:"threads_ui_visibility"`
				UsageDashboardVisibility string `json:"usage_dashboard_visibility"`
			} `json:"settings"`
			Title string `json:"title"`
		} `json:"data"`
		Object string `json:"object"`
	} `json:"orgs"`
	PhoneNumber interface{} `json:"phone_number"`
	Picture     string      `json:"picture"`
}

// 忽略请求头key列表
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

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func RunMirror() {
	mirrorConfig := config.ChatGptMirror()
	tls := mirrorConfig.Tls
	tokens := mirrorConfig.Tokens

	proxyClient := http.Client{}

	e := echo.New()

	// 创建并加载模板
	tmpl := &Template{
		templates: template.Must(template.ParseGlob("templates/chatgpt/*.html")),
	}
	e.Renderer = tmpl

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.GET("/", HandleIndex)
	e.GET("/c/*", HandleIndex)
	e.GET("/g/*", HandleIndex)
	e.Any("/*", func(c echo.Context) error {

		u := buildUrl(c)
		sc := u.Scheme

		// 构建目标url
		sourceUrl := u.String()
		targetUrl, _ := url.Parse(sourceUrl)
		// 判断url前缀是否/static，是的话替换host为cdn.oaistatic.com
		if strings.HasPrefix(targetUrl.Path, "/assets") {
			targetUrl.Host = "cdn.oaistatic.com"
		} else if strings.HasPrefix(targetUrl.Path, "/ab") {
			// 判断url前缀是否/ab，是的话替换host为ab.chatgpt.com
			targetUrl.Host = "ab.chatgpt.com"
			// 去除前缀
			targetUrl.Path = strings.TrimPrefix(targetUrl.Path, "/ab")
		} else {
			// 其他情况，替换host为chatgpt.com
			targetUrl.Host = "chatgpt.com"
		}

		// 构建目标headers
		sourceHost := u.Host
		targetHeaders := make(http.Header)
		for k, v := range c.Request().Header {
			if contains(strings.ToLower(k)) {
				continue
			}
			newV := strings.ReplaceAll(strings.Join(v, ","), sourceHost, targetUrl.Host)
			targetHeaders.Add(k, newV)
		}

		targetHeaders.Set("Referer", targetUrl.String())
		targetHeaders.Set("Origin", targetUrl.Scheme+"://"+targetUrl.Host)

		reqBs, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		// 重置body
		reqBody := strings.ReplaceAll(string(reqBs), sourceHost, targetUrl.Host)
		reqReader := strings.NewReader(reqBody)

		// 转发请求到目标url
		req, err := http.NewRequest(c.Request().Method, targetUrl.String(), reqReader)
		if err != nil {
			return err
		}
		req.Header = targetHeaders
		if !strings.HasSuffix(u.Path, ".js") && !strings.HasSuffix(u.Path, ".css") && !strings.HasSuffix(u.Path, ".webp") {
			token, err := c.Cookie("token")

			if err == nil && token.Value != "" {
				req.Header.Set("Authorization", "Bearer "+tokens[token.Value])
			}
		}

		resp, err := proxyClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		c.Response().Header().Set("Content-Encoding", resp.Header.Get("Content-Encoding"))
		c.Response().Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		c.Response().Header().Set("Cache-Control", resp.Header.Get("Cache-Control"))
		c.Response().Header().Set("Expires", resp.Header.Get("Expires"))

		if u.Path == "/backend-api/conversation" {
			bs := make([]byte, 1)
			for {
				n, err := resp.Body.Read(bs)
				if err != nil && err != io.EOF {
					log.Println("read response body error: ", err)
					return nil
				}
				if n == 0 {
					break
				}
				_, _ = c.Response().Write(bs)
				c.Response().Flush()
			}
			return nil
		}

		// 设置响应状态码
		c.Response().WriteHeader(resp.StatusCode)

		if strings.HasSuffix(u.Path, ".js") || strings.HasSuffix(u.Path, ".css") || u.Path == "/backend-api/me" {
			// 读取响应内容，处理可能的压缩
			var reader io.ReadCloser
			contentEncoding := resp.Header.Get("Content-Encoding")
			switch contentEncoding {
			case "gzip":
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					return err
				}
				defer reader.Close()
			case "br":
				reader = io.NopCloser(brotli.NewReader(resp.Body))
			default:
				reader = resp.Body
			}
			bs, err := io.ReadAll(reader)
			if err != nil {
				return err
			}

			body := string(bs)

			if u.Path == "/backend-api/me" {
				var meJson Me
				err := json.Unmarshal(bs, &meJson)
				if err == nil {
					meJson.Email = "sam@openai.com"
					meJson.PhoneNumber = nil
					meJson.Name = "Sam Altman"
					for i, _ := range meJson.Orgs.Data {
						meJson.Orgs.Data[i].Description = "Personal org for " + meJson.Email
					}

					newMe, err := json.Marshal(meJson)
					if err == nil {
						body = string(newMe)
					}
				}
			} else {
				body = strings.ReplaceAll(body, "https://chatgpt.com", sc+"://"+sourceHost)
				body = strings.ReplaceAll(body, "https://ab.chatgpt.com", sc+"://"+sourceHost+"/ab")
				body = strings.ReplaceAll(body, "https://cdn.oaistatic.com", sc+"://"+sourceHost)
				body = strings.ReplaceAll(body, `if(s)return o.apply(this,arguments)`, `if(arguments[0] && typeof arguments[0] === 'string'){arguments[0] = arguments[0].replace('chatgpt.com', location.host).replace('ab.chatgpt.com', location.host + '/ab').replace('cdn.oaistatic.com', location.host)};if(s)return o.apply(this,arguments)`)
			}

			switch contentEncoding {
			case "gzip":
				// 把body压缩gzip写到res
				gz := gzip.NewWriter(c.Response())
				defer gz.Close()
				_, err = gz.Write([]byte(body))
			case "br":
				brW := brotli.NewWriter(c.Response())
				defer brW.Close()
				_, err = brW.Write([]byte(body))
			default:
				_, err = c.Response().Write([]byte(body))
			}
			return err
		}

		_, _ = io.Copy(c.Response(), resp.Body)

		return nil
	})

	if tls.Enabled {
		err := e.StartTLS(mirrorConfig.Address, tls.Cert, tls.Key)
		cobra.CheckErr(err)
	} else {
		err := e.Start(mirrorConfig.Address)
		cobra.CheckErr(err)

	}
}

func buildUrl(c echo.Context) *url.URL {
	u := c.Request().URL
	// Populate missing fields
	if u.Scheme == "" {
		if c.Request().TLS == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	if u.Host == "" {
		u.Host = c.Request().Host
	}
	if u.Path == "" {
		u.Path = c.Request().RequestURI
	}
	return u
}

func contains(header string) bool {
	_, exists := ignoreHeadersMap[header]
	return exists
}

func HandleIndex(c echo.Context) error {
	token := "announce"
	t := c.QueryParam("token")
	if t != "" {
		token = t
		c.SetCookie(&http.Cookie{
			Name:    "token",
			Value:   token,
			Secure:  true,
			Expires: time.Now().Add(24 * time.Hour),
		})
	}
	data := map[string]string{
		"StaticPrefixUrl": c.Scheme() + "://" + c.Request().Host,
		"Token":           token,
	}
	return c.Render(http.StatusOK, "index.html", data)
}
