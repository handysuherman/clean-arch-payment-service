package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/mapper"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror/tracing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func (u *usecaseImpl) Create(ctx context.Context, arg *models.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UsecaseImpl.Create")
	defer span.Finish()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if payload, err := u.repo.GetCreatePaymentIdempotencyKey(ctx, arg.XIdempotencyKey); err == nil && payload != nil {
		return payload, nil
	}

	phoneNumber, err := u.validateParams(span, arg)
	if err != nil {
		return nil, tracing.TraceWithError(span, err)
	}

	customer, err := u.processGetCustomer(ctx, span, arg, phoneNumber)
	if err != nil {
		return nil, tracing.TraceWithError(span, err)
	}

	switch arg.PaymentType {
	case payment.METHODE_TYPE_EWALLET:
		return u.processEwalletPayment(ctx, span, arg, customer)
	case payment.METHODE_TYPE_QR_CODE:
		return u.processQrCodePayment(ctx, span, arg, customer)
	case payment.METHODE_TYPE_VIRTUAL_ACCOUNT:
		return u.processVirtualAccountPayment(ctx, span, arg, customer)
	default:
		return nil, unierror.ErrUnsupportedPaymentType
	}
}

func (u *usecaseImpl) validateParams(span opentracing.Span, arg *models.CreatePaymentRequest) (string, error) {
	_, err := payment_method.NewPaymentMethodTypeFromValue(arg.PaymentType)
	if err != nil {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "payment_method.NewPaymentMethodTypeFromValue.err", arg.PaymentType),
			err,
		)
	}

	err = u.isSupportedPaymentType(arg.PaymentType)
	if err != nil {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "u.isSupportedPaymentType.err", arg.PaymentType),
			err,
		)
	}

	res, err := helper.ValidatePhoneNumber(arg.CustomerPhoneNumber)
	if err != nil {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "helper.ValidatePhoneNumber.err", arg.CustomerPhoneNumber),
			unierror.ErrInvalidCustomerPhoneNumberInput,
		)
	}

	if arg.PaymentReferenceId == "" {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "arg.GetPaymentReferenceId", arg.PaymentReferenceId),
			unierror.ErrReferenceIDShouldNotBeEmpty,
		)
	}

	if arg.PaymentAmount < 100 {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "arg.GetPaymentAmount", arg.PaymentAmount),
			unierror.ErrInvalidAmount,
		)
	}

	successURLOK := helper.IsValidURL(arg.PaymentSuccessReturnUrl)
	if !successURLOK {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "helper.IsValidURL", arg.PaymentSuccessReturnUrl),
			unierror.ErrInvalidSuccessURL,
		)
	}

	failureURLOK := helper.IsValidURL(arg.PaymentFailureReturnUrl)
	if !failureURLOK {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "helper.IsValidURL", arg.PaymentFailureReturnUrl),
			unierror.ErrInvalidFailureURL,
		)
	}

	if arg.ExpiryHour < 72 {
		return "", u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "arg.GetExpiryHour", arg.ExpiryHour),
			unierror.ErrExpiryLessThan3Days,
		)
	}

	return res, nil
}

func (u *usecaseImpl) isSupportedPaymentType(typ string) error {
	switch typ {
	case "EWALLET", "QR_CODE", "VIRTUAL_ACCOUNT":
		return nil
	default:
		return unierror.ErrUnsupportedPaymentType
	}
}

func (u *usecaseImpl) processGetCustomer(
	ctx context.Context,
	span opentracing.Span,
	arg *models.CreatePaymentRequest,
	phoneNumber string,
) (*repository.Customer, error) {
	if payload, err := u.repo.GetCustomerCache(ctx, *arg.CustomerUid); err == nil && payload != nil {
		return payload, nil
	}

	customer, err := u.repo.GetCustomerByCustomerAppID(ctx, *arg.CustomerUid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			customerUlid, err := helper.GenerateULID()
			if err != nil {
				return nil, u.errorResponse(span, "helper.GenerateULID.err", err)
			}

			customerArg := repository.CreateCustomerTxParams{
				CreateCustomer: repository.CreateCustomerParams{
					Uid:               customerUlid.String(),
					CustomerAppID:     customerUlid.String(),
					PaymentCustomerID: "",
					CustomerName:      arg.CustomerName,
					PhoneNumber: pgtype.Text{
						String: phoneNumber,
						Valid:  true,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:  time.Now(),
						Valid: true,
					},
				},
			}

			customerTx, err := u.repo.CreateCustomerTx(ctx, &customerArg)
			if err != nil {
				return nil, u.errorResponse(span, "u.repo.CreateCustomerTx.err", err)
			}

			u.repo.PutCustomerCache(ctx, customerTx.Customer)
			return customerTx.Customer, nil
		}
		return nil, u.errorResponse(span, "u.repo.GetCustomerByCustomerAppID.err", err)
	}

	u.repo.PutCustomerCache(ctx, customer)
	return customer, nil
}

