package semestercalendar

import (
	"context"
	"time"
)

type Service struct{ repository Repository }

type Upsert struct {
	SemesterCode     string
	Title            string
	Subtitle         string
	CalendarStart    time.Time
	CalendarEnd      time.Time
	SemesterStart    time.Time
	SemesterEnd      time.Time
	Notes            []CalendarNote
	CustomWeekRanges []CustomWeekRange
}

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) List(ctx context.Context) ([]Entity, error) {
	return s.repository.List(ctx)
}
func (s *Service) ListSummaries(ctx context.Context) ([]Entity, error) {
	return s.repository.ListSummaries(ctx)
}
func (s *Service) Get(ctx context.Context, code string) (Entity, error) {
	return s.repository.Get(ctx, code)
}

func (s *Service) Create(ctx context.Context, input Upsert) (Entity, error) {
	return s.repository.Create(ctx, fromUpsert(input))
}

func (s *Service) Update(ctx context.Context, code string, input Upsert) (Entity, error) {
	return s.repository.Update(ctx, code, fromUpsert(input))
}

func (s *Service) Delete(ctx context.Context, code string) error {
	return s.repository.Delete(ctx, code)
}

func fromUpsert(input Upsert) Entity {
	return Entity{SemesterCode: input.SemesterCode, Title: input.Title, Subtitle: input.Subtitle, CalendarStart: input.CalendarStart, CalendarEnd: input.CalendarEnd, SemesterStart: input.SemesterStart, SemesterEnd: input.SemesterEnd, Notes: normalizeNotes(input.Notes), CustomWeekRanges: normalizeRanges(input.CustomWeekRanges)}
}

func normalizeNotes(value []CalendarNote) []CalendarNote {
	if value == nil {
		return []CalendarNote{}
	}
	return value
}
func normalizeRanges(value []CustomWeekRange) []CustomWeekRange {
	if value == nil {
		return []CustomWeekRange{}
	}
	return value
}
