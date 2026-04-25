package controller

import (
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/service"
)

type Dependencies struct {
	DB                     *gorm.DB
	AuthService            service.AuthService
	ElectricityTaskService service.ElectricityTaskService
	AdminAppVersionService service.AdminAppVersionService
}

type Handler struct {
	db                     *gorm.DB
	authService            service.AuthService
	electricityTaskService service.ElectricityTaskService
	adminAppVersionService service.AdminAppVersionService
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		db:                     deps.DB,
		authService:            deps.AuthService,
		electricityTaskService: deps.ElectricityTaskService,
		adminAppVersionService: deps.AdminAppVersionService,
	}
}
