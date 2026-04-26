package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/service"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// Login godoc
// @Summary      用户登录
// @Description  使用Token进行登录，如果用户不存在则自动注册。登录成功后返回含有JWT的凭证信息。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "登录请求，需包含获取的token"
// @Success      200      {object}  dto.LoginResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的参数，请提供 token")
		return
	}

	loginResponse, err := h.authService.Login(req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			response.ResponseError(c, http.StatusUnauthorized, "获取用户信息失败或 Token 已过期")
		case errors.Is(err, service.ErrUserQueryFailed):
			log.Printf("[ERROR] 查询用户失败: %v", err)
			response.ResponseError(c, http.StatusInternalServerError, "数据库查询出错")
		case errors.Is(err, service.ErrUserCreateFailed):
			log.Printf("[ERROR] 创建用户失败: %v", err)
			response.ResponseError(c, http.StatusInternalServerError, "创建用户失败")
		case errors.Is(err, service.ErrTokenGenerateFailed):
			log.Printf("[ERROR] 生成令牌失败: %v", err)
			response.ResponseError(c, http.StatusInternalServerError, "生成令牌失败")
		default:
			log.Printf("[ERROR] 登录失败: %v", err)
			response.ResponseError(c, http.StatusInternalServerError, "登录失败")
		}
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}
