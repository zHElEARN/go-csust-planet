package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type SemesterCalendarListResponse struct {
	SemesterCode string `json:"semesterCode"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
}

type SemesterCalendarDetailResponse struct {
	SemesterCode     string                  `json:"semesterCode"`
	Title            string                  `json:"title"`
	Subtitle         string                  `json:"subtitle"`
	CalendarStart    time.Time               `json:"calendarStart"`
	CalendarEnd      time.Time               `json:"calendarEnd"`
	SemesterStart    time.Time               `json:"semesterStart"`
	SemesterEnd      time.Time               `json:"semesterEnd"`
	Notes            []model.CalendarNote    `json:"notes"`
	CustomWeekRanges []model.CustomWeekRange `json:"customWeekRanges"`
}

func FromSemesterCalendarListModel(c model.SemesterCalendar) SemesterCalendarListResponse {
	return SemesterCalendarListResponse{
		SemesterCode: c.SemesterCode,
		Title:        c.Title,
		Subtitle:     c.Subtitle,
	}
}

func MapSemesterCalendarList(calendars []model.SemesterCalendar) []SemesterCalendarListResponse {
	res := make([]SemesterCalendarListResponse, 0, len(calendars))
	for _, c := range calendars {
		res = append(res, FromSemesterCalendarListModel(c))
	}
	return res
}

func FromSemesterCalendarDetailModel(c model.SemesterCalendar) SemesterCalendarDetailResponse {
	return SemesterCalendarDetailResponse{
		SemesterCode:     c.SemesterCode,
		Title:            c.Title,
		Subtitle:         c.Subtitle,
		CalendarStart:    c.CalendarStart,
		CalendarEnd:      c.CalendarEnd,
		SemesterStart:    c.SemesterStart,
		SemesterEnd:      c.SemesterEnd,
		Notes:            c.Notes,
		CustomWeekRanges: c.CustomWeekRanges,
	}
}
