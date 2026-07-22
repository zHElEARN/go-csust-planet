package semestercalendar

import "time"

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

func NewService(repository Repository) *Service     { return &Service{repository: repository} }
func (s *Service) List() ([]Entity, error)          { return s.repository.List() }
func (s *Service) ListSummaries() ([]Entity, error) { return s.repository.ListSummaries() }
func (s *Service) Get(code string) (Entity, error)  { return s.repository.Get(code) }

func (s *Service) Create(input Upsert) (Entity, error) { return s.repository.Create(fromUpsert(input)) }

func (s *Service) Update(code string, input Upsert) (Entity, error) {
	entity, err := s.repository.Get(code)
	if err != nil {
		return Entity{}, err
	}
	entity.SemesterCode, entity.Title, entity.Subtitle = input.SemesterCode, input.Title, input.Subtitle
	entity.CalendarStart, entity.CalendarEnd = input.CalendarStart, input.CalendarEnd
	entity.SemesterStart, entity.SemesterEnd = input.SemesterStart, input.SemesterEnd
	entity.Notes, entity.CustomWeekRanges = normalizeNotes(input.Notes), normalizeRanges(input.CustomWeekRanges)
	return s.repository.Update(entity)
}

func (s *Service) Delete(code string) error { return s.repository.Delete(code) }

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
