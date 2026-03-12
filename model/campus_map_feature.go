package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CampusMapFeature struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Type string `gorm:"type:varchar(20);not null;default:'Feature';comment:GeoJSON要素类型"`

	Properties datatypes.JSON `gorm:"type:jsonb;not null;comment:业务属性(如名称、分类、校区)"`
	Geometry   datatypes.JSON `gorm:"type:jsonb;not null;comment:几何数据(Polygon及坐标点)"`
}
