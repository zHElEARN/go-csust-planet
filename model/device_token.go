package model

import (
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Token     string    `gorm:"type:varchar;uniqueIndex;not null;column:device_token;comment:设备推送Token"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

	ElectricityTasks []ElectricityTask `gorm:"foreignKey:DeviceTokenID"`
}
