package controller

import (
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/service"
)

type Dependencies struct {
	DB                     *gorm.DB
	AdminAppVersionService service.AdminAppVersionService
}

type Handler struct {
	db                     *gorm.DB
	adminAppVersionService service.AdminAppVersionService
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		db:                     deps.DB,
		adminAppVersionService: deps.AdminAppVersionService,
	}
}
