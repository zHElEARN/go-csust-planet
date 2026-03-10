package model

import (
	"time"

	"github.com/google/uuid"
)

type ElectricityTask struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DeviceTokenID uuid.UUID `gorm:"type:uuid;not null"`
	NotifyTime    time.Time `gorm:"type:timetz;not null;comment:每天的通知时间"`
	Campus        string    `gorm:"type:varchar;not null;comment:校区"`
	Building      string    `gorm:"type:varchar;not null;comment:宿舍楼"`
	Room          string    `gorm:"type:varchar;not null;comment:房间号"`
}
