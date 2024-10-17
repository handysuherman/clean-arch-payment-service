package unierror

import "errors"

var (
	ErrNoRateLimitType   = errors.New("error code: WK-29501")
	ErrNoLevelRegistered = errors.New("error code: WK-93193")
	ErrRateLimitExceeded = errors.New("error code: WK-429")

	// Err Source Data
	ErrUnableToRetrieveDataFromRedisOrDB  = errors.New("error code: WK-30001")
	ErrUnableToUpdateDataToDB             = errors.New("error code: WK-30002")
	ErrNotEquivalentUIDAndTokenClaimerUID = errors.New("error code: WK-30003")

	// parse error
	ErrTimeParsingYYYYMMDD = errors.New("error code: WK-11001")
	ErrInvalidULID         = errors.New("error code: WK-11002")
)
