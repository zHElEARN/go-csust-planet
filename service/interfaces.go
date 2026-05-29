package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
	"github.com/zHElEARN/go-csust-planet/utils/apns"
	"github.com/zHElEARN/go-csust-planet/utils/sso"
)

type AuthService interface {
	Login(token string) (dto.LoginResponse, error)
}

type ElectricityTaskService interface {
	Sync(ctx context.Context, userID uuid.UUID, req dto.SyncElectricityTaskRequest) error
}

type AdminAppVersionService interface {
	List() ([]model.AppVersion, error)
	ListByPlatform(platform string) ([]model.AppVersion, error)
	Get(id uuid.UUID) (model.AppVersion, error)
	CheckUpdate(platform string, currentVersionCode int) (AppVersionCheckResult, error)
	Create(req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Update(id uuid.UUID, req dto.AdminAppVersionUpsertRequest) (model.AppVersion, error)
	Delete(id uuid.UUID) error
}

type AppVersionCheckResult struct {
	HasUpdate     bool
	IsForceUpdate bool
	LatestVersion *model.AppVersion
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

type ElectricityRoomValidator interface {
	ValidateRoom(ctx context.Context, campusName, buildingName, roomName string) error
}

type ElectricityFetcher interface {
	GetElectricity(ctx context.Context, campusName, buildingName, roomName string) (float64, error)
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

type ElectricityRoomValidatorFunc func(ctx context.Context, campusName, buildingName, roomName string) error

func (f ElectricityRoomValidatorFunc) ValidateRoom(ctx context.Context, campusName, buildingName, roomName string) error {
	return f(ctx, campusName, buildingName, roomName)
}

type ElectricityFetcherFunc func(ctx context.Context, campusName, buildingName, roomName string) (float64, error)

func (f ElectricityFetcherFunc) GetElectricity(ctx context.Context, campusName, buildingName, roomName string) (float64, error) {
	return f(ctx, campusName, buildingName, roomName)
}

type NotificationSenderFunc func(notification apns.PushNotification) error

func (f NotificationSenderFunc) SendPushNotification(notification apns.PushNotification) error {
	return f(notification)
}
