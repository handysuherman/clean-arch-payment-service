package unierror

import "errors"

var (
	ErrAuthorizationHeaderIsNotProvided = errors.New("error code: WK-29701")
	ErrInvalidAuthorizationHeaderFormat = errors.New("error code: WK-29702")
	ErrUnsupportedAuthorizationType     = errors.New("error code: WK-29703")
	ErrInvalidTokenType                 = errors.New("error code: WK-29704")
	ErrTokenUnmarshal                   = errors.New("error code: WK-29705")
	ErrTokenParsePublicKey              = errors.New("error code: WK-29706")
	ErrUnableGetTokenFromRedis          = errors.New("error code: WK-29707")
	ErrNotEquivalentUUIDString          = errors.New("error code: WK-29708")
	ErrTokenExpired                     = errors.New("error code: WK-29709")
	ErrUnableToVerifyToken              = errors.New("error code: WK-29710")
	ErrUnableToCreateToken              = errors.New("error code: WK-29711")
	ErrTokenNotBelongToPlatform         = errors.New("error code: WK-29712")
	ErrTokenIsBlocked                   = errors.New("error code: WK-29713")
	ErrTokenAccessDenied                = errors.New("error code: WK-29714")
)
