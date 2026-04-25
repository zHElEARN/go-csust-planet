package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type AdminAppVersionResponse struct {
	ID            string    `json:"id"`
	Platform      string    `json:"platform"`
	VersionCode   int       `json:"versionCode"`
	VersionName   string    `json:"versionName"`
	IsForceUpdate bool      `json:"isForceUpdate"`
	ReleaseNotes  string    `json:"releaseNotes"`
	DownloadURL   string    `json:"downloadUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AdminAppVersionUpsertRequest struct {
	Platform      string `json:"platform" binding:"required,oneof=ios android"`
	VersionCode   *int   `json:"versionCode" binding:"required"`
	VersionName   string `json:"versionName" binding:"required"`
	IsForceUpdate *bool  `json:"isForceUpdate" binding:"required"`
	ReleaseNotes  string `json:"releaseNotes" binding:"required"`
	DownloadURL   string `json:"downloadUrl" binding:"required"`
}

func FromAdminAppVersionModel(v model.AppVersion) AdminAppVersionResponse {
	return AdminAppVersionResponse{
		ID:            v.ID.String(),
		Platform:      v.Platform,
		VersionCode:   v.VersionCode,
		VersionName:   v.VersionName,
		IsForceUpdate: v.IsForceUpdate,
		ReleaseNotes:  v.ReleaseNotes,
		DownloadURL:   v.DownloadURL,
		CreatedAt:     v.CreatedAt,
	}
}

func MapAdminAppVersions(versions []model.AppVersion) []AdminAppVersionResponse {
	res := make([]AdminAppVersionResponse, 0, len(versions))
	for _, v := range versions {
		res = append(res, FromAdminAppVersionModel(v))
	}
	return res
}
