package semestercalendar

import (
	"time"

	"github.com/google/uuid"
)

type CalendarNote struct {
	Row        int    `json:"row"`
	Content    string `json:"content"`
	NeedNumber bool   `json:"needNumber,omitempty"`
}

type CustomWeekRange struct {
	StartRow int    `json:"startRow"`
	EndRow   int    `json:"endRow"`
	Content  string `json:"content"`
}

type Entity struct {
	ID               uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SemesterCode     string            `gorm:"type:varchar;not null;uniqueIndex;comment:学期代码(如: 2024-2025-1)"`
	Title            string            `gorm:"type:varchar;not null;comment:校历标题(如: 2024-2025学年度校历)"`
	Subtitle         string            `gorm:"type:varchar;not null;comment:校历副标题(如: 第一学期)"`
	CalendarStart    time.Time         `gorm:"type:date;not null;comment:校历开始日期"`
	CalendarEnd      time.Time         `gorm:"type:date;not null;comment:校历结束日期"`
	SemesterStart    time.Time         `gorm:"type:date;not null;comment:学期开学日期"`
	SemesterEnd      time.Time         `gorm:"type:date;not null;comment:学期结束日期"`
	Notes            []CalendarNote    `gorm:"type:jsonb;not null;serializer:json;default:'[]';comment:校历底部备注(JSON数组)"`
	CustomWeekRanges []CustomWeekRange `gorm:"type:jsonb;not null;serializer:json;default:'[]';comment:自定义周次与假期范围(JSON数组)"`
	CreatedAt        time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

func (Entity) TableName() string { return "semester_calendars" }
