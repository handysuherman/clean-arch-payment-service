package unierror

import "errors"

var (
	ErrNoPlatformKey           = errors.New("error code: WK-29801")
	ErrMismatchPlatformKey     = errors.New("error code: WK-29802")
	ErrMismatchAuthPlatformKey = errors.New("error code: WK-29803")
)
