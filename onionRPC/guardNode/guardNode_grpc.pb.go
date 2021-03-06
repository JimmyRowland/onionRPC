// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package guardNode

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GuardNodeServiceClient is the client API for GuardNodeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GuardNodeServiceClient interface {
	ExchangePublicKey(ctx context.Context, in *PublicKey, opts ...grpc.CallOption) (*PublicKey, error)
	ForwardRequest(ctx context.Context, in *ReqEncrypted, opts ...grpc.CallOption) (*ResEncrypted, error)
}

type guardNodeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGuardNodeServiceClient(cc grpc.ClientConnInterface) GuardNodeServiceClient {
	return &guardNodeServiceClient{cc}
}

func (c *guardNodeServiceClient) ExchangePublicKey(ctx context.Context, in *PublicKey, opts ...grpc.CallOption) (*PublicKey, error) {
	out := new(PublicKey)
	err := c.cc.Invoke(ctx, "/guardNode.GuardNodeService/ExchangePublicKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardNodeServiceClient) ForwardRequest(ctx context.Context, in *ReqEncrypted, opts ...grpc.CallOption) (*ResEncrypted, error) {
	out := new(ResEncrypted)
	err := c.cc.Invoke(ctx, "/guardNode.GuardNodeService/ForwardRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GuardNodeServiceServer is the server API for GuardNodeService service.
// All implementations must embed UnimplementedGuardNodeServiceServer
// for forward compatibility
type GuardNodeServiceServer interface {
	ExchangePublicKey(context.Context, *PublicKey) (*PublicKey, error)
	ForwardRequest(context.Context, *ReqEncrypted) (*ResEncrypted, error)
	mustEmbedUnimplementedGuardNodeServiceServer()
}

// UnimplementedGuardNodeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGuardNodeServiceServer struct {
}

func (UnimplementedGuardNodeServiceServer) ExchangePublicKey(context.Context, *PublicKey) (*PublicKey, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExchangePublicKey not implemented")
}
func (UnimplementedGuardNodeServiceServer) ForwardRequest(context.Context, *ReqEncrypted) (*ResEncrypted, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ForwardRequest not implemented")
}
func (UnimplementedGuardNodeServiceServer) mustEmbedUnimplementedGuardNodeServiceServer() {}

// UnsafeGuardNodeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GuardNodeServiceServer will
// result in compilation errors.
type UnsafeGuardNodeServiceServer interface {
	mustEmbedUnimplementedGuardNodeServiceServer()
}

func RegisterGuardNodeServiceServer(s grpc.ServiceRegistrar, srv GuardNodeServiceServer) {
	s.RegisterService(&GuardNodeService_ServiceDesc, srv)
}

func _GuardNodeService_ExchangePublicKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublicKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardNodeServiceServer).ExchangePublicKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/guardNode.GuardNodeService/ExchangePublicKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardNodeServiceServer).ExchangePublicKey(ctx, req.(*PublicKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardNodeService_ForwardRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReqEncrypted)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardNodeServiceServer).ForwardRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/guardNode.GuardNodeService/ForwardRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardNodeServiceServer).ForwardRequest(ctx, req.(*ReqEncrypted))
	}
	return interceptor(ctx, in, info, handler)
}

// GuardNodeService_ServiceDesc is the grpc.ServiceDesc for GuardNodeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GuardNodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "guardNode.GuardNodeService",
	HandlerType: (*GuardNodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExchangePublicKey",
			Handler:    _GuardNodeService_ExchangePublicKey_Handler,
		},
		{
			MethodName: "ForwardRequest",
			Handler:    _GuardNodeService_ForwardRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "guardNode/guardNode.proto",
}
