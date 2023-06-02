// Package query -
package query

import (
	in "go-ddd-architecture/app/usecase/port/in/query"
)

// usecase -
type usecase struct {
}

// New -
func New() in.QueryUsecase {
	return &usecase{}
}
