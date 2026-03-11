package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Profile 用户信息
type Profile struct {
	CategoryName      string  `json:"categoryName"`      // 学生类别
	UserAccount       string  `json:"userAccount"`       // 学号
	UserName          string  `json:"userName"`          // 姓名
	CertCode          string  `json:"certCode"`          // 身份证号（打码）
	Phone             string  `json:"phone"`             // 手机号（打码）
	Email             *string `json:"email"`             // 邮箱
	DeptName          string  `json:"deptName"`          // 学院名称
	DefaultUserAvatar string  `json:"defaultUserAvatar"` // 默认头像链接
	HeadImageIcon     *string `json:"headImageIcon"`     // 用户设置的头像链接
	Avatar            string  `json:"avatar"`            // 最终头像链接
}

// getAvatarURL 获取头像链接
func (p *Profile) getAvatarURL() string {
	if p.HeadImageIcon != nil && *p.HeadImageIcon != "" {
		return *p.HeadImageIcon
	}
	return p.DefaultUserAvatar
}

type ssoResponse struct {
	ErrCode string   `json:"errcode"`
	ErrMsg  string   `json:"errmsg"`
	Data    *Profile `json:"data"`
}

// GetUserProfile 根据 token 获取用户信息
func GetUserProfile(token string) (*Profile, error) {
	req, err := http.NewRequest("GET", "https://ehall.csust.edu.cn/getLoginUser", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "WISCPSID",
		Value: token,
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %w", err)
	}

	var ssoResp ssoResponse
	if err := json.Unmarshal(body, &ssoResp); err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	if ssoResp.Data == nil {
		return nil, fmt.Errorf("Token 无效或数据为空，错误码: %s，错误信息: %s", ssoResp.ErrCode, ssoResp.ErrMsg)
	}

	ssoResp.Data.Avatar = ssoResp.Data.getAvatarURL()
	return ssoResp.Data, nil
}
