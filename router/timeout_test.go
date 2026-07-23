package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/internal/announcement"
	"github.com/zHElEARN/go-csust-planet/utils/response"
)

type canceledAnnouncementRepository struct {
	ctx context.Context
}

func (r *canceledAnnouncementRepository) List(context.Context) ([]announcement.Entity, error) {
	return nil, nil
}
func (r *canceledAnnouncementRepository) ListActive(ctx context.Context) ([]announcement.Entity, error) {
	r.ctx = ctx
	return nil, ctx.Err()
}
func (r *canceledAnnouncementRepository) Get(context.Context, uuid.UUID) (announcement.Entity, error) {
	return announcement.Entity{}, nil
}
func (r *canceledAnnouncementRepository) Create(context.Context, announcement.Entity) (announcement.Entity, error) {
	return announcement.Entity{}, nil
}
func (r *canceledAnnouncementRepository) Update(context.Context, uuid.UUID, announcement.Entity) (announcement.Entity, error) {
	return announcement.Entity{}, nil
}
func (r *canceledAnnouncementRepository) Delete(context.Context, uuid.UUID) error { return nil }

func TestRequestTimeoutReturnsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestTimeout(10 * time.Millisecond))
	r.GET("/slow", func(c *gin.Context) {
		<-c.Request.Context().Done()
		response.HandleContextError(c, c.Request.Context().Err())
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, resp.Code)
	}
	if resp.Body.String() != `{"error":"请求处理超时"}` {
		t.Fatalf("unexpected timeout response: %q", resp.Body.String())
	}
	if value := resp.Header().Get("Retry-After"); value != "" {
		t.Fatalf("expected no Retry-After header, got %q", value)
	}
}

func TestRequestTimeoutPreservesSuccessfulResponseAndSetsDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestTimeout(time.Second))
	r.GET("/fast", func(c *gin.Context) {
		if _, ok := c.Request.Context().Deadline(); !ok {
			t.Error("expected request context to contain a deadline")
		}
		c.JSON(http.StatusAccepted, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted || resp.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected successful response: status=%d body=%q", resp.Code, resp.Body.String())
	}
}

func TestRequestTimeoutDiscardsLateWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestTimeout(5 * time.Millisecond))
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(20 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "late"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, resp.Code)
	}
	if resp.Body.String() != `{"error":"请求处理超时"}` {
		t.Fatalf("expected only the timeout response, got %q", resp.Body.String())
	}
}

func TestCanceledRequestDoesNotWriteResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &canceledAnnouncementRepository{}
	handler := announcement.NewHandler(announcement.NewService(repository))
	r := gin.New()
	r.Use(requestTimeout(time.Second))
	r.GET("/canceled", handler.GetAnnouncements)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/canceled", nil).WithContext(ctx)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Body.Len() != 0 || resp.Header().Get("Content-Type") != "" {
		t.Fatalf("expected canceled request not to write a response, got status=%d body=%q", resp.Code, resp.Body.String())
	}
	if repository.ctx == nil || !errors.Is(repository.ctx.Err(), context.Canceled) {
		t.Fatalf("expected repository to receive canceled request context, got %v", repository.ctx)
	}
}
