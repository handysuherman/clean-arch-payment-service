package httpError

import (
	"context"
	"database/sql"
	"encoding/json"

	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror"
	"github.com/pkg/errors"

	"github.com/go-playground/validator/v10"
)

const (
	ErrBadRequest          = "Bad request"
	ErrForbidden           = "Forbidden"
	ErrNotFound            = "Not Found"
	ErrUnauthorized        = "Unauthorized"
	ErrRequestTimeout      = "Request Timeout"
	ErrInvalidEmail        = "Invalid email"
	ErrInvalidPassword     = "Invalid password"
	ErrInvalidField        = "Invalid field"
	ErrInternalServerError = "Internal Server Error"
	ErrTooManyRequests     = "Too Many Requests"
)

var (
	BadRequest          = errors.New("Bad Request")
	WrongCredentials    = errors.New("Wrong Credentials")
	NotFound            = errors.New("Not Found")
	Unauthorized        = errors.New("Unauthorized")
	Forbidden           = errors.New("Forbidden")
	InternalServerError = errors.New("Internal Server Error")
	TooManyRequests     = errors.New("Too Many Requests")
)

type RestErr interface {
	Status() int
	Error() string
	Causes() interface{}
	ErrBody() RestError
}

type RestError struct {
	ErrStatus  int         `json:"status,omitempty" example:"403"`
	ErrError   string      `json:"error,omitempty" example:"Forbidden"`
	ErrMessage interface{} `json:"message,omitempty"`
	Timestamp  int64       `json:"timestamp,omitempty" example:"1692753495"`
}

func (e RestError) ErrBody() RestError {
	return e
}

func (e RestError) Error() string {
	return fmt.Sprintf("status: %d - errors: %s - causes: %v", e.ErrStatus, e.ErrError, e.ErrMessage)
}

func (e RestError) Status() int {
	return e.ErrStatus
}

func (e RestError) Causes() interface{} {
	return e.ErrMessage
}

func NewRestError(status int, err string, causes interface{}, debug bool) RestErr {
	restError := RestError{
		ErrStatus: status,
		ErrError:  err,
		Timestamp: time.Now().Unix(),
	}
	if debug {
		restError.ErrMessage = causes
	}
	return restError
}

func NewRestErrorWithMessage(status int, err string, causes interface{}) RestErr {
	return RestError{
		ErrStatus:  status,
		ErrError:   err,
		ErrMessage: causes,
		Timestamp:  time.Now().Unix(),
	}
}

func NewRestErrorFromBytes(bytes []byte) (RestErr, error) {
	var apiErr RestErr
	if err := json.Unmarshal(bytes, &apiErr); err != nil {
		return nil, errors.New("invalid json")
	}
	return apiErr, nil
}

func NewBadRequestError(c *app.RequestContext, causes interface{}, debug bool) (int, interface{}) {
	restError := RestError{
		ErrStatus: consts.StatusBadRequest,
		ErrError:  BadRequest.Error(),
		Timestamp: time.Now().Unix(),
	}
	if debug {
		restError.ErrMessage = causes
	}

	return consts.StatusBadRequest, restError
}

func NewNotFoundError(c *app.RequestContext, causes interface{}, debug bool) (int, interface{}) {
	restError := RestError{
		ErrStatus: consts.StatusNotFound,
		ErrError:  NotFound.Error(),
		Timestamp: time.Now().Unix(),
	}

	if debug {
		restError.ErrMessage = causes
	}

	return consts.StatusNotFound, restError
}

func NewUnauthorizedError(c *app.RequestContext, causes interface{}, debug bool) (int, interface{}) {
	restError := RestError{
		ErrStatus: consts.StatusUnauthorized,
		ErrError:  Unauthorized.Error(),
		Timestamp: time.Now().Unix(),
	}

	if debug {
		restError.ErrMessage = causes
	}

	return consts.StatusUnauthorized, restError
}

func NewForbiddenError(c *app.RequestContext, causes interface{}, debug bool) (int, interface{}) {
	restError := RestError{
		ErrStatus: consts.StatusForbidden,
		ErrError:  Forbidden.Error(),
		Timestamp: time.Now().Unix(),
	}

	if debug {
		restError.ErrMessage = causes
	}

	return consts.StatusForbidden, restError
}

func NewInternalServerError(c *app.RequestContext, causes interface{}, debug bool) (int, interface{}) {
	restError := RestError{
		ErrStatus: consts.StatusInternalServerError,
		ErrError:  InternalServerError.Error(),
		Timestamp: time.Now().Unix(),
	}

	if debug {
		restError.ErrMessage = causes
	}

	return consts.StatusInternalServerError, restError
}

