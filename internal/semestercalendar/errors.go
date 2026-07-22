package semestercalendar

import "errors"

var (
	ErrNotFound = errors.New("semester calendar not found")
	ErrConflict = errors.New("semester calendar conflict")
)
