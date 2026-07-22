package campusmap

import "context"

type Repository interface {
	List(context.Context) ([]Entity, error)
}
