package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/utils/sso"
)

func TestAuthServiceLoginCreatesUserAndReusesExistingUser(t *testing.T) {
	db := openServiceTestDB(t)

	type tokenCall struct {
		userID    uuid.UUID
		studentID string
		duration  time.Duration
	}

	var tokenCalls []tokenCall
	authService := NewAuthService(
		db,
		ProfileFetcherFunc(func(token string) (*sso.Profile, error) {
			if token == "first-token" || token == "second-token" {
				return &sso.Profile{UserAccount: "20240001"}, nil
			}
			return nil, errors.New("bad token")
		}),
		TokenGeneratorFunc(func(userID uuid.UUID, studentID string, duration time.Duration) (string, error) {
			tokenCalls = append(tokenCalls, tokenCall{userID: userID, studentID: studentID, duration: duration})
			return fmt.Sprintf("jwt-%s", studentID), nil
		}),
	)

	firstResp, err := authService.Login("first-token")
	if err != nil {
		t.Fatalf("expected first login to succeed: %v", err)
	}

	var users []struct {
		ID        uuid.UUID
		StudentID string
	}
	if err := db.Table("users").Find(&users).Error; err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected exactly one user after first login, got %d", len(users))
	}
	if firstResp.Token != "jwt-20240001" || firstResp.Profile == nil || firstResp.Profile.UserAccount != "20240001" {
		t.Fatalf("unexpected first login response: %+v", firstResp)
	}

	secondResp, err := authService.Login("second-token")
	if err != nil {
		t.Fatalf("expected second login to succeed: %v", err)
	}

	if err := db.Table("users").Find(&users).Error; err != nil {
		t.Fatalf("failed to query users after second login: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected existing user to be reused, got %d rows", len(users))
	}
	if len(tokenCalls) != 2 {
		t.Fatalf("expected token generator to be called twice, got %d", len(tokenCalls))
	}
	if tokenCalls[0].userID != users[0].ID || tokenCalls[1].userID != users[0].ID {
		t.Fatalf("expected both logins to reuse the same user id, calls=%+v user=%+v", tokenCalls, users[0])
	}
	if tokenCalls[0].duration != authTokenDuration || tokenCalls[1].duration != authTokenDuration {
		t.Fatalf("expected auth token duration %s, got %+v", authTokenDuration, tokenCalls)
	}
	if secondResp.Token != "jwt-20240001" {
		t.Fatalf("unexpected second login token: %+v", secondResp)
	}
}

func TestAuthServiceLoginReturnsUnauthorizedOnInvalidToken(t *testing.T) {
	db := openServiceTestDB(t)

	authService := NewAuthService(
		db,
		ProfileFetcherFunc(func(string) (*sso.Profile, error) {
			return nil, errors.New("expired")
		}),
		TokenGeneratorFunc(func(userID uuid.UUID, studentID string, duration time.Duration) (string, error) {
			return "", nil
		}),
	)

	_, err := authService.Login("bad-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
