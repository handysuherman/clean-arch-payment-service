// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.2
// source: payment_channel.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PaymentChannel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid                string  `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	PaymentChannelName string  `protobuf:"bytes,2,opt,name=payment_channel_name,json=paymentChannelName,proto3" json:"payment_channel_name,omitempty"`
	PaymentChannelType string  `protobuf:"bytes,3,opt,name=payment_channel_type,json=paymentChannelType,proto3" json:"payment_channel_type,omitempty"`
	PaymentLogoSrc     string  `protobuf:"bytes,4,opt,name=payment_logo_src,json=paymentLogoSrc,proto3" json:"payment_logo_src,omitempty"`
	MinAmount          float64 `protobuf:"fixed64,5,opt,name=min_amount,json=minAmount,proto3" json:"min_amount,omitempty"`
	MaxAmount          float64 `protobuf:"fixed64,6,opt,name=max_amount,json=maxAmount,proto3" json:"max_amount,omitempty"`
	Tax                float64 `protobuf:"fixed64,7,opt,name=tax,proto3" json:"tax,omitempty"`
	IsTaxPercentage    bool    `protobuf:"varint,8,opt,name=is_tax_percentage,json=isTaxPercentage,proto3" json:"is_tax_percentage,omitempty"`
	IsActive           bool    `protobuf:"varint,9,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
	IsAvailable        bool    `protobuf:"varint,10,opt,name=is_available,json=isAvailable,proto3" json:"is_available,omitempty"`
}

func (x *PaymentChannel) Reset() {
	*x = PaymentChannel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_payment_channel_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PaymentChannel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaymentChannel) ProtoMessage() {}

func (x *PaymentChannel) ProtoReflect() protoreflect.Message {
	mi := &file_payment_channel_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaymentChannel.ProtoReflect.Descriptor instead.
func (*PaymentChannel) Descriptor() ([]byte, []int) {
	return file_payment_channel_proto_rawDescGZIP(), []int{0}
}

func (x *PaymentChannel) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *PaymentChannel) GetPaymentChannelName() string {
	if x != nil {
		return x.PaymentChannelName
	}
	return ""
}

func (x *PaymentChannel) GetPaymentChannelType() string {
	if x != nil {
		return x.PaymentChannelType
	}
	return ""
}

func (x *PaymentChannel) GetPaymentLogoSrc() string {
	if x != nil {
		return x.PaymentLogoSrc
	}
	return ""
}

func (x *PaymentChannel) GetMinAmount() float64 {
	if x != nil {
		return x.MinAmount
	}
	return 0
}

func (x *PaymentChannel) GetMaxAmount() float64 {
	if x != nil {
		return x.MaxAmount
	}
	return 0
}

func (x *PaymentChannel) GetTax() float64 {
	if x != nil {
		return x.Tax
	}
	return 0
}

func (x *PaymentChannel) GetIsTaxPercentage() bool {
	if x != nil {
		return x.IsTaxPercentage
	}
	return false
}

func (x *PaymentChannel) GetIsActive() bool {
	if x != nil {
		return x.IsActive
	}
	return false
}

func (x *PaymentChannel) GetIsAvailable() bool {
	if x != nil {
		return x.IsAvailable
	}
	return false
}

var File_payment_channel_proto protoreflect.FileDescriptor

var file_payment_channel_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xec, 0x02, 0x0a, 0x0e, 0x50, 0x61, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x30, 0x0a, 0x14,
	0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x70, 0x61, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x30,
	0x0a, 0x14, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x70, 0x61,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x28, 0x0a, 0x10, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x6f,
	0x5f, 0x73, 0x72, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x70, 0x61, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x67, 0x6f, 0x53, 0x72, 0x63, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x69,
	0x6e, 0x5f, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09,
	0x6d, 0x69, 0x6e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x61, 0x78,
	0x5f, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x6d,
	0x61, 0x78, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x78, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x74, 0x61, 0x78, 0x12, 0x2a, 0x0a, 0x11, 0x69, 0x73,
	0x5f, 0x74, 0x61, 0x78, 0x5f, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x69, 0x73, 0x54, 0x61, 0x78, 0x50, 0x65, 0x72, 0x63,
	0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x73, 0x5f, 0x61, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x41, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x69, 0x73, 0x5f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61,
	0x62, 0x6c, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x69, 0x73, 0x41, 0x76, 0x61,
	0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x69, 0x6b, 0x61, 0x6e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x77, 0x64, 0x2d, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x2d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_payment_channel_proto_rawDescOnce sync.Once
	file_payment_channel_proto_rawDescData = file_payment_channel_proto_rawDesc
)

func file_payment_channel_proto_rawDescGZIP() []byte {
	file_payment_channel_proto_rawDescOnce.Do(func() {
		file_payment_channel_proto_rawDescData = protoimpl.X.CompressGZIP(file_payment_channel_proto_rawDescData)
	})
	return file_payment_channel_proto_rawDescData
}

var file_payment_channel_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_payment_channel_proto_goTypes = []interface{}{
	(*PaymentChannel)(nil), // 0: PaymentChannel
}
var file_payment_channel_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_payment_channel_proto_init() }
func file_payment_channel_proto_init() {
	if File_payment_channel_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_payment_channel_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PaymentChannel); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_payment_channel_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_payment_channel_proto_goTypes,
		DependencyIndexes: file_payment_channel_proto_depIdxs,
		MessageInfos:      file_payment_channel_proto_msgTypes,
	}.Build()
	File_payment_channel_proto = out.File
	file_payment_channel_proto_rawDesc = nil
	file_payment_channel_proto_goTypes = nil
	file_payment_channel_proto_depIdxs = nil
}
