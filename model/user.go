package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StudentID string    `gorm:"type:varchar;uniqueIndex;not null;comment:学号"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

	DeviceTokens []DeviceToken `gorm:"foreignKey:UserID"`
}
