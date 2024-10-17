package unierror

import "errors"

var (
	ErrUnsupportedPaymentType          = errors.New("unsupported payment type, error code: WK-700001")
	ErrUnsupportedPaymentChannel       = errors.New("unsupported payment channel, error code: WK-700002")
	ErrInvalidAmount                   = errors.New("to start a payment transaction, min amount was 100, error code: WK-700003")
	ErrInvalidSuccessURL               = errors.New("success return url is not a valid URL, error code: WK-700004")
	ErrInvalidFailureURL               = errors.New("failure return url is not a valid URL, error code: WK-700005")
	ErrExpiryLessThan3Days             = errors.New("please set an expiry payment or at least set it more than 72 hours or ( 3 days ), error code: WK-700006")
	ErrInvalidCustomerPhoneNumberInput = errors.New("invalid customer phone number input should be only digits, error code: WK-700007")
	ErrReferenceIDShouldNotBeEmpty     = errors.New("reference id should not be empty, error code: WK-700008")
)
