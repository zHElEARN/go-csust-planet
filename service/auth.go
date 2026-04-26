package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

const authTokenDuration = 30 * 24 * time.Hour

type authService struct {
	db             *gorm.DB
	profileFetcher ProfileFetcher
	tokenGenerator TokenGenerator
}

func NewAuthService(db *gorm.DB, profileFetcher ProfileFetcher, tokenGenerator TokenGenerator) AuthService {
	return &authService{
		db:             db,
		profileFetcher: profileFetcher,
		tokenGenerator: tokenGenerator,
	}
}

func (s *authService) Login(token string) (dto.LoginResponse, error) {
	profile, err := s.profileFetcher.GetUserProfile(token)
	if err != nil {
		return dto.LoginResponse{}, ErrUnauthorized
	}

	var user model.User
	result := s.db.Where("student_id = ?", profile.UserAccount).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = model.User{
				StudentID: profile.UserAccount,
			}
			if err := s.db.Create(&user).Error; err != nil {
				return dto.LoginResponse{}, fmt.Errorf("%w: %v", ErrUserCreateFailed, err)
			}
		} else {
			return dto.LoginResponse{}, fmt.Errorf("%w: %v", ErrUserQueryFailed, result.Error)
		}
	}

	jwtToken, err := s.tokenGenerator.GenerateToken(user.ID, profile.UserAccount, authTokenDuration)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: %v", ErrTokenGenerateFailed, err)
	}

	return dto.LoginResponse{
		Token:   jwtToken,
		Profile: profile,
	}, nil
}
