// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.0
// source: captures.proto

package capturespb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// models
type Filters struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ips       []string `protobuf:"bytes,1,rep,name=ips,proto3" json:"ips,omitempty"`
	Protocols []string `protobuf:"bytes,2,rep,name=protocols,proto3" json:"protocols,omitempty"`
	Ports     []string `protobuf:"bytes,3,rep,name=ports,proto3" json:"ports,omitempty"`
	Macs      []string `protobuf:"bytes,4,rep,name=macs,proto3" json:"macs,omitempty"`
}

func (x *Filters) Reset() {
	*x = Filters{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Filters) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Filters) ProtoMessage() {}

func (x *Filters) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Filters.ProtoReflect.Descriptor instead.
func (*Filters) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{0}
}

func (x *Filters) GetIps() []string {
	if x != nil {
		return x.Ips
	}
	return nil
}

func (x *Filters) GetProtocols() []string {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *Filters) GetPorts() []string {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *Filters) GetMacs() []string {
	if x != nil {
		return x.Macs
	}
	return nil
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{1}
}

// requests
type CaptureRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Duration  int32                  `protobuf:"varint,1,opt,name=duration,proto3" json:"duration,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Filter    *Filters               `protobuf:"bytes,3,opt,name=filter,proto3" json:"filter,omitempty"`
}

func (x *CaptureRequest) Reset() {
	*x = CaptureRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CaptureRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CaptureRequest) ProtoMessage() {}

func (x *CaptureRequest) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CaptureRequest.ProtoReflect.Descriptor instead.
func (*CaptureRequest) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{2}
}

func (x *CaptureRequest) GetDuration() int32 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *CaptureRequest) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *CaptureRequest) GetFilter() *Filters {
	if x != nil {
		return x.Filter
	}
	return nil
}

type HCNRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *HCNRequest) Reset() {
	*x = HCNRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HCNRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HCNRequest) ProtoMessage() {}

func (x *HCNRequest) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HCNRequest.ProtoReflect.Descriptor instead.
func (*HCNRequest) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{3}
}

func (x *HCNRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

// responses
type CaptureResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result    string                 `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *CaptureResponse) Reset() {
	*x = CaptureResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CaptureResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CaptureResponse) ProtoMessage() {}

func (x *CaptureResponse) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CaptureResponse.ProtoReflect.Descriptor instead.
func (*CaptureResponse) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{4}
}

func (x *CaptureResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

func (x *CaptureResponse) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

type HCNResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HcnResult []byte `protobuf:"bytes,1,opt,name=hcn_result,json=hcnResult,proto3" json:"hcn_result,omitempty"`
}

func (x *HCNResponse) Reset() {
	*x = HCNResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_captures_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HCNResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HCNResponse) ProtoMessage() {}

func (x *HCNResponse) ProtoReflect() protoreflect.Message {
	mi := &file_captures_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HCNResponse.ProtoReflect.Descriptor instead.
func (*HCNResponse) Descriptor() ([]byte, []int) {
	return file_captures_proto_rawDescGZIP(), []int{5}
}

func (x *HCNResponse) GetHcnResult() []byte {
	if x != nil {
		return x.HcnResult
	}
	return nil
}

var File_captures_proto protoreflect.FileDescriptor

var file_captures_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x63, 0x0a, 0x07, 0x46,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x70, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x70, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x12, 0x0a, 0x04,
	0x6d, 0x61, 0x63, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6d, 0x61, 0x63, 0x73,
	0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x91, 0x01, 0x0a, 0x0e, 0x43, 0x61,
	0x70, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x12, 0x29, 0x0a, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x73, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0x20, 0x0a,
	0x0a, 0x48, 0x43, 0x4e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22,
	0x63, 0x0a, 0x0f, 0x43, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x22, 0x2c, 0x0a, 0x0b, 0x48, 0x43, 0x4e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x68, 0x63, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x68, 0x63, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x32, 0xc9, 0x01, 0x0a, 0x0e, 0x43, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x47, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x72, 0x74, 0x43, 0x61,
	0x70, 0x74, 0x75, 0x72, 0x65, 0x12, 0x18, 0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73,
	0x2e, 0x43, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x19, 0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x43, 0x61, 0x70, 0x74, 0x75,
	0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x31,
	0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x70, 0x43, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x12, 0x0f, 0x2e,
	0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0f,
	0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x12, 0x3b, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x48, 0x43, 0x4e, 0x4c, 0x6f, 0x67, 0x73, 0x12,
	0x14, 0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2e, 0x48, 0x43, 0x4e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73,
	0x2e, 0x48, 0x43, 0x4e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x0d,
	0x5a, 0x0b, 0x2f, 0x63, 0x61, 0x70, 0x74, 0x75, 0x72, 0x65, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_captures_proto_rawDescOnce sync.Once
	file_captures_proto_rawDescData = file_captures_proto_rawDesc
)

func file_captures_proto_rawDescGZIP() []byte {
	file_captures_proto_rawDescOnce.Do(func() {
		file_captures_proto_rawDescData = protoimpl.X.CompressGZIP(file_captures_proto_rawDescData)
	})
	return file_captures_proto_rawDescData
}

var file_captures_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_captures_proto_goTypes = []interface{}{
	(*Filters)(nil),               // 0: captures.Filters
	(*Empty)(nil),                 // 1: captures.Empty
	(*CaptureRequest)(nil),        // 2: captures.CaptureRequest
	(*HCNRequest)(nil),            // 3: captures.HCNRequest
	(*CaptureResponse)(nil),       // 4: captures.CaptureResponse
	(*HCNResponse)(nil),           // 5: captures.HCNResponse
	(*timestamppb.Timestamp)(nil), // 6: google.protobuf.Timestamp
}
var file_captures_proto_depIdxs = []int32{
	6, // 0: captures.CaptureRequest.timestamp:type_name -> google.protobuf.Timestamp
	0, // 1: captures.CaptureRequest.filter:type_name -> captures.Filters
	6, // 2: captures.CaptureResponse.timestamp:type_name -> google.protobuf.Timestamp
	2, // 3: captures.CaptureService.StartCapture:input_type -> captures.CaptureRequest
	1, // 4: captures.CaptureService.StopCapture:input_type -> captures.Empty
	3, // 5: captures.CaptureService.GetHCNLogs:input_type -> captures.HCNRequest
	4, // 6: captures.CaptureService.StartCapture:output_type -> captures.CaptureResponse
	1, // 7: captures.CaptureService.StopCapture:output_type -> captures.Empty
	5, // 8: captures.CaptureService.GetHCNLogs:output_type -> captures.HCNResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_captures_proto_init() }
func file_captures_proto_init() {
	if File_captures_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_captures_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Filters); i {
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
		file_captures_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_captures_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CaptureRequest); i {
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
		file_captures_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HCNRequest); i {
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
		file_captures_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CaptureResponse); i {
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
		file_captures_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HCNResponse); i {
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
			RawDescriptor: file_captures_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_captures_proto_goTypes,
		DependencyIndexes: file_captures_proto_depIdxs,
		MessageInfos:      file_captures_proto_msgTypes,
	}.Build()
	File_captures_proto = out.File
	file_captures_proto_rawDesc = nil
	file_captures_proto_goTypes = nil
	file_captures_proto_depIdxs = nil
}