func ParseErrors(err error, debug bool) RestErr {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return NewRestError(consts.StatusNotFound, ErrNotFound, err.Error(), debug)
	case errors.Is(err, context.DeadlineExceeded):
		return NewRestError(consts.StatusRequestTimeout, ErrRequestTimeout, errors.New("error: 1101").Error(), debug)
	case errors.Is(err, Unauthorized):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case errors.Is(err, WrongCredentials):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "field validation"):
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return NewRestError(consts.StatusBadRequest, ErrBadRequest, validationErrors.Error(), debug)
		}
		return parseValidatorError(err, debug)
	case strings.Contains(strings.ToLower(err.Error()), "Lock wait timeout exceeded; try restarting transaction"):
		return NewRestError(consts.StatusRequestTimeout, ErrRequestTimeout, errors.New("error: 3311").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), ".kafkaService-prod.server: i/o timeout"):
		return NewRestError(consts.StatusInternalServerError, ErrInternalServerError, errors.New("error: 1102").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "sqlstate"):
		return parseSqlErrors(err, debug)
	case strings.Contains(strings.ToLower(err.Error()), "required header"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "rate limit exceeded"):
		return NewRestError(consts.StatusTooManyRequests, ErrTooManyRequests, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "either filename/content-type is empty"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "invalid authorization header format"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "base64"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "unmarshal"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "uuid"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "unauthorized"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "cookie"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "token"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "bcrypt"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "no document in result"):
		return NewRestError(consts.StatusNotFound, ErrNotFound, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "no documents in result"):
		return NewRestError(consts.StatusNotFound, ErrNotFound, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "there is no uploaded file"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "too many requests"):
		return NewRestError(consts.StatusTooManyRequests, ErrTooManyRequests, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "invalid memory address or nil pointer"):
		return NewRestError(consts.StatusInternalServerError, ErrInternalServerError, errors.New("something went wrong").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "no rows in result set"):
		return NewRestError(consts.StatusNotFound, ErrNotFound, errors.New("requested resources are not found").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "room ownership is not match"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("only the owner of this room is permitted to forcibly end this meeting").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "rpc error: code = Internal desc = err: notFound"):
		return NewRestError(consts.StatusNotFound, ErrNotFound, errors.New("not found").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "supplied id are not valid"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "cannot be updated"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "bad data size when unmarshaling"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("bad body parameters supplied").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "invalid UUID length"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("bad parameters").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "the provided value is not a valid url"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "file size too large, cannot exceed > 150MB"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "authorization header is not provided"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "invalid authorization header format"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "unsupported authorization"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "token is invalid"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "token has expired"):
		return NewRestError(consts.StatusUnauthorized, ErrUnauthorized, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "user not found"):
		return NewRestError(consts.StatusNotFound, ErrNotFound, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "incorrect password"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("incorrect username/password").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "this resources already used or expired"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "blocked session"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "should supply email or phone number"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("should supply email or phone number").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "invalid_grant"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "this email already registered"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "this user is not being activated by the stokist yet"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "failed on the"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, errors.New("bad input").Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "nomor whatsapp / email should supply a valid email or phone number"):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), "mismatch password"):
		return NewRestError(consts.StatusForbidden, ErrForbidden, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), unierror.ErrNoPlatformKey.Error()):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), unierror.ErrMismatchPlatformKey.Error()):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	case strings.Contains(strings.ToLower(err.Error()), unierror.ErrMismatchAuthPlatformKey.Error()):
		return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
	default:
		if restErr, ok := err.(*RestError); ok {
			return restErr
		}
		return NewRestError(consts.StatusInternalServerError, ErrInternalServerError, errors.Cause(err).Error(), debug)
	}
}

func parseSqlErrors(err error, debug bool) RestErr {
	return NewRestError(consts.StatusBadRequest, ErrBadRequest, err.Error(), debug)
}

func parseValidatorError(err error, debug bool) RestErr {
	if strings.Contains(err.Error(), "Password") {
		return NewRestError(consts.StatusBadRequest, ErrInvalidPassword, err, debug)
	}

	if strings.Contains(err.Error(), "Email") {
		return NewRestError(consts.StatusBadRequest, ErrInvalidEmail, err, debug)
	}

	return NewRestError(consts.StatusBadRequest, ErrInvalidField, err, debug)
}

func ErrorResponse(err error, debug bool) (int, interface{}) {
	return ParseErrors(err, debug).Status(), ParseErrors(err, debug)
}

func ErrorCtxResponse(err error, debug bool) (int, interface{}) {
	restErr := ParseErrors(err, debug)

	return restErr.Status(), restErr.ErrBody()
}
