package controller

import (
	"errors"
	"hash/crc32"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
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
func GetAdminAppVersions(c *gin.Context) {
	var versions []model.AppVersion
	if err := config.DB.Order("platform asc, version_code desc").Find(&versions).Error; err != nil {
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
func GetAdminAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var version model.AppVersion
	if err := config.DB.First(&version, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
func CreateAppVersion(c *gin.Context) {
	var req dto.AdminAppVersionUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var version model.AppVersion
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := lockAppVersion(tx, req.Platform, *req.VersionCode); err != nil {
			return err
		}

		exists, err := appVersionExists(tx, req.Platform, *req.VersionCode, nil)
		if err != nil {
			return err
		}
		if exists {
			return gorm.ErrDuplicatedKey
		}

		version = model.AppVersion{
			Platform:      req.Platform,
			VersionCode:   *req.VersionCode,
			VersionName:   req.VersionName,
			IsForceUpdate: *req.IsForceUpdate,
			ReleaseNotes:  req.ReleaseNotes,
			DownloadURL:   req.DownloadURL,
		}

		return tx.Create(&version).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isDuplicateKeyError(err) {
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
func UpdateAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AdminAppVersionUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var version model.AppVersion
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&version, "id = ?", id).Error; err != nil {
			return err
		}

		if err := lockAppVersion(tx, req.Platform, *req.VersionCode); err != nil {
			return err
		}

		exists, err := appVersionExists(tx, req.Platform, *req.VersionCode, &id)
		if err != nil {
			return err
		}
		if exists {
			return gorm.ErrDuplicatedKey
		}

		version.Platform = req.Platform
		version.VersionCode = *req.VersionCode
		version.VersionName = req.VersionName
		version.IsForceUpdate = *req.IsForceUpdate
		version.ReleaseNotes = req.ReleaseNotes
		version.DownloadURL = req.DownloadURL

		return tx.Save(&version).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该版本")
			return
		}

		if errors.Is(err, gorm.ErrDuplicatedKey) || isDuplicateKeyError(err) {
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
func DeleteAppVersion(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var version model.AppVersion
	if err := config.DB.First(&version, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该版本")
			return
		}

		log.Printf("[ERROR] 查询待删除版本失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除版本失败")
		return
	}

	if err := config.DB.Delete(&version).Error; err != nil {
		log.Printf("[ERROR] 删除版本失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除版本失败")
		return
	}

	c.Status(http.StatusNoContent)
}

func appVersionExists(tx *gorm.DB, platform string, versionCode int, excludeID *uuid.UUID) (bool, error) {
	query := tx.Model(&model.AppVersion{}).Where("platform = ? AND version_code = ?", platform, versionCode)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func lockAppVersion(tx *gorm.DB, platform string, versionCode int) error {
	lockKeyPlatform := int32(crc32.ChecksumIEEE([]byte(platform)))
	lockKeyVersion := int32(versionCode)

	return tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", lockKeyPlatform, lockKeyVersion).Error
}
