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

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数，请提供 token"})
		return
	}

	// 获取用户信息
	profile, err := utils.GetUserProfile(req.Token)
	if err != nil {
		// 如果无法获取用户信息，通常是 token 无效
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户信息失败或 Token 已过期"})
		return
	}

	// 数据库操作：查找或创建用户
	var user model.User
	result := config.DB.Where("student_id = ?", profile.UserAccount).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果不存在，则新建用户记录
			user = model.User{
				StudentID: profile.UserAccount,
			}
			if err := config.DB.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询出错: " + result.Error.Error()})
			return
		}
	}

	// 生成 JWT
	jwtToken, err := utils.GenerateToken(user.ID, 30*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败: " + err.Error()})
		return
	}

	// 4. 返回 JWT 和用户信息
	c.JSON(http.StatusOK, gin.H{
		"token":   jwtToken,
		"profile": profile,
		"avatar":  profile.Avatar(),
	})
}
