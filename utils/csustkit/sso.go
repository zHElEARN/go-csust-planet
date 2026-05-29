package csustkit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const ehallLoginService = "https%3A%2F%2Fehall.csust.edu.cn%2Flogin"

type SSOHelper struct {
	client *Client
}

type SSOLoginForm struct {
	PwdEncryptSalt string
	Execution      string
}

func (h *SSOHelper) IsLoggedIn(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.client.makeURL(ServiceEhall, "/getLoginUser"), nil)
	if err != nil {
		return false
	}

	resp, err := h.client.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	var loginResp struct {
		Data *json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return false
	}
	return loginResp.Data != nil && string(*loginResp.Data) != "null"
}

func (h *SSOHelper) GetLoginForm(ctx context.Context) (SSOLoginForm, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.client.makeURL(ServiceAuthServer, "/authserver/login?service="+ehallLoginService), nil)
	if err != nil {
		return SSOLoginForm{}, err
	}

	resp, err := h.client.httpClient.Do(req)
	if err != nil {
		return SSOLoginForm{}, fmt.Errorf("获取登录表单失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.Request != nil && resp.Request.URL.String() == h.client.makeURL(ServiceEhall, "/index.html") {
		return SSOLoginForm{}, fmt.Errorf("获取登录表单失败: 账号已登录")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return SSOLoginForm{}, fmt.Errorf("解析登录表单失败: %w", err)
	}

	pwdEncryptSalt, ok := doc.Find("input#pwdEncryptSalt").First().Attr("value")
	if !ok {
		return SSOLoginForm{}, fmt.Errorf("获取登录表单失败: 未找到pwdEncryptSalt输入框")
	}
	execution, ok := doc.Find("input#execution").First().Attr("value")
	if !ok {
		return SSOLoginForm{}, fmt.Errorf("获取登录表单失败: 未找到execution输入框")
	}

	return SSOLoginForm{PwdEncryptSalt: pwdEncryptSalt, Execution: execution}, nil
}

func (h *SSOHelper) Login(ctx context.Context, form SSOLoginForm, username, password, captcha string) error {
	encryptedPassword, err := encryptPassword(password, form.PwdEncryptSalt)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	values := url.Values{}
	values.Set("username", username)
	values.Set("password", encryptedPassword)
	values.Set("captcha", captcha)
	values.Set("_eventId", "submit")
	values.Set("cllt", "userNameLogin")
	values.Set("dllt", "generalLogin")
	values.Set("lt", "")
	values.Set("execution", form.Execution)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.client.makeURL(ServiceAuthServer, "/authserver/login?service="+ehallLoginService), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("登录统一身份认证失败: %w", err)
	}
	defer resp.Body.Close()

	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	if finalURL == h.client.makeURL(ServiceEhall, "/index.html") || finalURL == h.client.makeURL(ServiceEhall, "/default/index.html") {
		return nil
	}

	if message, ok := loginErrorMessage(resp); ok {
		return fmt.Errorf("登录统一身份认证失败: %s", message)
	}
	if finalURL == "" {
		return fmt.Errorf("登录统一身份认证失败: 未找到重定向链接")
	}
	return fmt.Errorf("登录统一身份认证失败: 重定向URL异常: %s", finalURL)
}

func (h *SSOHelper) LoginToCampusCard(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.client.makeURL(ServiceCampusCard, "/berserker-auth/cas/login/wisedu?targetUrl=https://hxyxh5.csust.edu.cn/plat/?name=loginTransit"), nil)
	if err != nil {
		return "", err
	}

	resp, err := h.client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("登录校园卡系统失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.Request == nil || resp.Request.URL == nil {
		return "", fmt.Errorf("登录校园卡系统失败: 未找到重定向URL")
	}
	finalURL := resp.Request.URL
	if !strings.HasPrefix(finalURL.String(), h.client.makeURL(ServiceCampusCard, "/plat")) {
		if finalURL.Host == "authserver.csust.edu.cn" && strings.HasPrefix(finalURL.Path, "/authserver/login") {
			return "", ErrSSONotLoggedIn
		}
		return "", fmt.Errorf("登录校园卡系统失败: 重定向URL异常: %s", finalURL.String())
	}

	ticket := finalURL.Query().Get("ticket")
	if ticket == "" {
		return "", fmt.Errorf("登录校园卡系统失败: 无法获取登录凭据")
	}
	decodedTicket, err := url.QueryUnescape(ticket)
	if err != nil {
		return "", fmt.Errorf("登录校园卡系统失败: 解码登录凭据失败: %w", err)
	}
	return decodedTicket, nil
}

func loginErrorMessage(resp *http.Response) (string, bool) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", false
	}

	message := strings.TrimSpace(doc.Find("#showErrorTip").First().Text())
	if message == "" {
		return "", false
	}
	return message, true
}
