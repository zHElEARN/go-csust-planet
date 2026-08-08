package announcement

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title     string    `gorm:"type:varchar;not null;comment:公告标题"`
	Content   string    `gorm:"type:text;not null;comment:公告正文内容"`
	Platform  string    `gorm:"type:varchar;not null;index:idx_active_platform_created,priority:2;comment:发布平台(ios、android或all)"`
	IsActive  bool      `gorm:"type:boolean;not null;default:true;index:idx_active_platform_created,priority:1;comment:是否生效(控制公告上下线)"`
	IsBanner  bool      `gorm:"type:boolean;not null;default:false;comment:是否在App头部Banner处显示"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;index:idx_active_platform_created,priority:3,sort:desc"`
}

func (Entity) TableName() string { return "announcements" }

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformAll     = "all"
)
