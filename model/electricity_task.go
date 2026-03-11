package model

import (
	"time"

	"github.com/google/uuid"
)

type ElectricityTask struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DeviceTokenID uuid.UUID `gorm:"type:uuid;not null"`
	NotifyTime    time.Time `gorm:"type:timetz;not null;comment:每天的通知时间"`

	NextRunAt time.Time `gorm:"type:timestamptz;not null;index:idx_poll_queue,priority:1;comment:下次执行的精确时间戳"`
	Status    string    `gorm:"type:varchar(20);not null;default:'pending';index:idx_poll_queue,priority:2;comment:任务状态"`

	Campus   string `gorm:"type:varchar;not null;comment:校区"`
	Building string `gorm:"type:varchar;not null;comment:宿舍楼"`
	Room     string `gorm:"type:varchar;not null;comment:房间号"`
}
