package unierror

import "errors"

var (
	ErrProductVariationShouldNotBeEmpty              = errors.New("error code: WK-100000")
	ErrProductVariationShouldNotExceedingLimit       = errors.New("error code: WK-100001")
	ErrProductVariationOptionShouldNotBeEmpty        = errors.New("error code: WK-100002")
	ErrProductVariationOptionShouldNotExceedingLimit = errors.New("error code: WK-100003")
	ErrProductVariationOptionPricesShouldNotBeEmpty  = errors.New("error code: WK-100004")

	ErrDuplicateTypeAgeBandPrice    = errors.New("error code: WK-100005")
	ErrDuplicateTypeVariationOption = errors.New("error code: WK-100006")
	ErrDuplicateTypeVariation       = errors.New("error code: WK-100007")
)
