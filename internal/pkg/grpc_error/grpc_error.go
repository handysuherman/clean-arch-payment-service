package grpcError

import (
	"context"
	"database/sql"
	"errors"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNoCtxMetaData = errors.New("No ctx metadata")
)

func ErrorResponse(err error) error {
	return status.Error(GetErrorStatusCode(err), err.Error())
}

func GetErrorStatusCode(err error) codes.Code {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return codes.NotFound
	case errors.Is(err, context.Canceled):
		return codes.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	case errors.Is(err, ErrNoCtxMetaData):
		return codes.Unauthenticated
	case CheckErrMessage(err, constants.Validate):
		return codes.InvalidArgument
	case CheckErrMessage(err, constants.Redis):
		return codes.NotFound
	case CheckErrMessage(err, constants.FieldValidation):
		return codes.InvalidArgument
	case CheckErrMessage(err, constants.RequiredHeaders):
		return codes.Unauthenticated
	case CheckErrMessage(err, constants.Base64):
		return codes.InvalidArgument
	case CheckErrMessage(err, constants.Unmarshal):
		return codes.InvalidArgument
	case CheckErrMessage(err, constants.Uuid):
		return codes.InvalidArgument
	case CheckErrMessage(err, constants.Cookie):
		return codes.Unauthenticated
	case CheckErrMessage(err, constants.Token):
		return codes.Unauthenticated
	case CheckErrMessage(err, constants.Bcrypt):
		return codes.InvalidArgument
	case CheckErrMessage(err, pgx.ErrNoRows.Error()):
		return codes.NotFound
	case CheckErrMessage(err, unierror.ErrUnsupportedPaymentType.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrUnsupportedPaymentChannel.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrInvalidAmount.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrInvalidSuccessURL.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrInvalidFailureURL.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrExpiryLessThan3Days.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrInvalidCustomerPhoneNumberInput.Error()):
		return codes.InvalidArgument
	case CheckErrMessage(err, unierror.ErrReferenceIDShouldNotBeEmpty.Error()):
		return codes.InvalidArgument
	}
	return codes.Internal
}

func CheckErrMessage(err error, msg string) bool {
	return helper.CheckErrMessages(err, msg)
}
