package model

import (
	"time"

	"github.com/google/uuid"
)

type Announcement struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title     string    `gorm:"type:varchar;not null;comment:公告标题"`
	Content   string    `gorm:"type:text;not null;comment:公告正文内容"`
	IsActive  bool      `gorm:"type:boolean;not null;default:true;index:idx_active_created,priority:1;comment:是否生效(控制公告上下线)"`
	IsBanner  bool      `gorm:"type:boolean;not null;default:false;comment:是否在App头部Banner处显示"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;index:idx_active_created,priority:2"`
}
