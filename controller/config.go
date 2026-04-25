package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /config/announcements [get]
func (h *Handler) GetAnnouncements(c *gin.Context) {
	var announcements []model.Announcement
	if err := h.db.Where("is_active = ?", true).Order("created_at desc").Find(&announcements).Error; err != nil {
		log.Printf("[ERROR] 获取公告失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取公告失败")
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
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /config/campus-map [get]
func (h *Handler) GetCampusMap(c *gin.Context) {
	var features []model.CampusMapFeature
	if err := h.db.Find(&features).Error; err != nil {
		log.Printf("[ERROR] 获取校园地图数据失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校园地图数据失败")
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
// @Failure      400       {object}  dto.ErrorResponse
// @Failure      500       {object}  dto.ErrorResponse
// @Router       /config/app-versions [get]
func (h *Handler) GetAppVersions(c *gin.Context) {
	var req dto.AppVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var versions []model.AppVersion
	if err := h.db.Where("platform = ?", req.Platform).Order("version_code desc").Find(&versions).Error; err != nil {
		log.Printf("[ERROR] 获取版本信息失败 platform=%s: %v", req.Platform, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取版本信息失败")
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
// @Failure      400                   {object}  dto.ErrorResponse
// @Failure      500                   {object}  dto.ErrorResponse
// @Router       /config/app-versions/check [get]
func (h *Handler) CheckAppVersion(c *gin.Context) {
	var req dto.CheckAppVersionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var latestVersion model.AppVersion
	err := h.db.Where("platform = ?", req.Platform).Order("version_code desc").First(&latestVersion).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] 检查版本更新失败 platform=%s current_version_code=%d: %v", req.Platform, req.CurrentVersionCode, err)
			response.ResponseError(c, http.StatusInternalServerError, "检查版本更新失败")
			return
		}

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
		err := h.db.Select("id").Where("platform = ? AND version_code > ? AND is_force_update = ?", req.Platform, req.CurrentVersionCode, true).First(&forceUpdate).Error
		if err == nil {
			isForceUpdate = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] 检查强制更新失败 platform=%s current_version_code=%d: %v", req.Platform, req.CurrentVersionCode, err)
			response.ResponseError(c, http.StatusInternalServerError, "检查版本更新失败")
			return
		}
	}

	latestVersionDto := dto.FromAppVersionModel(latestVersion)
	c.JSON(http.StatusOK, dto.CheckAppVersionResponse{
		HasUpdate:     hasUpdate,
		IsForceUpdate: isForceUpdate,
		LatestVersion: &latestVersionDto,
	})
}

// GetSemesterCalendars godoc
// @Summary      获取校历列表
// @Description  获取所有校历的列表，按学期代码倒序排列
// @Tags         config
// @Produce      json
// @Success      200  {array}   dto.SemesterCalendarListResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /config/semester-calendars [get]
func (h *Handler) GetSemesterCalendars(c *gin.Context) {
	var calendars []model.SemesterCalendar
	if err := h.db.Select("semester_code", "title", "subtitle").Order("semester_code desc").Find(&calendars).Error; err != nil {
		log.Printf("[ERROR] 获取校历列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历列表失败")
		return
	}

	res := dto.MapSemesterCalendarList(calendars)
	c.JSON(http.StatusOK, res)
}

// GetSemesterCalendarDetail godoc
// @Summary      获取校历详情
// @Description  根据学期代码获取该学期的详细校历信息
// @Tags         config
// @Produce      json
// @Param        semester_code path     string  true  "学期代码(如: 2024-2025-1)"
// @Success      200           {object} dto.SemesterCalendarDetailResponse
// @Failure      400           {object} dto.ErrorResponse
// @Failure      404           {object} dto.ErrorResponse
// @Router       /config/semester-calendars/{semester_code} [get]
func (h *Handler) GetSemesterCalendarDetail(c *gin.Context) {
	semesterCode := c.Param("semester_code")
	if semesterCode == "" {
		response.ResponseError(c, http.StatusBadRequest, "学期代码不能为空")
		return
	}

	var calendar model.SemesterCalendar
	if err := h.db.Where("semester_code = ?", semesterCode).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该校历信息")
			return
		}

		log.Printf("[ERROR] 获取校历详情失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历详情失败")
		return
	}

	res := dto.FromSemesterCalendarDetailModel(calendar)
	c.JSON(http.StatusOK, res)
}
