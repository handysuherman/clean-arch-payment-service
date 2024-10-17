package grpc

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/metrics"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/domain"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	grpcError "github.com/handysuherman/clean-arch-payment-service/internal/pkg/grpc_error"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/opentracing/opentracing-go"
)

type grpcHandler struct {
	pb.UnimplementedPaymentServiceServer
	log     logger.Logger
	cfg     *config.App
	v       *validator.Validate
	usecase domain.Usecase
	metrics *metrics.Metrics
}

func New(
	log logger.Logger,
	cfg *config.App,
	v *validator.Validate,
	usecase domain.Usecase,
	metrics *metrics.Metrics,
) *grpcHandler {
	return &grpcHandler{
		log:     log.WithPrefix(fmt.Sprintf("%s-%s", "payment", constants.Handler)),
		cfg:     cfg,
		usecase: usecase,
		v:       v,
		metrics: metrics,
	}
}

func (h *grpcHandler) Create(ctx context.Context, arg *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	h.metrics.CreatePaymentGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcHandler.Create")
	defer span.Finish()

	params := models.NewCreatePaymentRequestParams(arg)
	if err := h.v.StructCtx(ctx, params); err != nil {
		return nil, h.errorResponse(span, err, "h.v.StructCtx.err", true)
	}

	res, err := h.usecase.Create(ctx, params)
	if err != nil {
		return nil, h.errorResponse(span, err, "h.usecase.Create.err", false)
	}

	h.metrics.SuccessGrpcRequest.Inc()
	return res, nil
}

func (h *grpcHandler) GetByID(ctx context.Context, arg *pb.GetByIDPaymentRequest) (*pb.GetByIDPaymentResponse, error) {
	h.metrics.GetPaymentByIDGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcHandler.GetByID")
	defer span.Finish()

	params := models.NewGetByIDPaymentRequestParams(arg)
	if err := h.v.StructCtx(ctx, params); err != nil {
		return nil, h.errorResponse(span, err, "h.v.StructCtx.err", true)
	}

	res, err := h.usecase.GetByID(ctx, params)
	if err != nil {
		return nil, h.errorResponse(span, err, "h.usecase.GetByID.err", false)
	}

	h.metrics.SuccessGrpcRequest.Inc()
	return res, nil
}

func (h *grpcHandler) GetChannel(ctx context.Context, arg *pb.GetPaymentChannelRequest) (*pb.GetPaymentChannelResponse, error) {
	h.metrics.GetPaymentChannelGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcHandler.GetChannel")
	defer span.Finish()

	params := models.NewGetPaymentChannelRequestParams(arg)
	if err := h.v.StructCtx(ctx, params); err != nil {
		return nil, h.errorResponse(span, err, "h.v.StructCtx.err", true)
	}

	res, err := h.usecase.GetAvailableChannel(ctx, params)
	if err != nil {
		return nil, h.errorResponse(span, err, "h.usecase.GetAvailableChannel.err", false)
	}

	h.metrics.SuccessGrpcRequest.Inc()
	return res, nil
}

func (h *grpcHandler) GetAvailableChannels(ctx context.Context, arg *pb.GetPaymentChannelsRequest) (*pb.GetPaymentChannelsResponse, error) {
	h.metrics.GetAvailablePaymentChannelsGrpcRequest.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcHandler.GetAvailableChannels")
	defer span.Finish()

	params := models.NewGetPaymentChannelsRequestParams(arg)
	if err := h.v.StructCtx(ctx, params); err != nil {
		return nil, h.errorResponse(span, err, "h.v.StructCtx.err", true)
	}

	res, err := h.usecase.GetAvailableChannels(ctx, params)
	if err != nil {
		return nil, h.errorResponse(span, err, "h.usecase.GetAvailableChannels.err", false)
	}

	h.metrics.SuccessGrpcRequest.Inc()
	return res, nil
}

func (h *grpcHandler) errorResponse(span opentracing.Span, err error, details string, logError bool) error {
	if logError {
		errfmt := fmt.Errorf("%s: %v", details, err)
		h.log.Warn(errfmt)
		tracing.TraceWithError(span, errfmt)
	}

	h.metrics.ErrorGrpcRequest.Inc()
	return grpcError.ErrorResponse(err)
}

func (h *grpcHandler) OnConfigUpdate(key string, config *config.App) {
	h.log.Infof("received an update from '%s' key", key)

	h.cfg = config

	h.log.Infof("updated configuration from '%s' key successfully applied", key)
}
