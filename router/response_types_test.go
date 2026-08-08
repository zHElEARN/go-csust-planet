package router

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/internal/campusmap"
	"github.com/zHElEARN/go-csust-planet/internal/semestercalendar"
)

// These types decode integration-test responses without exporting production HTTP DTOs.
type AnnouncementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
}

type AdminAnnouncementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Platform  string    `json:"platform"`
	IsActive  bool      `json:"isActive"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
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

type AdminAppVersionResponse struct {
	ID string `json:"id"`
	AppVersionResponse
}

type CheckAppVersionResponse struct {
	HasUpdate     bool                `json:"hasUpdate"`
	IsForceUpdate bool                `json:"isForceUpdate"`
	LatestVersion *AppVersionResponse `json:"latestVersion"`
}

type SemesterCalendarListResponse struct {
	SemesterCode string `json:"semesterCode"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
}
type SemesterCalendarDetailResponse struct {
	SemesterCode     string                             `json:"semesterCode"`
	Title            string                             `json:"title"`
	Subtitle         string                             `json:"subtitle"`
	CalendarStart    time.Time                          `json:"calendarStart"`
	CalendarEnd      time.Time                          `json:"calendarEnd"`
	SemesterStart    time.Time                          `json:"semesterStart"`
	SemesterEnd      time.Time                          `json:"semesterEnd"`
	Notes            []semestercalendar.CalendarNote    `json:"notes"`
	CustomWeekRanges []semestercalendar.CustomWeekRange `json:"customWeekRanges"`
}
type AdminSemesterCalendarResponse struct {
	SemesterCalendarDetailResponse
	CreatedAt time.Time `json:"createdAt"`
}
type CampusMapFeatureResponse struct {
	Type       string                      `json:"type"`
	Properties campusmap.FeatureProperties `json:"properties"`
	Geometry   campusmap.FeatureGeometry   `json:"geometry"`
}
type CampusMapResponse struct {
	Type     string                     `json:"type"`
	Features []CampusMapFeatureResponse `json:"features"`
}
