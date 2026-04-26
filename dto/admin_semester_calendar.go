package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type AdminSemesterCalendarResponse struct {
	SemesterCode     string                  `json:"semesterCode"`
	Title            string                  `json:"title"`
	Subtitle         string                  `json:"subtitle"`
	CalendarStart    time.Time               `json:"calendarStart"`
	CalendarEnd      time.Time               `json:"calendarEnd"`
	SemesterStart    time.Time               `json:"semesterStart"`
	SemesterEnd      time.Time               `json:"semesterEnd"`
	Notes            []model.CalendarNote    `json:"notes"`
	CustomWeekRanges []model.CustomWeekRange `json:"customWeekRanges"`
	CreatedAt        time.Time               `json:"createdAt"`
}

type AdminSemesterCalendarUpsertRequest struct {
	SemesterCode     string                  `json:"semesterCode" binding:"required"`
	Title            string                  `json:"title" binding:"required"`
	Subtitle         string                  `json:"subtitle" binding:"required"`
	CalendarStart    *time.Time              `json:"calendarStart" binding:"required"`
	CalendarEnd      *time.Time              `json:"calendarEnd" binding:"required"`
	SemesterStart    *time.Time              `json:"semesterStart" binding:"required"`
	SemesterEnd      *time.Time              `json:"semesterEnd" binding:"required"`
	Notes            []model.CalendarNote    `json:"notes"`
	CustomWeekRanges []model.CustomWeekRange `json:"customWeekRanges"`
}

func FromAdminSemesterCalendarModel(c model.SemesterCalendar) AdminSemesterCalendarResponse {
	return AdminSemesterCalendarResponse{
		SemesterCode:     c.SemesterCode,
		Title:            c.Title,
		Subtitle:         c.Subtitle,
		CalendarStart:    c.CalendarStart,
		CalendarEnd:      c.CalendarEnd,
		SemesterStart:    c.SemesterStart,
		SemesterEnd:      c.SemesterEnd,
		Notes:            c.Notes,
		CustomWeekRanges: c.CustomWeekRanges,
		CreatedAt:        c.CreatedAt,
	}
}

func MapAdminSemesterCalendars(calendars []model.SemesterCalendar) []AdminSemesterCalendarResponse {
	res := make([]AdminSemesterCalendarResponse, 0, len(calendars))
	for _, calendar := range calendars {
		res = append(res, FromAdminSemesterCalendarModel(calendar))
	}
	return res
}
