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

type appVersionsRequest struct {
	Platform string `form:"platform" binding:"required,oneof=ios android"`
}

// GetAppVersions godoc
// @Summary      获取App所有版本
// @Description  获取指定平台的所有App版本历史
// @Tags         config
// @Produce      json
// @Param        platform  query     string  true  "平台(ios或android)" Enums(ios, android)
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /config/app-versions [get]
func GetAppVersions(c *gin.Context) {
	var req appVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var versions []model.AppVersion
	if err := config.DB.Where("platform = ?", req.Platform).Order("version_code desc").Find(&versions).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取版本信息失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, versions)
}

type checkVersionRequest struct {
	Platform           string `form:"platform" binding:"required,oneof=ios android"`
	CurrentVersionCode int    `form:"current_version_code" binding:"required"`
}

// CheckAppVersion godoc
// @Summary      检查App版本更新
// @Description  检查指定平台的App是否有更新
// @Tags         config
// @Produce      json
// @Param        platform              query     string  true  "平台(ios或android)" Enums(ios, android)
// @Param        current_version_code  query     int     true  "当前版本号"
// @Success      200                   {object}  map[string]interface{}
// @Failure      400                   {object}  map[string]interface{}
// @Failure      500                   {object}  map[string]interface{}
// @Router       /config/app-version/check [get]
func CheckAppVersion(c *gin.Context) {
	var req checkVersionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var latestVersion model.AppVersion
	err := config.DB.Where("platform = ?", req.Platform).Order("version_code desc").First(&latestVersion).Error
	if err != nil {
		// 没有该平台的版本记录
		c.JSON(http.StatusOK, gin.H{
			"has_update":      false,
			"is_force_update": false,
			"latest_version":  nil,
		})
		return
	}

	hasUpdate := latestVersion.VersionCode > req.CurrentVersionCode
	var isForceUpdate bool
	if hasUpdate {
		var forceUpdate model.AppVersion
		err := config.DB.Select("id").Where("platform = ? AND version_code > ? AND is_force_update = ?", req.Platform, req.CurrentVersionCode, true).First(&forceUpdate).Error
		if err == nil {
			isForceUpdate = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"has_update":      hasUpdate,
		"is_force_update": isForceUpdate,
		"latest_version":  latestVersion,
	})
}
