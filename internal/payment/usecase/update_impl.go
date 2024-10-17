package usecase

import (
	"context"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

func (u *usecaseImpl) Update(ctx context.Context, arg *models.UpdatePaymentRequest) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UsecaseImpl.Update")
	defer span.Finish()

	updateArg := repository.UpdatePaymentMethodCustomerParams{
		PaymentMethodID:   arg.PaymentMethodId,
		PaymentCustomerID: arg.PaymentCustomerId,
		PaymentStatus: pgtype.Text{
			String: arg.PaymentStatus,
			Valid:  true,
		},
	}

	if arg.PaymentFailureCode != nil {
		updateArg.PaymentFailureCode = pgtype.Text{
			String: *arg.PaymentFailureCode,
			Valid:  arg.PaymentFailureCode != nil,
		}
	}

	if arg.PaymentStatus == payment.STATUS_SUCCEEDED {
		updateArg.PaidAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: arg.UpdatedAt != nil,
		}
	}

	if arg.UpdatedAt != nil {
		updatedAt := arg.UpdatedAt

		updateArg.UpdatedAt = pgtype.Timestamptz{
			Time:  *updatedAt,
			Valid: arg.UpdatedAt != nil,
		}

	}

	res, err := u.repo.UpdateTx(ctx, &repository.UpdateTxParams{
		UpdateParams: updateArg,
	})
	if err != nil {
		return u.errorResponse(span, "u.repo.UpdateTx.err", err)
	}

	err = u.worker.PaymentStatusUpdated(ctx, &models.PaymentStatusUpdatedTask{PaymentMethod: res.Payment})
	if err != nil {
		return u.errorResponse(span, "u.worker.PaymentStatusUpdated.err", err)
	}

	u.repo.PutCache(ctx, res.Payment)
	return nil
}
