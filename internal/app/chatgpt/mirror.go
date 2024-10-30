package chatgpt

import (
	"encoding/json"
	"github.com/dhbin/ai-connect/internal/common"
	"github.com/dhbin/ai-connect/internal/common/code"
	"github.com/dhbin/ai-connect/internal/config"
	"github.com/dhbin/ai-connect/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 忽略请求头key列表

func RunMirror() {
	mirrorConfig := config.ChatGptMirror()
	tls := mirrorConfig.Tls

	proxyClient := http.Client{}

	e := echo.New()

	// 创建并加载模板
	tmpl := &templates.Template{
		Templates: template.Must(template.ParseFS(templates.TemplateFs, "chatgpt/*.html")),
	}
	e.Renderer = tmpl

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.GET("/", handleIndex)
	e.GET("/c/*", handleIndex)
	e.GET("/g/*", handleIndex)
	e.POST("/backend-api/accounts/logout_all", func(c echo.Context) error {
		return c.JSON(http.StatusForbidden, nil)
	})

	e.Any("/*", func(c echo.Context) error {

		u := buildUrl(c)
		sourceHost := u.Host
		sourceScheme := u.Scheme
		sourceUrl := u.String()

		if strings.HasSuffix(u.Path, ".map") {
			return c.NoContent(http.StatusNotFound)
		}

		// 构建目标url
		targetUrl := buildTargetUrl(sourceUrl)

		// 构建目标headers
		targetHeaders := make(http.Header)
		for k, v := range c.Request().Header {
			if filterHeader(k) {
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
		if needAuth(u.Path) {
			token, err := c.Cookie("token")

			if err == nil && token.Value != "" {
				req.Header.Set("Authorization", "Bearer "+dealToken(token.Value))
			}
		}

		resp, err := proxyClient.Do(req)
		if err != nil {
			return err
		}

		contentEncoding := resp.Header.Get("Content-Encoding")

		setIfNotEmpty(c.Response().Header(), resp.Header, "Content-Encoding")
		setIfNotEmpty(c.Response().Header(), resp.Header, "Content-Type")
		setIfNotEmpty(c.Response().Header(), resp.Header, "Cache-Control")
		setIfNotEmpty(c.Response().Header(), resp.Header, "Expires")

		if u.Path == "/backend-api/conversation" {
			reader, err := code.WarpReader(resp.Body, contentEncoding)
			if err != nil {
				return err
			}
			defer common.IgnoreErr(reader.Close)
			writer, err := code.WarpWriter(c.Response(), contentEncoding)
			if err != nil {
				return err
			}
			defer common.IgnoreErr(writer.Close)
			bs := make([]byte, 1)
			for {
				n, err := reader.Read(bs)
				if err != nil && err != io.EOF {
					return err
				}
				if n == 0 {
					break
				}
				_, _ = writer.Write(bs)
				c.Response().Flush()
			}
			return nil
		}

		// 设置响应状态码
		c.Response().WriteHeader(resp.StatusCode)

		if bodyNeedHandle(u) && resp.StatusCode < http.StatusMultipleChoices {
			reader, err := code.WarpReader(resp.Body, contentEncoding)
			if err != nil {
				return err
			}
			defer common.IgnoreErr(reader.Close)
			writer, err := code.WarpWriter(c.Response(), contentEncoding)
			if err != nil {
				return err
			}
			defer common.IgnoreErr(writer.Close)
			bs, err := io.ReadAll(reader)
			if err != nil {
				return err
			}

			body := string(bs)

			if u.Path == "/backend-api/me" {
				var meJson me
				err := json.Unmarshal(bs, &meJson)
				if err == nil {
					meJson.Email = "sam@openai.com"
					meJson.PhoneNumber = nil
					meJson.Name = "Sam Altman"
					for i := range meJson.Orgs.Data {
						meJson.Orgs.Data[i].Description = "Personal org for " + meJson.Email
					}

					newMe, err := json.Marshal(meJson)
					if err == nil {
						body = string(newMe)
					}
				}
			} else {
				body = strings.ReplaceAll(body, "https://chatgpt.com", sourceScheme+"://"+sourceHost)
				body = strings.ReplaceAll(body, "https://ab.chatgpt.com", sourceScheme+"://"+sourceHost+"/ab")
				body = strings.ReplaceAll(body, "https://cdn.oaistatic.com", sourceScheme+"://"+sourceHost)
				body = strings.ReplaceAll(body, "chatgpt.com", sourceHost)
			}

			_, err = writer.Write([]byte(body))
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

func buildTargetUrl(sourceUrl string) *url.URL {
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
	return targetUrl
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

func handleIndex(c echo.Context) error {
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
