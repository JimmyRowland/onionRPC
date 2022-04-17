// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: exitNode/exitNode.proto

package exitNode

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

type PublicKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey []byte `protobuf:"bytes,1,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	Token     []byte `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *PublicKey) Reset() {
	*x = PublicKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_exitNode_exitNode_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublicKey) ProtoMessage() {}

func (x *PublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_exitNode_exitNode_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublicKey.ProtoReflect.Descriptor instead.
func (*PublicKey) Descriptor() ([]byte, []int) {
	return file_exitNode_exitNode_proto_rawDescGZIP(), []int{0}
}

func (x *PublicKey) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *PublicKey) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

type ReqEncrypted struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Encrypted []byte `protobuf:"bytes,1,opt,name=encrypted,proto3" json:"encrypted,omitempty"`
	SessionId string `protobuf:"bytes,2,opt,name=sessionId,proto3" json:"sessionId,omitempty"`
	Token     []byte `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *ReqEncrypted) Reset() {
	*x = ReqEncrypted{}
	if protoimpl.UnsafeEnabled {
		mi := &file_exitNode_exitNode_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqEncrypted) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqEncrypted) ProtoMessage() {}

func (x *ReqEncrypted) ProtoReflect() protoreflect.Message {
	mi := &file_exitNode_exitNode_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqEncrypted.ProtoReflect.Descriptor instead.
func (*ReqEncrypted) Descriptor() ([]byte, []int) {
	return file_exitNode_exitNode_proto_rawDescGZIP(), []int{1}
}

func (x *ReqEncrypted) GetEncrypted() []byte {
	if x != nil {
		return x.Encrypted
	}
	return nil
}

func (x *ReqEncrypted) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *ReqEncrypted) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

type ResEncrypted struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Encrypted []byte `protobuf:"bytes,1,opt,name=encrypted,proto3" json:"encrypted,omitempty"`
	Token     []byte `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *ResEncrypted) Reset() {
	*x = ResEncrypted{}
	if protoimpl.UnsafeEnabled {
		mi := &file_exitNode_exitNode_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResEncrypted) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResEncrypted) ProtoMessage() {}

func (x *ResEncrypted) ProtoReflect() protoreflect.Message {
	mi := &file_exitNode_exitNode_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResEncrypted.ProtoReflect.Descriptor instead.
func (*ResEncrypted) Descriptor() ([]byte, []int) {
	return file_exitNode_exitNode_proto_rawDescGZIP(), []int{2}
}

func (x *ResEncrypted) GetEncrypted() []byte {
	if x != nil {
		return x.Encrypted
	}
	return nil
}

func (x *ResEncrypted) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

var File_exitNode_exitNode_proto protoreflect.FileDescriptor

var file_exitNode_exitNode_proto_rawDesc = []byte{
	0x0a, 0x17, 0x65, 0x78, 0x69, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x2f, 0x65, 0x78, 0x69, 0x74, 0x4e,
	0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x65, 0x78, 0x69, 0x74, 0x4e,
	0x6f, 0x64, 0x65, 0x22, 0x3f, 0x0a, 0x09, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79,
	0x12, 0x1c, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x74,
	0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x60, 0x0a, 0x0c, 0x52, 0x65, 0x71, 0x45, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x42, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x45, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x65, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x32, 0x96, 0x01, 0x0a, 0x0f, 0x45,
	0x78, 0x69, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3f,
	0x0a, 0x11, 0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63,
	0x4b, 0x65, 0x79, 0x12, 0x13, 0x2e, 0x65, 0x78, 0x69, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x2e, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x1a, 0x13, 0x2e, 0x65, 0x78, 0x69, 0x74, 0x4e,
	0x6f, 0x64, 0x65, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x22, 0x00, 0x12,
	0x42, 0x0a, 0x0e, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x16, 0x2e, 0x65, 0x78, 0x69, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x2e, 0x52, 0x65, 0x71,
	0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x1a, 0x16, 0x2e, 0x65, 0x78, 0x69, 0x74,
	0x4e, 0x6f, 0x64, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65,
	0x64, 0x22, 0x00, 0x42, 0x25, 0x5a, 0x23, 0x63, 0x73, 0x2e, 0x75, 0x62, 0x63, 0x2e, 0x63, 0x61,
	0x2f, 0x63, 0x70, 0x73, 0x63, 0x34, 0x31, 0x36, 0x2f, 0x6f, 0x6e, 0x69, 0x6f, 0x6e, 0x52, 0x50,
	0x43, 0x2f, 0x65, 0x78, 0x69, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_exitNode_exitNode_proto_rawDescOnce sync.Once
	file_exitNode_exitNode_proto_rawDescData = file_exitNode_exitNode_proto_rawDesc
)

func file_exitNode_exitNode_proto_rawDescGZIP() []byte {
	file_exitNode_exitNode_proto_rawDescOnce.Do(func() {
		file_exitNode_exitNode_proto_rawDescData = protoimpl.X.CompressGZIP(file_exitNode_exitNode_proto_rawDescData)
	})
	return file_exitNode_exitNode_proto_rawDescData
}

var file_exitNode_exitNode_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_exitNode_exitNode_proto_goTypes = []interface{}{
	(*PublicKey)(nil),    // 0: exitNode.PublicKey
	(*ReqEncrypted)(nil), // 1: exitNode.ReqEncrypted
	(*ResEncrypted)(nil), // 2: exitNode.ResEncrypted
}
var file_exitNode_exitNode_proto_depIdxs = []int32{
	0, // 0: exitNode.ExitNodeService.ExchangePublicKey:input_type -> exitNode.PublicKey
	1, // 1: exitNode.ExitNodeService.ForwardRequest:input_type -> exitNode.ReqEncrypted
	0, // 2: exitNode.ExitNodeService.ExchangePublicKey:output_type -> exitNode.PublicKey
	2, // 3: exitNode.ExitNodeService.ForwardRequest:output_type -> exitNode.ResEncrypted
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_exitNode_exitNode_proto_init() }
func file_exitNode_exitNode_proto_init() {
	if File_exitNode_exitNode_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_exitNode_exitNode_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublicKey); i {
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
		file_exitNode_exitNode_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqEncrypted); i {
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
		file_exitNode_exitNode_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResEncrypted); i {
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
			RawDescriptor: file_exitNode_exitNode_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_exitNode_exitNode_proto_goTypes,
		DependencyIndexes: file_exitNode_exitNode_proto_depIdxs,
		MessageInfos:      file_exitNode_exitNode_proto_msgTypes,
	}.Build()
	File_exitNode_exitNode_proto = out.File
	file_exitNode_exitNode_proto_rawDesc = nil
	file_exitNode_exitNode_proto_goTypes = nil
	file_exitNode_exitNode_proto_depIdxs = nil
}
