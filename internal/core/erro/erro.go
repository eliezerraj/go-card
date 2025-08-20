package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrUpdate			= errors.New("update unsuccessful")
	ErrUpdateRows		= errors.New("update affect 0 rows")
	ErrTimeout			= errors.New("timeout: context deadline exceeded.")
	ErrHTTPForbiden		= errors.New("forbiden request")
	ErrUnauthorized 	= errors.New("not authorized")
	ErrServer		 	= errors.New("server identified error")
)