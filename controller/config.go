package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// GetAnnouncements godoc
// @Summary      获取公告列表
// @Description  获取当前生效的公告列表，按创建时间倒序排列
// @Tags         config
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /config/announcements [get]
func GetAnnouncements(c *gin.Context) {
	var announcements []model.Announcement
	if err := config.DB.Where("is_active = ?", true).Order("created_at desc").Find(&announcements).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取公告失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, announcements)
}

// GetCampusMap godoc
// @Summary      获取校园地图数据
// @Description  获取GeoJSON格式的校园地图数据
// @Tags         config
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /config/campus-map [get]
func GetCampusMap(c *gin.Context) {
	var features []model.CampusMapFeature
	if err := config.DB.Find(&features).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取校园地图数据失败: "+err.Error())
		return
	}

	geoJsonFeatures := make([]map[string]interface{}, 0, len(features))
	for _, f := range features {
		geoJsonFeatures = append(geoJsonFeatures, map[string]interface{}{
			"type":       f.Type,
			"properties": f.Properties,
			"geometry":   f.Geometry,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "FeatureCollection",
		"features": geoJsonFeatures,
	})
}
