package announcement

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

type announcementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
}

type adminAnnouncementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"isActive"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
}

type upsertRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsActive *bool  `json:"isActive" binding:"required"`
	IsBanner *bool  `json:"isBanner" binding:"required"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// GetAnnouncements godoc
// @Summary 获取公告列表
// @Description 获取当前生效的公告列表，按创建时间倒序排列
// @Tags config
// @Produce json
// @Success 200 {array} announcementResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /config/announcements [get]
func (h *Handler) GetAnnouncements(c *gin.Context) {
	entities, err := h.service.ListActive()
	if err != nil {
		log.Printf("[ERROR] 获取公告失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取公告失败")
		return
	}
	result := make([]announcementResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toPublicResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// GetAdminAnnouncements godoc
// @Summary 获取后台公告列表
// @Description 获取全部公告列表，按创建时间倒序排列
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Success 200 {array} adminAnnouncementResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/announcements [get]
func (h *Handler) GetAdminAnnouncements(c *gin.Context) {
	entities, err := h.service.List()
	if err != nil {
		log.Printf("[ERROR] 获取后台公告列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取公告列表失败")
		return
	}
	result := make([]adminAnnouncementResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toAdminResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// GetAdminAnnouncement godoc
// @Summary 获取后台公告详情
// @Description 根据公告ID获取后台公告详情
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "公告ID"
// @Success 200 {object} adminAnnouncementResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/announcements/{id} [get]
func (h *Handler) GetAdminAnnouncement(c *gin.Context) {
	entity, ok := h.get(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// CreateAnnouncement godoc
// @Summary 创建公告
// @Description 创建一条新的后台公告
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param request body upsertRequest true "公告信息"
// @Success 201 {object} adminAnnouncementResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/announcements [post]
func (h *Handler) CreateAnnouncement(c *gin.Context) {
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Create(input)
	if err != nil {
		log.Printf("[ERROR] 创建公告失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "创建公告失败")
		return
	}
	c.JSON(http.StatusCreated, toAdminResponse(entity))
}

// UpdateAnnouncement godoc
// @Summary 更新公告
// @Description 根据公告ID更新后台公告
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "公告ID"
// @Param request body upsertRequest true "公告信息"
// @Success 200 {object} adminAnnouncementResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/announcements/{id} [put]
func (h *Handler) UpdateAnnouncement(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Update(id, input)
	if !h.handleError(c, err, "更新公告失败") {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// DeleteAnnouncement godoc
// @Summary 删除公告
// @Description 根据公告ID删除后台公告
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param id path string true "公告ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/announcements/{id} [delete]
func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if !h.handleError(c, h.service.Delete(id), "删除公告失败") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) get(c *gin.Context) (Entity, bool) {
	id, ok := parseID(c)
	if !ok {
		return Entity{}, false
	}
	entity, err := h.service.Get(id)
	if !h.handleError(c, err, "获取公告详情失败") {
		return Entity{}, false
	}
	return entity, true
}

func (h *Handler) handleError(c *gin.Context, err error, message string) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, ErrNotFound) {
		response.ResponseError(c, http.StatusNotFound, "未找到该公告")
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

func bindUpsert(c *gin.Context) (Upsert, bool) {
	var request upsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return Upsert{}, false
	}
	return Upsert{Title: request.Title, Content: request.Content, IsActive: *request.IsActive, IsBanner: *request.IsBanner}, true
}

func toPublicResponse(entity Entity) announcementResponse {
	return announcementResponse{ID: entity.ID.String(), Title: entity.Title, Content: entity.Content, IsBanner: entity.IsBanner, CreatedAt: entity.CreatedAt}
}
func toAdminResponse(entity Entity) adminAnnouncementResponse {
	return adminAnnouncementResponse{ID: entity.ID.String(), Title: entity.Title, Content: entity.Content, IsActive: entity.IsActive, IsBanner: entity.IsBanner, CreatedAt: entity.CreatedAt}
}
