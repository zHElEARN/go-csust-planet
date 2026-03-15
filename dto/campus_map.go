package dto

import "github.com/zHElEARN/go-csust-planet/model"

type CampusMapFeatureResponse struct {
	Type       string                  `json:"type"`
	Properties model.FeatureProperties `json:"properties"`
	Geometry   model.FeatureGeometry   `json:"geometry"`
}

type CampusMapResponse struct {
	Type     string                     `json:"type"`
	Features []CampusMapFeatureResponse `json:"features"`
}

func FromCampusMapFeatureModel(f model.CampusMapFeature) CampusMapFeatureResponse {
	return CampusMapFeatureResponse{
		Type:       f.Type,
		Properties: f.Properties,
		Geometry:   f.Geometry,
	}
}

func MapCampusMapFeatures(features []model.CampusMapFeature) CampusMapResponse {
	res := make([]CampusMapFeatureResponse, 0, len(features))
	for _, f := range features {
		res = append(res, FromCampusMapFeatureModel(f))
	}
	return CampusMapResponse{
		Type:     "FeatureCollection",
		Features: res,
	}
}
