package appversion

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type Handler struct{ service *Service }

type listRequest struct {
	Platform string `form:"platform" binding:"required,oneof=ios android"`
}
type checkRequest struct {
	Platform           string `form:"platform" binding:"required,oneof=ios android"`
	CurrentVersionCode int    `form:"currentVersionCode" binding:"required"`
}
type upsertRequest struct {
	Platform      string `json:"platform" binding:"required,oneof=ios android"`
	VersionCode   *int   `json:"versionCode" binding:"required"`
	VersionName   string `json:"versionName" binding:"required"`
	IsForceUpdate *bool  `json:"isForceUpdate" binding:"required"`
	ReleaseNotes  string `json:"releaseNotes" binding:"required"`
	DownloadURL   string `json:"downloadUrl" binding:"required"`
}
type publicResponse struct {
	Platform      string    `json:"platform"`
	VersionCode   int       `json:"versionCode"`
	VersionName   string    `json:"versionName"`
	IsForceUpdate bool      `json:"isForceUpdate"`
	ReleaseNotes  string    `json:"releaseNotes"`
	DownloadURL   string    `json:"downloadUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}
type adminResponse struct {
	ID            string    `json:"id"`
	Platform      string    `json:"platform"`
	VersionCode   int       `json:"versionCode"`
	VersionName   string    `json:"versionName"`
	IsForceUpdate bool      `json:"isForceUpdate"`
	ReleaseNotes  string    `json:"releaseNotes"`
	DownloadURL   string    `json:"downloadUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}
type checkResponse struct {
	HasUpdate     bool            `json:"hasUpdate"`
	IsForceUpdate bool            `json:"isForceUpdate"`
	LatestVersion *publicResponse `json:"latestVersion"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// GetAppVersions godoc
// @Summary 获取App所有版本
// @Description 获取指定平台的所有App版本历史
// @Tags config
// @Produce json
// @Param platform query string true "平台(ios或android)" Enums(ios, android)
// @Success 200 {array} publicResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /config/app-versions [get]
func (h *Handler) GetAppVersions(c *gin.Context) {
	var request listRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	entities, err := h.service.ListByPlatform(c.Request.Context(), request.Platform)
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 获取版本信息失败 platform=%s: %v", request.Platform, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取版本信息失败")
		return
	}
	result := make([]publicResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toPublicResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// CheckAppVersion godoc
// @Summary 检查App版本更新
// @Description 检查指定平台的App是否有更新
// @Tags config
// @Produce json
// @Param platform query string true "平台(ios或android)" Enums(ios, android)
// @Param currentVersionCode query int true "当前版本号"
// @Success 200 {object} checkResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /config/app-versions/check [get]
func (h *Handler) CheckAppVersion(c *gin.Context) {
	var request checkRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	result, err := h.service.CheckUpdate(c.Request.Context(), request.Platform, request.CurrentVersionCode)
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 检查版本更新失败 platform=%s current_version_code=%d: %v", request.Platform, request.CurrentVersionCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "检查版本更新失败")
		return
	}
	if result.LatestVersion == nil {
		c.JSON(http.StatusOK, checkResponse{})
		return
	}
	latest := toPublicResponse(*result.LatestVersion)
	c.JSON(http.StatusOK, checkResponse{HasUpdate: result.HasUpdate, IsForceUpdate: result.IsForceUpdate, LatestVersion: &latest})
}

// GetAdminAppVersions godoc
// @Summary 获取后台 App 版本列表
// @Description 获取全部 App 版本列表，按平台升序、版本号降序排列
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Success 200 {array} adminResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /admin/app-versions [get]
func (h *Handler) GetAdminAppVersions(c *gin.Context) {
	entities, err := h.service.List(c.Request.Context())
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 获取后台版本列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取版本列表失败")
		return
	}
	result := make([]adminResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toAdminResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// GetAdminAppVersion godoc
// @Summary 获取后台 App 版本详情
// @Description 根据版本ID获取后台 App 版本详情
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "版本ID"
// @Success 200 {object} adminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /admin/app-versions/{id} [get]
func (h *Handler) GetAdminAppVersion(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	entity, err := h.service.Get(c.Request.Context(), id)
	if !h.handleError(c, err, "获取版本详情失败") {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// CreateAppVersion godoc
// @Summary 创建 App 版本
// @Description 创建一条新的 App 版本记录
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param request body upsertRequest true "版本信息"
// @Success 201 {object} adminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /admin/app-versions [post]
func (h *Handler) CreateAppVersion(c *gin.Context) {
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Create(c.Request.Context(), input)
	if !h.handleError(c, err, "创建版本失败") {
		return
	}
	c.JSON(http.StatusCreated, toAdminResponse(entity))
}

// UpdateAppVersion godoc
// @Summary 更新 App 版本
// @Description 根据版本ID更新 App 版本信息
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "版本ID"
// @Param request body upsertRequest true "版本信息"
// @Success 200 {object} adminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /admin/app-versions/{id} [put]
func (h *Handler) UpdateAppVersion(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Update(c.Request.Context(), id, input)
	if !h.handleError(c, err, "更新版本失败") {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// DeleteAppVersion godoc
// @Summary 删除 App 版本
// @Description 根据版本ID删除 App 版本
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "版本ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 504 {object} response.ErrorResponse
// @Router /admin/app-versions/{id} [delete]
func (h *Handler) DeleteAppVersion(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if !h.handleError(c, h.service.Delete(c.Request.Context(), id), "删除版本失败") {
		return
	}
	c.Status(http.StatusNoContent)
}

func bindUpsert(c *gin.Context) (Upsert, bool) {
	var request upsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return Upsert{}, false
	}
	return Upsert{Platform: request.Platform, VersionCode: *request.VersionCode, VersionName: request.VersionName, IsForceUpdate: *request.IsForceUpdate, ReleaseNotes: request.ReleaseNotes, DownloadURL: request.DownloadURL}, true
}

func (h *Handler) handleError(c *gin.Context, err error, message string) bool {
	if err == nil {
		return true
	}
	if response.HandleContextError(c, err) {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		response.ResponseError(c, http.StatusNotFound, "未找到该版本")
		return false
	}
	if errors.Is(err, ErrConflict) {
		response.ResponseError(c, http.StatusConflict, "该平台版本号已存在")
		return false
	}
	log.Printf("[ERROR] %s: %v", message, err)
	response.ResponseError(c, http.StatusInternalServerError, message)
	return false
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的资源ID")
		return uuid.Nil, false
	}
	return id, true
}
func toPublicResponse(entity Entity) publicResponse {
	return publicResponse{Platform: entity.Platform, VersionCode: entity.VersionCode, VersionName: entity.VersionName, IsForceUpdate: entity.IsForceUpdate, ReleaseNotes: entity.ReleaseNotes, DownloadURL: entity.DownloadURL, CreatedAt: entity.CreatedAt}
}
func toAdminResponse(entity Entity) adminResponse {
	return adminResponse{ID: entity.ID.String(), Platform: entity.Platform, VersionCode: entity.VersionCode, VersionName: entity.VersionName, IsForceUpdate: entity.IsForceUpdate, ReleaseNotes: entity.ReleaseNotes, DownloadURL: entity.DownloadURL, CreatedAt: entity.CreatedAt}
}
