package model

import (
	"github.com/google/uuid"
)

type FeatureProperties struct {
	Name     string `json:"name"`
	Campus   string `json:"campus"`
	Category string `json:"category"`
}

type FeatureGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type CampusMapFeature struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Type string `gorm:"type:varchar(20);not null;default:'Feature';comment:GeoJSON要素类型"`

	Properties FeatureProperties `gorm:"type:jsonb;serializer:json;not null;comment:业务属性(如名称、分类、校区)"`
	Geometry   FeatureGeometry   `gorm:"type:jsonb;serializer:json;not null;comment:几何数据(Polygon及坐标点)"`
}
