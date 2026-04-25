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

// GetAdminAppVersions godoc
// @Summary      获取后台 App 版本列表
// @Description  获取全部 App 版本列表，按平台升序、版本号降序排列
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Success      200            {array}   dto.AdminAppVersionResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/app-versions [get]
func (h *Handler) GetAdminAppVersions(c *gin.Context) {
	versions, err := h.adminAppVersionService.List()
	if err != nil {
		log.Printf("[ERROR] 获取后台版本列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取版本列表失败")
		return
	}

	c.JSON(http.StatusOK, dto.MapAdminAppVersions(versions))
}

// GetAdminAppVersion godoc
// @Summary      获取后台 App 版本详情
// @Description  根据版本ID获取后台 App 版本详情
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        id             path      string  true  "版本ID"
// @Success      200            {object}  dto.AdminAppVersionResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/app-versions/{id} [get]
func (h *Handler) GetAdminAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	version, err := h.adminAppVersionService.Get(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该版本")
			return
		}

		log.Printf("[ERROR] 获取后台版本详情失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取版本详情失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminAppVersionModel(version))
}

// CreateAppVersion godoc
// @Summary      创建 App 版本
// @Description  创建一条新的 App 版本记录
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                           true  "Bearer admin token"
// @Param        request        body      dto.AdminAppVersionUpsertRequest  true  "版本信息"
// @Success      201            {object}  dto.AdminAppVersionResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      409            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/app-versions [post]
func (h *Handler) CreateAppVersion(c *gin.Context) {
	var req dto.AdminAppVersionUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	version, err := h.adminAppVersionService.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrConflict) {
			response.ResponseError(c, http.StatusConflict, "该平台版本号已存在")
			return
		}

		log.Printf("[ERROR] 创建版本失败 platform=%s version_code=%d: %v", req.Platform, *req.VersionCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "创建版本失败")
		return
	}

	c.JSON(http.StatusCreated, dto.FromAdminAppVersionModel(version))
}

// UpdateAppVersion godoc
// @Summary      更新 App 版本
// @Description  根据版本ID更新 App 版本信息
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                           true  "Bearer admin token"
// @Param        id             path      string                           true  "版本ID"
// @Param        request        body      dto.AdminAppVersionUpsertRequest  true  "版本信息"
// @Success      200            {object}  dto.AdminAppVersionResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      409            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/app-versions/{id} [put]
func (h *Handler) UpdateAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AdminAppVersionUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	version, err := h.adminAppVersionService.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该版本")
			return
		}

		if errors.Is(err, service.ErrConflict) {
			response.ResponseError(c, http.StatusConflict, "该平台版本号已存在")
			return
		}

		log.Printf("[ERROR] 更新版本失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "更新版本失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminAppVersionModel(version))
}

// DeleteAppVersion godoc
// @Summary      删除 App 版本
// @Description  根据版本ID删除 App 版本
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        id             path      string  true  "版本ID"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /admin/app-versions/{id} [delete]
func (h *Handler) DeleteAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.adminAppVersionService.Delete(id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该版本")
			return
		}

		log.Printf("[ERROR] 查询待删除版本失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除版本失败")
		return
	}

	c.Status(http.StatusNoContent)
}
