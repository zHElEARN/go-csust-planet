package model

import (
	"time"

	"github.com/google/uuid"
)

type AppVersion struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Platform      string    `gorm:"type:varchar;not null;comment:平台(ios或android)"`
	VersionCode   int       `gorm:"type:integer;not null;comment:内部版本号(用于逻辑比对)"`
	VersionName   string    `gorm:"type:varchar;not null;comment:展示版本号(例如1.5.1)"`
	IsForceUpdate bool      `gorm:"type:boolean;not null;default:false;comment:是否强制更新"`
	ReleaseNotes  string    `gorm:"type:text;not null;comment:更新日志"`
	DownloadURL   string    `gorm:"type:varchar;not null;comment:下载地址"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}
