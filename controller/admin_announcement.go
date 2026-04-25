package controller

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

// GetAdminAnnouncements godoc
// @Summary      获取后台公告列表
// @Description  获取全部公告列表，按创建时间倒序排列
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Success      200            {array}   dto.AdminAnnouncementResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/announcements [get]
func (h *Handler) GetAdminAnnouncements(c *gin.Context) {
	var announcements []model.Announcement
	if err := h.db.Order("created_at desc").Find(&announcements).Error; err != nil {
		log.Printf("[ERROR] 获取后台公告列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取公告列表失败")
		return
	}

	c.JSON(http.StatusOK, dto.MapAdminAnnouncements(announcements))
}

// GetAdminAnnouncement godoc
// @Summary      获取后台公告详情
// @Description  根据公告ID获取后台公告详情
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        id             path      string  true  "公告ID"
// @Success      200            {object}  dto.AdminAnnouncementResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/announcements/{id} [get]
func (h *Handler) GetAdminAnnouncement(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var announcement model.Announcement
	if err := h.db.First(&announcement, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该公告")
			return
		}

		log.Printf("[ERROR] 获取后台公告详情失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取公告详情失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminAnnouncementModel(announcement))
}

// CreateAnnouncement godoc
// @Summary      创建公告
// @Description  创建一条新的后台公告
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                             true  "Bearer admin token"
// @Param        request        body      dto.AdminAnnouncementUpsertRequest  true  "公告信息"
// @Success      201            {object}  dto.AdminAnnouncementResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/announcements [post]
func (h *Handler) CreateAnnouncement(c *gin.Context) {
	var req dto.AdminAnnouncementUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	announcement := model.Announcement{
		ID:        uuid.New(),
		Title:     req.Title,
		Content:   req.Content,
		IsActive:  *req.IsActive,
		IsBanner:  *req.IsBanner,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.db.Model(&model.Announcement{}).Create(map[string]any{
		"id":         announcement.ID,
		"title":      announcement.Title,
		"content":    announcement.Content,
		"is_active":  announcement.IsActive,
		"is_banner":  announcement.IsBanner,
		"created_at": announcement.CreatedAt,
	}).Error; err != nil {
		log.Printf("[ERROR] 创建公告失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "创建公告失败")
		return
	}
	if err := h.db.First(&announcement, "id = ?", announcement.ID).Error; err != nil {
		log.Printf("[ERROR] 查询新建公告失败 id=%s: %v", announcement.ID, err)
		response.ResponseError(c, http.StatusInternalServerError, "创建公告失败")
		return
	}

	c.JSON(http.StatusCreated, dto.FromAdminAnnouncementModel(announcement))
}

// UpdateAnnouncement godoc
// @Summary      更新公告
// @Description  根据公告ID更新后台公告
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                             true  "Bearer admin token"
// @Param        id             path      string                             true  "公告ID"
// @Param        request        body      dto.AdminAnnouncementUpsertRequest  true  "公告信息"
// @Success      200            {object}  dto.AdminAnnouncementResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/announcements/{id} [put]
func (h *Handler) UpdateAnnouncement(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AdminAnnouncementUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var announcement model.Announcement
	if err := h.db.First(&announcement, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该公告")
			return
		}

		log.Printf("[ERROR] 查询待更新公告失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "更新公告失败")
		return
	}

	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.IsActive = *req.IsActive
	announcement.IsBanner = *req.IsBanner

	if err := h.db.Save(&announcement).Error; err != nil {
		log.Printf("[ERROR] 更新公告失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "更新公告失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminAnnouncementModel(announcement))
}

// DeleteAnnouncement godoc
// @Summary      删除公告
// @Description  根据公告ID删除后台公告
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        id             path      string  true  "公告ID"
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /admin/announcements/{id} [delete]
func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var announcement model.Announcement
	if err := h.db.First(&announcement, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该公告")
			return
		}

		log.Printf("[ERROR] 查询待删除公告失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除公告失败")
		return
	}

	if err := h.db.Delete(&announcement).Error; err != nil {
		log.Printf("[ERROR] 删除公告失败 id=%s: %v", id, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除公告失败")
		return
	}

	c.Status(http.StatusNoContent)
}
