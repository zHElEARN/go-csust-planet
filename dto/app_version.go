package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type AppVersionsRequest struct {
	Platform string `form:"platform" binding:"required,oneof=ios android"`
}

type CheckAppVersionRequest struct {
	Platform           string `form:"platform" binding:"required,oneof=ios android"`
	CurrentVersionCode int    `form:"currentVersionCode" binding:"required"`
}

type AppVersionResponse struct {
	Platform      string    `json:"platform"`
	VersionCode   int       `json:"versionCode"`
	VersionName   string    `json:"versionName"`
	IsForceUpdate bool      `json:"isForceUpdate"`
	ReleaseNotes  string    `json:"releaseNotes"`
	DownloadURL   string    `json:"downloadUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CheckAppVersionResponse struct {
	HasUpdate     bool                `json:"hasUpdate"`
	IsForceUpdate bool                `json:"isForceUpdate"`
	LatestVersion *AppVersionResponse `json:"latestVersion"`
}

func FromAppVersionModel(v model.AppVersion) AppVersionResponse {
	return AppVersionResponse{
		Platform:      v.Platform,
		VersionCode:   v.VersionCode,
		VersionName:   v.VersionName,
		IsForceUpdate: v.IsForceUpdate,
		ReleaseNotes:  v.ReleaseNotes,
		DownloadURL:   v.DownloadURL,
		CreatedAt:     v.CreatedAt,
	}
}

func MapAppVersions(versions []model.AppVersion) []AppVersionResponse {
	res := make([]AppVersionResponse, 0, len(versions))
	for _, v := range versions {
		res = append(res, FromAppVersionModel(v))
	}
	return res
}
