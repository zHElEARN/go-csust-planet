package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type AdminAnnouncementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"isActive"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
}

type AdminAnnouncementUpsertRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsActive *bool  `json:"isActive" binding:"required"`
	IsBanner *bool  `json:"isBanner" binding:"required"`
}

func FromAdminAnnouncementModel(a model.Announcement) AdminAnnouncementResponse {
	return AdminAnnouncementResponse{
		ID:        a.ID.String(),
		Title:     a.Title,
		Content:   a.Content,
		IsActive:  a.IsActive,
		IsBanner:  a.IsBanner,
		CreatedAt: a.CreatedAt,
	}
}

func MapAdminAnnouncements(announcements []model.Announcement) []AdminAnnouncementResponse {
	res := make([]AdminAnnouncementResponse, 0, len(announcements))
	for _, a := range announcements {
		res = append(res, FromAdminAnnouncementModel(a))
	}
	return res
}
