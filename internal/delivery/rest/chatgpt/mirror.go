package chatgpt

import (
	"encoding/json"
	"github.com/dhbin/ai-connect/internal/common"
	"github.com/dhbin/ai-connect/internal/common/code"
	"github.com/dhbin/ai-connect/internal/common/web"
	"github.com/dhbin/ai-connect/internal/domain"
	"github.com/dhbin/ai-connect/internal/util"
	"github.com/dhbin/ai-connect/templates"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
	"time"
)

type MirrorHandler struct {
}

func NewMirrorHandler(e *echo.Echo) {
	m := &MirrorHandler{}
	e.GET("/", m.HandleIndex)
	e.GET("/chatgpt/hook.js", m.ReturnHookJs)
	e.GET("/c/*", m.HandleIndex)
	e.POST("/backend-api/accounts/logout_all", func(c echo.Context) error {
		return c.JSON(http.StatusForbidden, nil)
	})
	e.GET("/gpts", m.HandleGptsIndex)
	e.Any("/webrtc/*", web.ProxyWebSocket(util.BuildTargetUrl))
	e.GET(".map", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})
	e.Any("/g/*", m.HandleGptsSession)
	e.POST("/backend-api/conversation", m.HandleConversation)
	e.Any("/*", m.Handle)
}

func (m *MirrorHandler) HandleIndex(c echo.Context) error {
	token := "announce"
	t := c.QueryParam("token")
	if t != "" {
		token = t
		c.SetCookie(&http.Cookie{
			Name:    "token",
			Value:   token,
			Expires: time.Now().Add(24 * time.Hour),
		})
	}

	// 这里硬编码为 https://example.com
	data := map[string]string{
		"StaticPrefixUrl": "https://obbi.ymyuuu.workers.dev",
		"Token":           token,
	}
	return c.Render(http.StatusOK, "index.html", data)
}

func (m *MirrorHandler) ReturnHookJs(c echo.Context) error {
	bs, err := templates.TemplateFs.ReadFile("chatgpt/hook.js")
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/javascript", bs)
}

func (m *MirrorHandler) HandleGptsIndex(c echo.Context) error {
	if c.QueryParam("_data") == "routes/gpts._index" {
		return c.JSON(http.StatusOK, domain.CheckGpts{
			Kind:     "store",
			Referrer: "https://chatgpt.com/",
		})
	}
	return m.HandleGIndex(c)
}

func (m *MirrorHandler) HandleGIndex(c echo.Context) error {
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

	// 同样这里也硬编码
	data := map[string]string{
		"StaticPrefixUrl": "https://example.com",
		"Token":           token,
	}
	return c.Render(http.StatusOK, "gpts.html", data)
}

func (m *MirrorHandler) Handle(c echo.Context) error {
	u := c.Request().URL
	sourceHost := u.Host

	resp, err := util.NewHttpProxy(&c).Do()
	if err != nil {
		return err
	}

	contentEncoding := resp.Header.Get("Content-Encoding")
	setResponseHeaders(c, resp)

	// 设置响应状态码
	c.Response().WriteHeader(resp.StatusCode)

	if util.BodyNeedHandle(u) && resp.StatusCode < http.StatusMultipleChoices {
		reader, err := code.WarpReader(resp.Body, contentEncoding)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(reader.Close)
		bs, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		body := string(bs)

		if u.Path == "/backend-api/me" {
			var meJson domain.Me
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
			sourceScheme := u.Scheme
			body = strings.ReplaceAll(body, "https://chatgpt.com", sourceScheme+"://"+sourceHost)
			body = strings.ReplaceAll(body, "https://ab.chatgpt.com", sourceScheme+"://"+sourceHost+"/ab")
			body = strings.ReplaceAll(body, "https://cdn.oaistatic.com", sourceScheme+"://"+sourceHost)
			body = strings.ReplaceAll(body, "chatgpt.com", sourceHost)
		}

		writer, err := code.WarpWriter(c.Response(), contentEncoding)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(writer.Close)
		_, err = writer.Write([]byte(body))
		return err
	}

	_, _ = io.Copy(c.Response(), resp.Body)

	return nil
}

func (m *MirrorHandler) HandleGptsSession(c echo.Context) error {
	if c.QueryParam("_data") == "" {
		return m.HandleGIndex(c)
	}
	if c.QueryParam("_data") == "routes/g.$gizmoId._index" {
		resp, err := util.NewHttpProxy(&c).Do()
		if err != nil {
			return err
		}
		setResponseHeaders(c, resp)
		contentEncoding := resp.Header.Get("Content-Encoding")
		reader, err := code.WarpReader(resp.Body, contentEncoding)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(reader.Close)

		targetBs, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		r := make(map[string]interface{})
		err = json.Unmarshal(targetBs, &r)
		if err != nil {
			return err
		}

		for k, v := range domain.GptsInfoInject {
			r[k] = v
		}

		writer, err := code.WarpWriter(c.Response(), contentEncoding)
		if err != nil {
			return err
		}
		defer common.IgnoreErr(writer.Close)
		newR, err := json.Marshal(r)
		if err != nil {
			return err
		}
		_, err = writer.Write(newR)

		return err
	}
	return m.Handle(c)
}

func (m *MirrorHandler) HandleConversation(c echo.Context) error {
	resp, err := util.NewHttpProxy(&c).Do()
	if err != nil {
		return err
	}
	setResponseHeaders(c, resp)
	contentEncoding := resp.Header.Get("Content-Encoding")
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

func setResponseHeaders(c echo.Context, resp *http.Response) {
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Content-Encoding")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Content-Type")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Cache-Control")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Expires")
}
