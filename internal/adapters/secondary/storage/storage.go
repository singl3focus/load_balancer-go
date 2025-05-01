package storage

import "errors"

var (
	ErrEmptyAction       = errors.New("action did not affect the data")
	ErrDataNotFound      = errors.New("data not found")
	ErrEmptyData         = errors.New("empty data")
	ErrInvalidDataFormat = errors.New("data is stored in an invalid format")
)