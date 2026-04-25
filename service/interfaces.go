package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/campuscard"
	"github.com/zHElEARN/go-csust-planet/utils/sso"
)

type AuthService interface {
	Login(token string) (dto.LoginResponse, error)
}

type ElectricityTaskService interface {
	Sync(userID uuid.UUID, req dto.SyncElectricityTaskRequest) error
}

type AdminAppVersionService interface {
	List() ([]model.AppVersion, error)
	Get(id uuid.UUID) (model.AppVersion, error)
	Create(req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Update(id uuid.UUID, req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Delete(id uuid.UUID) error
}

type ElectricityPushService interface {
	PollAndProcess(now time.Time) error
}

type ProfileFetcher interface {
	GetUserProfile(token string) (*sso.Profile, error)
}

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, studentID string, duration time.Duration) (string, error)
}

type BuildingResolver interface {
	GetBuildingByCampusName(campusName, buildingName string) (campuscard.Building, error)
}

type ElectricityFetcher interface {
	GetElectricity(building campuscard.Building, roomNum string) (float64, error)
}

type NotificationSender interface {
	SendPushNotification(notification apns.PushNotification) error
}

type ProfileFetcherFunc func(token string) (*sso.Profile, error)

func (f ProfileFetcherFunc) GetUserProfile(token string) (*sso.Profile, error) {
	return f(token)
}

type TokenGeneratorFunc func(userID uuid.UUID, studentID string, duration time.Duration) (string, error)

func (f TokenGeneratorFunc) GenerateToken(userID uuid.UUID, studentID string, duration time.Duration) (string, error) {
	return f(userID, studentID, duration)
}

type BuildingResolverFunc func(campusName, buildingName string) (campuscard.Building, error)

func (f BuildingResolverFunc) GetBuildingByCampusName(campusName, buildingName string) (campuscard.Building, error) {
	return f(campusName, buildingName)
}

type ElectricityFetcherFunc func(building campuscard.Building, roomNum string) (float64, error)

func (f ElectricityFetcherFunc) GetElectricity(building campuscard.Building, roomNum string) (float64, error) {
	return f(building, roomNum)
}

type NotificationSenderFunc func(notification apns.PushNotification) error

func (f NotificationSenderFunc) SendPushNotification(notification apns.PushNotification) error {
	return f(notification)
}
