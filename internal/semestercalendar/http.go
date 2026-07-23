package semestercalendar

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type Handler struct{ service *Service }

type listResponse struct {
	SemesterCode string `json:"semesterCode"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
}
type detailResponse struct {
	SemesterCode     string            `json:"semesterCode"`
	Title            string            `json:"title"`
	Subtitle         string            `json:"subtitle"`
	CalendarStart    time.Time         `json:"calendarStart"`
	CalendarEnd      time.Time         `json:"calendarEnd"`
	SemesterStart    time.Time         `json:"semesterStart"`
	SemesterEnd      time.Time         `json:"semesterEnd"`
	Notes            []CalendarNote    `json:"notes"`
	CustomWeekRanges []CustomWeekRange `json:"customWeekRanges"`
}
type adminResponse struct {
	SemesterCode     string            `json:"semesterCode"`
	Title            string            `json:"title"`
	Subtitle         string            `json:"subtitle"`
	CalendarStart    time.Time         `json:"calendarStart"`
	CalendarEnd      time.Time         `json:"calendarEnd"`
	SemesterStart    time.Time         `json:"semesterStart"`
	SemesterEnd      time.Time         `json:"semesterEnd"`
	Notes            []CalendarNote    `json:"notes"`
	CustomWeekRanges []CustomWeekRange `json:"customWeekRanges"`
	CreatedAt        time.Time         `json:"createdAt"`
}
type upsertRequest struct {
	SemesterCode     string            `json:"semesterCode" binding:"required"`
	Title            string            `json:"title" binding:"required"`
	Subtitle         string            `json:"subtitle" binding:"required"`
	CalendarStart    *time.Time        `json:"calendarStart" binding:"required"`
	CalendarEnd      *time.Time        `json:"calendarEnd" binding:"required"`
	SemesterStart    *time.Time        `json:"semesterStart" binding:"required"`
	SemesterEnd      *time.Time        `json:"semesterEnd" binding:"required"`
	Notes            []CalendarNote    `json:"notes"`
	CustomWeekRanges []CustomWeekRange `json:"customWeekRanges"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// GetSemesterCalendars godoc
// @Summary 获取校历列表
// @Description 获取所有校历的列表，按学期代码倒序排列
// @Tags config
// @Produce json
// @Success 200 {array} listResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /config/semester-calendars [get]
func (h *Handler) GetSemesterCalendars(c *gin.Context) {
	entities, err := h.service.ListSummaries(c.Request.Context())
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 获取校历列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历列表失败")
		return
	}
	result := make([]listResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toListResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// GetSemesterCalendarDetail godoc
// @Summary 获取校历详情
// @Description 根据学期代码获取该学期的详细校历信息
// @Tags config
// @Produce json
// @Param semester_code path string true "学期代码(如: 2024-2025-1)"
// @Success 200 {object} detailResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /config/semester-calendars/{semester_code} [get]
func (h *Handler) GetSemesterCalendarDetail(c *gin.Context) {
	code := c.Param("semester_code")
	if code == "" {
		response.ResponseError(c, http.StatusBadRequest, "学期代码不能为空")
		return
	}
	entity, err := h.service.Get(c.Request.Context(), code)
	if !h.handleError(c, err, "获取校历详情失败", "未找到该校历信息") {
		return
	}
	c.JSON(http.StatusOK, toDetailResponse(entity))
}

// GetAdminSemesterCalendars godoc
// @Summary 获取后台校历列表
// @Description 获取全部校历列表，按学期代码倒序排列
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Success 200 {array} adminResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/semester-calendars [get]
func (h *Handler) GetAdminSemesterCalendars(c *gin.Context) {
	entities, err := h.service.List(c.Request.Context())
	if err != nil {
		if response.HandleContextError(c, err) {
			return
		}
		log.Printf("[ERROR] 获取后台校历列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历列表失败")
		return
	}
	result := make([]adminResponse, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toAdminResponse(entity))
	}
	c.JSON(http.StatusOK, result)
}

// GetAdminSemesterCalendar godoc
// @Summary 获取后台校历详情
// @Description 根据学期代码获取后台校历详情
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param semester_code path string true "学期代码"
// @Success 200 {object} adminResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/semester-calendars/{semester_code} [get]
func (h *Handler) GetAdminSemesterCalendar(c *gin.Context) {
	entity, ok := h.get(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// CreateSemesterCalendar godoc
// @Summary 创建校历
// @Description 创建一条新的后台校历
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param request body upsertRequest true "校历信息"
// @Success 201 {object} adminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/semester-calendars [post]
func (h *Handler) CreateSemesterCalendar(c *gin.Context) {
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Create(c.Request.Context(), input)
	if !h.handleError(c, err, "创建校历失败", "未找到该校历") {
		return
	}
	c.JSON(http.StatusCreated, toAdminResponse(entity))
}

// UpdateSemesterCalendar godoc
// @Summary 更新校历
// @Description 根据学期代码更新后台校历
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param semester_code path string true "学期代码"
// @Param request body upsertRequest true "校历信息"
// @Success 200 {object} adminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/semester-calendars/{semester_code} [put]
func (h *Handler) UpdateSemesterCalendar(c *gin.Context) {
	code := c.Param("semester_code")
	input, ok := bindUpsert(c)
	if !ok {
		return
	}
	entity, err := h.service.Update(c.Request.Context(), code, input)
	if !h.handleError(c, err, "更新校历失败", "未找到该校历") {
		return
	}
	c.JSON(http.StatusOK, toAdminResponse(entity))
}

// DeleteSemesterCalendar godoc
// @Summary 删除校历
// @Description 根据学期代码删除后台校历
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer admin token"
// @Param semester_code path string true "学期代码"
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/semester-calendars/{semester_code} [delete]
func (h *Handler) DeleteSemesterCalendar(c *gin.Context) {
	code := c.Param("semester_code")
	if !h.handleError(c, h.service.Delete(c.Request.Context(), code), "删除校历失败", "未找到该校历") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) get(c *gin.Context) (Entity, bool) {
	entity, err := h.service.Get(c.Request.Context(), c.Param("semester_code"))
	if !h.handleError(c, err, "获取校历详情失败", "未找到该校历") {
		return Entity{}, false
	}
	return entity, true
}

func bindUpsert(c *gin.Context) (Upsert, bool) {
	var request upsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return Upsert{}, false
	}
	return Upsert{SemesterCode: request.SemesterCode, Title: request.Title, Subtitle: request.Subtitle, CalendarStart: *request.CalendarStart, CalendarEnd: *request.CalendarEnd, SemesterStart: *request.SemesterStart, SemesterEnd: *request.SemesterEnd, Notes: request.Notes, CustomWeekRanges: request.CustomWeekRanges}, true
}

func (h *Handler) handleError(c *gin.Context, err error, message string, notFoundMessage string) bool {
	if err == nil {
		return true
	}
	if response.HandleContextError(c, err) {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		response.ResponseError(c, http.StatusNotFound, notFoundMessage)
		return false
	}
	if errors.Is(err, ErrConflict) {
		response.ResponseError(c, http.StatusConflict, "该学期代码已存在")
		return false
	}
	log.Printf("[ERROR] %s: %v", message, err)
	response.ResponseError(c, http.StatusInternalServerError, message)
	return false
}

func toListResponse(entity Entity) listResponse {
	return listResponse{SemesterCode: entity.SemesterCode, Title: entity.Title, Subtitle: entity.Subtitle}
}
func toDetailResponse(entity Entity) detailResponse {
	return detailResponse{SemesterCode: entity.SemesterCode, Title: entity.Title, Subtitle: entity.Subtitle, CalendarStart: entity.CalendarStart, CalendarEnd: entity.CalendarEnd, SemesterStart: entity.SemesterStart, SemesterEnd: entity.SemesterEnd, Notes: entity.Notes, CustomWeekRanges: entity.CustomWeekRanges}
}
func toAdminResponse(entity Entity) adminResponse {
	return adminResponse{SemesterCode: entity.SemesterCode, Title: entity.Title, Subtitle: entity.Subtitle, CalendarStart: entity.CalendarStart, CalendarEnd: entity.CalendarEnd, SemesterStart: entity.SemesterStart, SemesterEnd: entity.SemesterEnd, Notes: entity.Notes, CustomWeekRanges: entity.CustomWeekRanges, CreatedAt: entity.CreatedAt}
}
