// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.0
// source: server/server.proto

package server

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

// ServerServiceClient is the client API for ServerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServerServiceClient interface {
	GetRandomNumber(ctx context.Context, in *RandomNumber, opts ...grpc.CallOption) (*RandomNumber, error)
}

type serverServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewServerServiceClient(cc grpc.ClientConnInterface) ServerServiceClient {
	return &serverServiceClient{cc}
}

func (c *serverServiceClient) GetRandomNumber(ctx context.Context, in *RandomNumber, opts ...grpc.CallOption) (*RandomNumber, error) {
	out := new(RandomNumber)
	err := c.cc.Invoke(ctx, "/server.ServerService/GetRandomNumber", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerServiceServer is the server API for ServerService service.
// All implementations must embed UnimplementedServerServiceServer
// for forward compatibility
type ServerServiceServer interface {
	GetRandomNumber(context.Context, *RandomNumber) (*RandomNumber, error)
	mustEmbedUnimplementedServerServiceServer()
}

// UnimplementedServerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServerServiceServer struct {
}

func (UnimplementedServerServiceServer) GetRandomNumber(context.Context, *RandomNumber) (*RandomNumber, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRandomNumber not implemented")
}
func (UnimplementedServerServiceServer) mustEmbedUnimplementedServerServiceServer() {}

// UnsafeServerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServerServiceServer will
// result in compilation errors.
type UnsafeServerServiceServer interface {
	mustEmbedUnimplementedServerServiceServer()
}

func RegisterServerServiceServer(s grpc.ServiceRegistrar, srv ServerServiceServer) {
	s.RegisterService(&ServerService_ServiceDesc, srv)
}

func _ServerService_GetRandomNumber_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RandomNumber)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServiceServer).GetRandomNumber(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.ServerService/GetRandomNumber",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServiceServer).GetRandomNumber(ctx, req.(*RandomNumber))
	}
	return interceptor(ctx, in, info, handler)
}

// ServerService_ServiceDesc is the grpc.ServiceDesc for ServerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ServerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.ServerService",
	HandlerType: (*ServerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRandomNumber",
			Handler:    _ServerService_GetRandomNumber_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/server.proto",
}
