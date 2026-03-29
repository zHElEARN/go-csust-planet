package dto

import (
	"time"

	"github.com/zHElEARN/go-csust-planet/model"
)

type AnnouncementResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsBanner  bool      `json:"isBanner"`
	CreatedAt time.Time `json:"createdAt"`
}

func FromAnnouncementModel(a model.Announcement) AnnouncementResponse {
	return AnnouncementResponse{
		ID:        a.ID.String(),
		Title:     a.Title,
		Content:   a.Content,
		IsBanner:  a.IsBanner,
		CreatedAt: a.CreatedAt,
	}
}

func MapAnnouncements(announcements []model.Announcement) []AnnouncementResponse {
	res := make([]AnnouncementResponse, 0, len(announcements))
	for _, a := range announcements {
		res = append(res, FromAnnouncementModel(a))
	}
	return res
}
