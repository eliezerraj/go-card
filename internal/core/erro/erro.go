package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrUpdate			= errors.New("update unsuccessful")
	ErrUpdateRows		= errors.New("update affect 0 rows")
)