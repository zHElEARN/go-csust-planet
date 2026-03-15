package dto

import "github.com/zHElEARN/go-csust-planet/utils/sso"

type LoginRequest struct {
	Token string `json:"token" binding:"required"`
}

type LoginResponse struct {
	Token   string       `json:"token"`
	Profile *sso.Profile `json:"profile"`
}
