package mapper

import (
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CustomerToDto(customer *repository.Customer) *pb.Customer {
	return &pb.Customer{
		Uid:               customer.Uid,
		CustomerAppId:     customer.CustomerAppID,
		PaymentCustomerId: customer.PaymentCustomerID,
		CustomerName:      customer.CustomerName,
		CreatedAt:         timestamppb.New(customer.CreatedAt.Time),
		Email:             &customer.Email.String,
		PhoneNumber:       &customer.PhoneNumber.String,
	}
}

func PaymentChannelToDto(paymentChannel *repository.PaymentChannel) *pb.PaymentChannel {
	return &pb.PaymentChannel{
		Uid:                paymentChannel.Uid,
		PaymentChannelName: paymentChannel.Pcname,
		PaymentChannelType: paymentChannel.PcType,
		PaymentLogoSrc:     paymentChannel.LogoSrc,
		MinAmount:          paymentChannel.MinAmount.InexactFloat64(),
		MaxAmount:          paymentChannel.MaxAmount.InexactFloat64(),
		Tax:                paymentChannel.Tax.InexactFloat64(),
		IsTaxPercentage:    paymentChannel.IsTaxPercentage,
		IsActive:           paymentChannel.IsActive,
		IsAvailable:        paymentChannel.IsAvailable,
	}
}

func PaymentChannelsToDto(args []*repository.PaymentChannel) []*pb.PaymentChannel {
	list := make([]*pb.PaymentChannel, 0, len(args))
	for _, pc := range args {
		list = append(list, PaymentChannelToDto(pc))
	}

	return list
}

func PaymentToDto(arg *repository.PaymentMethod) *pb.PaymentMethod {
	amount, _ := arg.PaymentAmount.Float64()
	return &pb.PaymentMethod{
		Uid:                         arg.Uid,
		PaymentMethodId:             arg.PaymentMethodID,
		PaymentRequestId:            &arg.PaymentRequestID.String,
		PaymentReferenceId:          arg.PaymentReferenceID,
		PaymentBusinessId:           arg.PaymentBusinessID,
		PaymentCustomerId:           arg.PaymentCustomerID,
		PaymentType:                 arg.PaymentType,
		PaymentStatus:               arg.PaymentStatus,
		PaymentReusability:          arg.PaymentReusability,
		PaymentChannel:              arg.PaymentChannel,
		PaymentAmount:               amount,
		PaymentQrCode:               &arg.PaymentQrCode.String,
		PaymentVirtualAccountNumber: &arg.PaymentVirtualAccountNumber.String,
		PaymentUrl:                  &arg.PaymentUrl.String,
		PaymentDescription:          arg.PaymentDescription,
		CreatedAt:                   timestamppb.New(arg.CreatedAt.Time),
		UpdatedAt:                   timestamppb.New(arg.UpdatedAt.Time),
		ExpiresAt:                   timestamppb.New(arg.ExpiresAt.Time),
		PaidAt:                      timestamppb.New(arg.PaidAt.Time),
	}
}
