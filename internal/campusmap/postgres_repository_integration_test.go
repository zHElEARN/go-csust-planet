package campusmap

import (
	"context"
	"errors"
	"testing"

	"github.com/zHElEARN/go-csust-planet/testsupport"
)

func TestPostgresRepositoryHonorsCanceledContext(t *testing.T) {
	db, _ := testsupport.OpenTestDB(t, true)
	repository := NewPostgresRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repository.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
