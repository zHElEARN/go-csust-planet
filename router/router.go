package router

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zHElEARN/go-csust-planet/campus"
)

var (
	buildingsCache sync.Map
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello world",
		})
	})

	r.GET("/electricity", func(c *gin.Context) {
		campusName := c.Query("campus")
		buildingName := c.Query("building")
		roomNum := c.Query("room")

		if campusName == "" || buildingName == "" || roomNum == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: campus, building, room 均为必填"})
			return
		}

		var targetCampus campus.Campus
		switch campusName {
		case "云塘":
			targetCampus = campus.CampusYuntang
		case "金盆岭":
			targetCampus = campus.CampusJinpenling
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的校区名称"})
			return
		}

		var buildings []campus.Building
		if val, ok := buildingsCache.Load(targetCampus.ID); ok {
			buildings = val.([]campus.Building)
		} else {
			var err error
			buildings, err = campus.GetBuildings(targetCampus)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取楼栋列表失败: " + err.Error()})
				return
			}
			buildingsCache.Store(targetCampus.ID, buildings)
		}

		var targetBuilding *campus.Building
		for _, b := range buildings {
			if b.Name == buildingName {
				targetBuilding = &b
				break
			}
		}

		if targetBuilding == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("在%s未找到楼栋: %s", campusName, buildingName)})
			return
		}

		balance, err := campus.GetElectricity(*targetBuilding, roomNum)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询电费失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"balance":  balance,
			"campus":   campusName,
			"building": buildingName,
			"room":     roomNum,
		})
	})

	r.GET("/profile", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: token 不能为空"})
			return
		}

		profile, err := campus.GetUserProfile(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"profile": profile,
			"avatar":  profile.Avatar(),
		})
	})

	return r
}