func (u *usecaseImpl) processEwalletPayment(
	ctx context.Context,
	span opentracing.Span,
	arg *models.CreatePaymentRequest,
	cust *repository.Customer,
) (*pb.CreatePaymentResponse, error) {
	channelCode, err := payment_method.NewEWalletChannelCodeFromValue(arg.PaymentChannel)
	if err != nil {
		return nil, u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "payment_method.NewEWalletChannelCodeFromValue.err", arg.PaymentChannel),
			err,
		)
	}

	createArg := repository.CreateEwalletPaymentParams{
		CustomerName:      cust.CustomerName,
		CustomerPaymentID: cust.PaymentCustomerID,
		CustomerNumber:    cust.PhoneNumber.String,
		ReferenceID:       arg.PaymentReferenceId,
		Description:       arg.PaymentDescription,
		Amount:            arg.PaymentAmount,
		ChannelCode:       *channelCode,
		Expiry:            time.Now().Add(time.Duration(arg.ExpiryHour+6) * time.Hour),
		SuccessReturnURL:  arg.PaymentSuccessReturnUrl,
		FailureReturnURL:  arg.PaymentFailureReturnUrl,
	}

	res, err := u.repo.CreateEwalletPaymentTx(ctx, &repository.CreateEwalletPaymentTxParams{CreateEwalletPayment: createArg})
	if err != nil {
		return nil, u.errorResponse(
			span,
			"u.repo.CreateEwalletPaymentTx.err",
			err,
		)
	}

	respDto := &pb.CreatePaymentResponse{
		Customer:      mapper.CustomerToDto(cust),
		PaymentMethod: mapper.PaymentToDto(res.Payment),
	}

	u.repo.PutCreatePaymentIdempotencyKey(ctx, arg.XIdempotencyKey, respDto)
	u.repo.PutCache(ctx, res.Payment)
	return respDto, nil
}

func (u *usecaseImpl) processQrCodePayment(
	ctx context.Context,
	span opentracing.Span,
	arg *models.CreatePaymentRequest,
	cust *repository.Customer,
) (*pb.CreatePaymentResponse, error) {
	channelCode, err := payment_method.NewQRCodeChannelCodeFromValue(arg.PaymentChannel)
	if err != nil {
		return nil, u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "payment_method.NewQRCodeChannelCodeFromValue.err", arg.PaymentChannel),
			err,
		)
	}

	createArg := repository.CreateQrCodePaymentParams{
		CustomerName:      cust.CustomerName,
		CustomerPaymentID: cust.PaymentCustomerID,
		CustomerNumber:    cust.PhoneNumber.String,
		ReferenceID:       arg.PaymentReferenceId,
		Description:       arg.PaymentDescription,
		Amount:            arg.PaymentAmount,
		ChannelCode:       *channelCode,
		Expiry:            time.Now().Add(time.Duration(arg.ExpiryHour) * time.Hour),
		SuccessReturnURL:  arg.PaymentSuccessReturnUrl,
		FailureReturnURL:  arg.PaymentFailureReturnUrl,
	}

	res, err := u.repo.CreateQrCodePaymentTx(ctx, &repository.CreateQrCodePaymentTxParams{CreateQrCodePayment: createArg})
	if err != nil {
		return nil, u.errorResponse(
			span,
			"u.repo.CreateQrCodePaymentTx.err",
			err,
		)
	}

	respDto := &pb.CreatePaymentResponse{
		Customer:      mapper.CustomerToDto(cust),
		PaymentMethod: mapper.PaymentToDto(res.Payment),
	}

	u.repo.PutCreatePaymentIdempotencyKey(ctx, arg.XIdempotencyKey, respDto)
	u.repo.PutCache(ctx, res.Payment)
	return respDto, nil
}

func (u *usecaseImpl) processVirtualAccountPayment(
	ctx context.Context,
	span opentracing.Span,
	arg *models.CreatePaymentRequest,
	cust *repository.Customer,
) (*pb.CreatePaymentResponse, error) {
	channelCode, err := payment_method.NewVirtualAccountChannelCodeFromValue(arg.PaymentChannel)
	if err != nil {
		return nil, u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "payment_method.NewVirtualAccountChannelCodeFromValue.err", arg.PaymentChannel),
			err,
		)
	}

	createArg := repository.CreateVirtualAccountBankPaymentParams{
		CustomerName:      cust.CustomerName,
		CustomerPaymentID: cust.PaymentCustomerID,
		CustomerNumber:    cust.PhoneNumber.String,
		ReferenceID:       arg.PaymentReferenceId,
		Description:       arg.PaymentDescription,
		Amount:            arg.PaymentAmount,
		ChannelCode:       *channelCode,
		Expiry:            time.Now().Add(time.Duration(arg.ExpiryHour) * time.Hour),
		SuccessReturnURL:  arg.PaymentSuccessReturnUrl,
		FailureReturnURL:  arg.PaymentFailureReturnUrl,
	}

	res, err := u.repo.CreateVirtualAccountBankPaymentTx(ctx, &repository.CreateVirtualAccountBankPaymentTxParams{CreateVirtualAccountBankPayment: createArg})
	if err != nil {
		return nil, u.errorResponse(
			span,
			"u.repo.CreateVirtualAccountBankPaymentTx.err",
			err,
		)
	}

	respDto := &pb.CreatePaymentResponse{
		Customer:      mapper.CustomerToDto(cust),
		PaymentMethod: mapper.PaymentToDto(res.Payment),
	}

	u.repo.PutCreatePaymentIdempotencyKey(ctx, arg.XIdempotencyKey, respDto)
	u.repo.PutCache(ctx, res.Payment)
	return respDto, nil
}
