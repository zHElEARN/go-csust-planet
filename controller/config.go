package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// GetAnnouncements godoc
// @Summary      获取公告列表
// @Description  获取当前生效的公告列表，按创建时间倒序排列
// @Tags         config
// @Produce      json
// @Success      200  {array}   dto.AnnouncementResponse
// @Failure      500  {object}  map[string]interface{}
// @Router       /config/announcements [get]
func GetAnnouncements(c *gin.Context) {
	var announcements []model.Announcement
	if err := config.DB.Where("is_active = ?", true).Order("created_at desc").Find(&announcements).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取公告失败: "+err.Error())
		return
	}

	res := dto.MapAnnouncements(announcements)
	c.JSON(http.StatusOK, res)
}

// GetCampusMap godoc
// @Summary      获取校园地图数据
// @Description  获取GeoJSON格式的校园地图数据
// @Tags         config
// @Produce      json
// @Success      200  {object}  dto.CampusMapResponse
// @Failure      500  {object}  map[string]interface{}
// @Router       /config/campus-map [get]
func GetCampusMap(c *gin.Context) {
	var features []model.CampusMapFeature
	if err := config.DB.Find(&features).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取校园地图数据失败: "+err.Error())
		return
	}

	res := dto.MapCampusMapFeatures(features)
	c.JSON(http.StatusOK, res)
}

// GetAppVersions godoc
// @Summary      获取App所有版本
// @Description  获取指定平台的所有App版本历史
// @Tags         config
// @Produce      json
// @Param        platform  query     string  true  "平台(ios或android)" Enums(ios, android)
// @Success      200       {array}   dto.AppVersionResponse
// @Failure      400       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /config/app-versions [get]
func GetAppVersions(c *gin.Context) {
	var req dto.AppVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var versions []model.AppVersion
	if err := config.DB.Where("platform = ?", req.Platform).Order("version_code desc").Find(&versions).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取版本信息失败: "+err.Error())
		return
	}

	res := dto.MapAppVersions(versions)
	c.JSON(http.StatusOK, res)
}

// CheckAppVersion godoc
// @Summary      检查App版本更新
// @Description  检查指定平台的App是否有更新
// @Tags         config
// @Produce      json
// @Param        platform              query     string  true  "平台(ios或android)" Enums(ios, android)
// @Param        currentVersionCode    query     int     true  "当前版本号"
// @Success      200                   {object}  dto.CheckAppVersionResponse
// @Failure      400                   {object}  map[string]interface{}
// @Failure      500                   {object}  map[string]interface{}
// @Router       /config/app-version/check [get]
func CheckAppVersion(c *gin.Context) {
	var req dto.CheckAppVersionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var latestVersion model.AppVersion
	err := config.DB.Where("platform = ?", req.Platform).Order("version_code desc").First(&latestVersion).Error
	if err != nil {
		// 没有该平台的版本记录
		c.JSON(http.StatusOK, dto.CheckAppVersionResponse{
			HasUpdate:     false,
			IsForceUpdate: false,
			LatestVersion: nil,
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

	latestVersionDto := dto.FromAppVersionModel(latestVersion)
	c.JSON(http.StatusOK, dto.CheckAppVersionResponse{
		HasUpdate:     hasUpdate,
		IsForceUpdate: isForceUpdate,
		LatestVersion: &latestVersionDto,
	})
}

type SemesterCalendarListResp struct {
	SemesterCode string `json:"semesterCode"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
}

// GetSemesterCalendars godoc
// @Summary      获取校历列表
// @Description  获取所有校历的列表，按学期代码倒序排列
// @Tags         config
// @Produce      json
// @Success      200  {array}   SemesterCalendarListResp
// @Failure      500  {object}  map[string]interface{}
// @Router       /config/semester-calendars [get]
func GetSemesterCalendars(c *gin.Context) {
	var calendars []model.SemesterCalendar
	if err := config.DB.Select("semester_code", "title", "subtitle").Order("semester_code desc").Find(&calendars).Error; err != nil {
		response.ResponseError(c, http.StatusInternalServerError, "获取校历列表失败: "+err.Error())
		return
	}

	resp := make([]SemesterCalendarListResp, 0, len(calendars))
	for _, cal := range calendars {
		resp = append(resp, SemesterCalendarListResp{
			SemesterCode: cal.SemesterCode,
			Title:        cal.Title,
			Subtitle:     cal.Subtitle,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GetSemesterCalendarDetail godoc
// @Summary      获取校历详情
// @Description  根据学期代码获取该学期的详细校历信息
// @Tags         config
// @Produce      json
// @Param        semester_code path     string  true  "学期代码(如: 2024-2025-1)"
// @Success      200           {object} model.SemesterCalendar
// @Failure      400           {object} map[string]interface{}
// @Failure      404           {object} map[string]interface{}
// @Router       /config/semester-calendars/{semester_code} [get]
func GetSemesterCalendarDetail(c *gin.Context) {
	semesterCode := c.Param("semester_code")
	if semesterCode == "" {
		response.ResponseError(c, http.StatusBadRequest, "学期代码不能为空")
		return
	}

	var calendar model.SemesterCalendar
	if err := config.DB.Where("semester_code = ?", semesterCode).First(&calendar).Error; err != nil {
		response.ResponseError(c, http.StatusNotFound, "未找到该校历信息: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, calendar)
}
