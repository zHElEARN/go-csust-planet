package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

var errSemesterCalendarConflict = errors.New("semester calendar conflict")

// GetAdminSemesterCalendars godoc
// @Summary      获取后台校历列表
// @Description  获取全部校历列表，按学期代码倒序排列
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Success      200            {array}   dto.AdminSemesterCalendarResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/semester-calendars [get]
func (h *Handler) GetAdminSemesterCalendars(c *gin.Context) {
	var calendars []model.SemesterCalendar
	if err := h.db.Order("semester_code desc").Find(&calendars).Error; err != nil {
		log.Printf("[ERROR] 获取后台校历列表失败: %v", err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历列表失败")
		return
	}

	c.JSON(http.StatusOK, dto.MapAdminSemesterCalendars(calendars))
}

// GetAdminSemesterCalendar godoc
// @Summary      获取后台校历详情
// @Description  根据学期代码获取后台校历详情
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        semester_code  path      string  true  "学期代码"
// @Success      200            {object}  dto.AdminSemesterCalendarResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/semester-calendars/{semester_code} [get]
func (h *Handler) GetAdminSemesterCalendar(c *gin.Context) {
	semesterCode := c.Param("semester_code")

	var calendar model.SemesterCalendar
	if err := h.db.Where("semester_code = ?", semesterCode).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该校历")
			return
		}

		log.Printf("[ERROR] 获取后台校历详情失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "获取校历详情失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminSemesterCalendarModel(calendar))
}

// CreateSemesterCalendar godoc
// @Summary      创建校历
// @Description  创建一条新的后台校历
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                                 true  "Bearer admin token"
// @Param        request        body      dto.AdminSemesterCalendarUpsertRequest  true  "校历信息"
// @Success      201            {object}  dto.AdminSemesterCalendarResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      409            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/semester-calendars [post]
func (h *Handler) CreateSemesterCalendar(c *gin.Context) {
	var req dto.AdminSemesterCalendarUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	calendar := model.SemesterCalendar{
		SemesterCode:     req.SemesterCode,
		Title:            req.Title,
		Subtitle:         req.Subtitle,
		CalendarStart:    *req.CalendarStart,
		CalendarEnd:      *req.CalendarEnd,
		SemesterStart:    *req.SemesterStart,
		SemesterEnd:      *req.SemesterEnd,
		Notes:            normalizeCalendarNotes(req.Notes),
		CustomWeekRanges: normalizeCustomWeekRanges(req.CustomWeekRanges),
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&calendar).Error; err != nil {
			if isDuplicateKeyError(err) {
				return errSemesterCalendarConflict
			}

			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, errSemesterCalendarConflict) {
			response.ResponseError(c, http.StatusConflict, "该学期代码已存在")
			return
		}

		log.Printf("[ERROR] 创建校历失败 semester_code=%s: %v", req.SemesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "创建校历失败")
		return
	}

	c.JSON(http.StatusCreated, dto.FromAdminSemesterCalendarModel(calendar))
}

// UpdateSemesterCalendar godoc
// @Summary      更新校历
// @Description  根据学期代码更新后台校历
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                                 true  "Bearer admin token"
// @Param        semester_code  path      string                                 true  "学期代码"
// @Param        request        body      dto.AdminSemesterCalendarUpsertRequest  true  "校历信息"
// @Success      200            {object}  dto.AdminSemesterCalendarResponse
// @Failure      400            {object}  dto.ErrorResponse
// @Failure      401            {object}  dto.ErrorResponse
// @Failure      404            {object}  dto.ErrorResponse
// @Failure      409            {object}  dto.ErrorResponse
// @Failure      500            {object}  dto.ErrorResponse
// @Router       /admin/semester-calendars/{semester_code} [put]
func (h *Handler) UpdateSemesterCalendar(c *gin.Context) {
	semesterCode := c.Param("semester_code")

	var req dto.AdminSemesterCalendarUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	var calendar model.SemesterCalendar
	if err := h.db.Where("semester_code = ?", semesterCode).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该校历")
			return
		}

		log.Printf("[ERROR] 查询待更新校历失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "更新校历失败")
		return
	}

	calendar.SemesterCode = req.SemesterCode
	calendar.Title = req.Title
	calendar.Subtitle = req.Subtitle
	calendar.CalendarStart = *req.CalendarStart
	calendar.CalendarEnd = *req.CalendarEnd
	calendar.SemesterStart = *req.SemesterStart
	calendar.SemesterEnd = *req.SemesterEnd
	calendar.Notes = normalizeCalendarNotes(req.Notes)
	calendar.CustomWeekRanges = normalizeCustomWeekRanges(req.CustomWeekRanges)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&calendar).Error; err != nil {
			if isDuplicateKeyError(err) {
				return errSemesterCalendarConflict
			}

			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, errSemesterCalendarConflict) {
			response.ResponseError(c, http.StatusConflict, "该学期代码已存在")
			return
		}

		log.Printf("[ERROR] 更新校历失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "更新校历失败")
		return
	}

	c.JSON(http.StatusOK, dto.FromAdminSemesterCalendarModel(calendar))
}

// DeleteSemesterCalendar godoc
// @Summary      删除校历
// @Description  根据学期代码删除后台校历
// @Tags         admin
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer admin token"
// @Param        semester_code  path      string  true  "学期代码"
// @Success      204
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /admin/semester-calendars/{semester_code} [delete]
func (h *Handler) DeleteSemesterCalendar(c *gin.Context) {
	semesterCode := c.Param("semester_code")

	var calendar model.SemesterCalendar
	if err := h.db.Where("semester_code = ?", semesterCode).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ResponseError(c, http.StatusNotFound, "未找到该校历")
			return
		}

		log.Printf("[ERROR] 查询待删除校历失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除校历失败")
		return
	}

	if err := h.db.Delete(&calendar).Error; err != nil {
		log.Printf("[ERROR] 删除校历失败 semester_code=%s: %v", semesterCode, err)
		response.ResponseError(c, http.StatusInternalServerError, "删除校历失败")
		return
	}

	c.Status(http.StatusNoContent)
}

func normalizeCalendarNotes(notes []model.CalendarNote) []model.CalendarNote {
	if notes == nil {
		return []model.CalendarNote{}
	}

	return notes
}

func normalizeCustomWeekRanges(ranges []model.CustomWeekRange) []model.CustomWeekRange {
	if ranges == nil {
		return []model.CustomWeekRange{}
	}

	return ranges
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
