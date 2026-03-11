package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils"
)

type loginRequest struct {
	Token string `json:"token" binding:"required"`
}

// Login godoc
// @Summary      用户登录
// @Description  使用Token进行登录，如果用户不存在则自动注册。登录成功后返回含有JWT的凭证信息。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      loginRequest  true  "登录请求，需包含获取的token"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的参数，请提供 token")
		return
	}

	// 获取用户信息
	profile, err := utils.GetUserProfile(req.Token)
	if err != nil {
		// 如果无法获取用户信息，通常是 token 无效
		utils.ResponseError(c, http.StatusUnauthorized, "获取用户信息失败或 Token 已过期")
		return
	}

	// 查找或创建用户
	var user model.User
	result := config.DB.Where("student_id = ?", profile.UserAccount).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果不存在，则新建用户记录
			user = model.User{
				StudentID: profile.UserAccount,
			}
			if err := config.DB.Create(&user).Error; err != nil {
				utils.ResponseError(c, http.StatusInternalServerError, "创建用户失败: "+err.Error())
				return
			}
		} else {
			utils.ResponseError(c, http.StatusInternalServerError, "数据库查询出错: "+result.Error.Error())
			return
		}
	}

	// 生成 JWT
	jwtToken, err := utils.GenerateToken(user.ID, 30*24*time.Hour)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "生成令牌失败: "+err.Error())
		return
	}

	// 返回 JWT 和用户信息
	c.JSON(http.StatusOK, gin.H{
		"token":   jwtToken,
		"profile": profile,
	})
}
