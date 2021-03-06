// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package coord

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

// CoordServiceClient is the client API for CoordService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CoordServiceClient interface {
	NodeJoin(ctx context.Context, in *NodeJoinRequest, opts ...grpc.CallOption) (*NodeJoinResponse, error)
	RequestChain(ctx context.Context, in *Chain, opts ...grpc.CallOption) (*Chain, error)
}

type coordServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCoordServiceClient(cc grpc.ClientConnInterface) CoordServiceClient {
	return &coordServiceClient{cc}
}

func (c *coordServiceClient) NodeJoin(ctx context.Context, in *NodeJoinRequest, opts ...grpc.CallOption) (*NodeJoinResponse, error) {
	out := new(NodeJoinResponse)
	err := c.cc.Invoke(ctx, "/coordNode.CoordService/NodeJoin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coordServiceClient) RequestChain(ctx context.Context, in *Chain, opts ...grpc.CallOption) (*Chain, error) {
	out := new(Chain)
	err := c.cc.Invoke(ctx, "/coordNode.CoordService/RequestChain", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CoordServiceServer is the server API for CoordService service.
// All implementations must embed UnimplementedCoordServiceServer
// for forward compatibility
type CoordServiceServer interface {
	NodeJoin(context.Context, *NodeJoinRequest) (*NodeJoinResponse, error)
	RequestChain(context.Context, *Chain) (*Chain, error)
	mustEmbedUnimplementedCoordServiceServer()
}

// UnimplementedCoordServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCoordServiceServer struct {
}

func (UnimplementedCoordServiceServer) NodeJoin(context.Context, *NodeJoinRequest) (*NodeJoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodeJoin not implemented")
}
func (UnimplementedCoordServiceServer) RequestChain(context.Context, *Chain) (*Chain, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestChain not implemented")
}
func (UnimplementedCoordServiceServer) mustEmbedUnimplementedCoordServiceServer() {}

// UnsafeCoordServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CoordServiceServer will
// result in compilation errors.
type UnsafeCoordServiceServer interface {
	mustEmbedUnimplementedCoordServiceServer()
}

func RegisterCoordServiceServer(s grpc.ServiceRegistrar, srv CoordServiceServer) {
	s.RegisterService(&CoordService_ServiceDesc, srv)
}

func _CoordService_NodeJoin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeJoinRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordServiceServer).NodeJoin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coordNode.CoordService/NodeJoin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordServiceServer).NodeJoin(ctx, req.(*NodeJoinRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CoordService_RequestChain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Chain)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordServiceServer).RequestChain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coordNode.CoordService/RequestChain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordServiceServer).RequestChain(ctx, req.(*Chain))
	}
	return interceptor(ctx, in, info, handler)
}

// CoordService_ServiceDesc is the grpc.ServiceDesc for CoordService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CoordService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "coordNode.CoordService",
	HandlerType: (*CoordServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NodeJoin",
			Handler:    _CoordService_NodeJoin_Handler,
		},
		{
			MethodName: "RequestChain",
			Handler:    _CoordService_RequestChain_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coord/coord.proto",
}
