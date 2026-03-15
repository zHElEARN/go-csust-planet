package model

import (
	"time"

	"github.com/google/uuid"
)

// CalendarNote 对应校历的备注事项
type CalendarNote struct {
	Row        int    `json:"row"`
	Content    string `json:"content"`
	NeedNumber bool   `json:"needNumber,omitempty"`
}

// CustomWeekRange 对应自定义假期/周次范围
type CustomWeekRange struct {
	StartRow int    `json:"startRow"`
	EndRow   int    `json:"endRow"`
	Content  string `json:"content"`
}

// SemesterCalendar 学期校历主模型
type SemesterCalendar struct {
	ID               uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"-"`
	SemesterCode     string            `gorm:"type:varchar;not null;uniqueIndex;comment:学期代码(如: 2024-2025-1)" json:"semesterCode"`
	Title            string            `gorm:"type:varchar;not null;comment:校历标题(如: 2024-2025学年度校历)" json:"title"`
	Subtitle         string            `gorm:"type:varchar;not null;comment:校历副标题(如: 第一学期)" json:"subtitle"`
	CalendarStart    time.Time         `gorm:"type:date;not null;comment:校历开始日期" json:"calendarStart"`
	CalendarEnd      time.Time         `gorm:"type:date;not null;comment:校历结束日期" json:"calendarEnd"`
	SemesterStart    time.Time         `gorm:"type:date;not null;comment:学期开学日期" json:"semesterStart"`
	SemesterEnd      time.Time         `gorm:"type:date;not null;comment:学期结束日期" json:"semesterEnd"`
	Notes            []CalendarNote    `gorm:"type:jsonb;serializer:json;default:'[]';comment:校历底部备注(JSON数组)" json:"notes"`
	CustomWeekRanges []CustomWeekRange `gorm:"type:jsonb;serializer:json;default:'[]';comment:自定义周次与假期范围(JSON数组)" json:"customWeekRanges"`
	CreatedAt        time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"-"`
}
