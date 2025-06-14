// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v3.21.12
// source: protofiles/services.proto

package protofiles

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// done
type ServerInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          string                 `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerInfo) Reset() {
	*x = ServerInfo{}
	mi := &file_protofiles_services_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerInfo) ProtoMessage() {}

func (x *ServerInfo) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerInfo.ProtoReflect.Descriptor instead.
func (*ServerInfo) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{0}
}

func (x *ServerInfo) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ServerInfo) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

type RegisterResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterResponse) Reset() {
	*x = RegisterResponse{}
	mi := &file_protofiles_services_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterResponse) ProtoMessage() {}

func (x *RegisterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterResponse.ProtoReflect.Descriptor instead.
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// done
type ServerLoad struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          string                 `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	CurrentLoad   float32                `protobuf:"fixed32,3,opt,name=current_load,json=currentLoad,proto3" json:"current_load,omitempty"` // Based upon CPU Usage
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerLoad) Reset() {
	*x = ServerLoad{}
	mi := &file_protofiles_services_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerLoad) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerLoad) ProtoMessage() {}

func (x *ServerLoad) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerLoad.ProtoReflect.Descriptor instead.
func (*ServerLoad) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{2}
}

func (x *ServerLoad) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ServerLoad) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *ServerLoad) GetCurrentLoad() float32 {
	if x != nil {
		return x.CurrentLoad
	}
	return 0
}

type UpdateResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	mi := &file_protofiles_services_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// done
type ClientRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ClientRequest) Reset() {
	*x = ClientRequest{}
	mi := &file_protofiles_services_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ClientRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientRequest) ProtoMessage() {}

func (x *ClientRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientRequest.ProtoReflect.Descriptor instead.
func (*ClientRequest) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{4}
}

type ServerResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port          string                 `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ServerResponse) Reset() {
	*x = ServerResponse{}
	mi := &file_protofiles_services_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerResponse) ProtoMessage() {}

func (x *ServerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerResponse.ProtoReflect.Descriptor instead.
func (*ServerResponse) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{5}
}

func (x *ServerResponse) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ServerResponse) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

// done
type ComputeTask struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	A             int32                  `protobuf:"varint,1,opt,name=a,proto3" json:"a,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ComputeTask) Reset() {
	*x = ComputeTask{}
	mi := &file_protofiles_services_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ComputeTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComputeTask) ProtoMessage() {}

func (x *ComputeTask) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComputeTask.ProtoReflect.Descriptor instead.
func (*ComputeTask) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{6}
}

func (x *ComputeTask) GetA() int32 {
	if x != nil {
		return x.A
	}
	return 0
}

type ComputeResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Result        int64                  `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ComputeResponse) Reset() {
	*x = ComputeResponse{}
	mi := &file_protofiles_services_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ComputeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComputeResponse) ProtoMessage() {}

func (x *ComputeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protofiles_services_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComputeResponse.ProtoReflect.Descriptor instead.
func (*ComputeResponse) Descriptor() ([]byte, []int) {
	return file_protofiles_services_proto_rawDescGZIP(), []int{7}
}

func (x *ComputeResponse) GetResult() int64 {
	if x != nil {
		return x.Result
	}
	return 0
}

var File_protofiles_services_proto protoreflect.FileDescriptor

var file_protofiles_services_proto_rawDesc = string([]byte{
	0x0a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3a, 0x0a, 0x0a, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x2c, 0x0a, 0x10, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x5d, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4c,
	0x6f, 0x61, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x6c, 0x6f, 0x61,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0b, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x4c, 0x6f, 0x61, 0x64, 0x22, 0x2a, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x22, 0x0f, 0x0a, 0x0d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x3e, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x22, 0x1b, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b,
	0x12, 0x0c, 0x0a, 0x01, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x61, 0x22, 0x29,
	0x0a, 0x0f, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0x9e, 0x01, 0x0a, 0x0c, 0x4c, 0x6f,
	0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x12, 0x30, 0x0a, 0x0e, 0x52, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0b, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x11, 0x2e, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x0a,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4c, 0x6f, 0x61, 0x64, 0x12, 0x0b, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x4c, 0x6f, 0x61, 0x64, 0x1a, 0x0f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x42,
	0x65, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0e, 0x2e, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x3b, 0x0a, 0x07, 0x42, 0x61,
	0x63, 0x6b, 0x65, 0x6e, 0x64, 0x12, 0x30, 0x0a, 0x0e, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0c, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74,
	0x65, 0x54, 0x61, 0x73, 0x6b, 0x1a, 0x10, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x14, 0x5a, 0x12, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_protofiles_services_proto_rawDescOnce sync.Once
	file_protofiles_services_proto_rawDescData []byte
)

func file_protofiles_services_proto_rawDescGZIP() []byte {
	file_protofiles_services_proto_rawDescOnce.Do(func() {
		file_protofiles_services_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_protofiles_services_proto_rawDesc), len(file_protofiles_services_proto_rawDesc)))
	})
	return file_protofiles_services_proto_rawDescData
}

var file_protofiles_services_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_protofiles_services_proto_goTypes = []any{
	(*ServerInfo)(nil),       // 0: ServerInfo
	(*RegisterResponse)(nil), // 1: RegisterResponse
	(*ServerLoad)(nil),       // 2: ServerLoad
	(*UpdateResponse)(nil),   // 3: UpdateResponse
	(*ClientRequest)(nil),    // 4: ClientRequest
	(*ServerResponse)(nil),   // 5: ServerResponse
	(*ComputeTask)(nil),      // 6: ComputeTask
	(*ComputeResponse)(nil),  // 7: ComputeResponse
}
var file_protofiles_services_proto_depIdxs = []int32{
	0, // 0: LoadBalancer.RegisterServer:input_type -> ServerInfo
	2, // 1: LoadBalancer.UpdateLoad:input_type -> ServerLoad
	4, // 2: LoadBalancer.GetBestServer:input_type -> ClientRequest
	6, // 3: Backend.ProcessRequest:input_type -> ComputeTask
	1, // 4: LoadBalancer.RegisterServer:output_type -> RegisterResponse
	3, // 5: LoadBalancer.UpdateLoad:output_type -> UpdateResponse
	5, // 6: LoadBalancer.GetBestServer:output_type -> ServerResponse
	7, // 7: Backend.ProcessRequest:output_type -> ComputeResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protofiles_services_proto_init() }
func file_protofiles_services_proto_init() {
	if File_protofiles_services_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_protofiles_services_proto_rawDesc), len(file_protofiles_services_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_protofiles_services_proto_goTypes,
		DependencyIndexes: file_protofiles_services_proto_depIdxs,
		MessageInfos:      file_protofiles_services_proto_msgTypes,
	}.Build()
	File_protofiles_services_proto = out.File
	file_protofiles_services_proto_goTypes = nil
	file_protofiles_services_proto_depIdxs = nil
}
