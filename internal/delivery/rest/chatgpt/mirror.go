package chatgpt

import (
	"encoding/json"
	"github.com/dhbin/ai-connect/internal/common"
	"github.com/dhbin/ai-connect/internal/common/code"
	"github.com/dhbin/ai-connect/internal/domain"
	"github.com/dhbin/ai-connect/internal/util"
	"github.com/dhbin/ai-connect/templates"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
	"time"
)

var proxyClient = http.Client{}

func HandleIndex(c echo.Context) error {
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
	data := map[string]string{
		"StaticPrefixUrl": c.Scheme() + "://" + c.Request().Host,
		"Token":           token,
	}
	return c.Render(http.StatusOK, "index.html", data)
}

func ReturnHookJs(c echo.Context) error {
	bs, err := templates.TemplateFs.ReadFile("chatgpt/hook.js")
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/javascript", bs)
}

func HandleGpts(c echo.Context) error {
	if c.QueryParam("_data") == "routes/gpts._index" {
		return c.JSON(http.StatusOK, domain.CheckGpts{
			Kind:     "store",
			Referrer: "https://chatgpt.com/",
		})
	}
	return HandleGIndex(c)
}

func HandleGIndex(c echo.Context) error {
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
	return c.Render(http.StatusOK, "gpts.html", data)
}

func Handle(c echo.Context) error {

	u := c.Request().URL
	sourceHost := u.Host

	if strings.HasSuffix(u.Path, ".map") {
		return c.NoContent(http.StatusNotFound)
	}

	if strings.HasPrefix(u.Path, "/g/") && c.QueryParam("_data") == "" {
		return HandleGIndex(c)
	}

	// 构建目标url
	targetUrl := util.BuildTargetUrl(u)

	// 构建目标headers
	targetHeaders := make(http.Header)
	for k, v := range c.Request().Header {
		if util.FilterHeader(k) {
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
	if util.NeedAuth(u.Path) {
		token, err := c.Cookie("token")

		if err == nil && token.Value != "" {
			req.Header.Set("Authorization", "Bearer "+util.DealToken(token.Value))
		}
	}

	resp, err := proxyClient.Do(req)
	if err != nil {
		return err
	}

	contentEncoding := resp.Header.Get("Content-Encoding")

	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Content-Encoding")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Content-Type")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Cache-Control")
	util.SetIfNotEmpty(c.Response().Header(), resp.Header, "Expires")

	if strings.HasPrefix(u.Path, "/g/") && c.QueryParam("_data") == "routes/g.$gizmoId._index" {
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
