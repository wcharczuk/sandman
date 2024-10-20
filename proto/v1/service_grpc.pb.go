// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: proto/v1/service.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Timers_CreateTimer_FullMethodName  = "/v1.Timers/CreateTimer"
	Timers_ListTimers_FullMethodName   = "/v1.Timers/ListTimers"
	Timers_GetTimer_FullMethodName     = "/v1.Timers/GetTimer"
	Timers_DeleteTimer_FullMethodName  = "/v1.Timers/DeleteTimer"
	Timers_DeleteTimers_FullMethodName = "/v1.Timers/DeleteTimers"
)

// TimersClient is the client API for Timers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TimersClient interface {
	CreateTimer(ctx context.Context, in *Timer, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListTimers(ctx context.Context, in *ListTimersArgs, opts ...grpc.CallOption) (*ListTimersResponse, error)
	GetTimer(ctx context.Context, in *GetTimerArgs, opts ...grpc.CallOption) (*Timer, error)
	DeleteTimer(ctx context.Context, in *DeleteTimerArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteTimers(ctx context.Context, in *DeleteTimersArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type timersClient struct {
	cc grpc.ClientConnInterface
}

func NewTimersClient(cc grpc.ClientConnInterface) TimersClient {
	return &timersClient{cc}
}

func (c *timersClient) CreateTimer(ctx context.Context, in *Timer, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Timers_CreateTimer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timersClient) ListTimers(ctx context.Context, in *ListTimersArgs, opts ...grpc.CallOption) (*ListTimersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListTimersResponse)
	err := c.cc.Invoke(ctx, Timers_ListTimers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timersClient) GetTimer(ctx context.Context, in *GetTimerArgs, opts ...grpc.CallOption) (*Timer, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Timer)
	err := c.cc.Invoke(ctx, Timers_GetTimer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timersClient) DeleteTimer(ctx context.Context, in *DeleteTimerArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Timers_DeleteTimer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *timersClient) DeleteTimers(ctx context.Context, in *DeleteTimersArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Timers_DeleteTimers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TimersServer is the server API for Timers service.
// All implementations must embed UnimplementedTimersServer
// for forward compatibility.
type TimersServer interface {
	CreateTimer(context.Context, *Timer) (*emptypb.Empty, error)
	ListTimers(context.Context, *ListTimersArgs) (*ListTimersResponse, error)
	GetTimer(context.Context, *GetTimerArgs) (*Timer, error)
	DeleteTimer(context.Context, *DeleteTimerArgs) (*emptypb.Empty, error)
	DeleteTimers(context.Context, *DeleteTimersArgs) (*emptypb.Empty, error)
	mustEmbedUnimplementedTimersServer()
}

// UnimplementedTimersServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTimersServer struct{}

func (UnimplementedTimersServer) CreateTimer(context.Context, *Timer) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTimer not implemented")
}
func (UnimplementedTimersServer) ListTimers(context.Context, *ListTimersArgs) (*ListTimersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTimers not implemented")
}
func (UnimplementedTimersServer) GetTimer(context.Context, *GetTimerArgs) (*Timer, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTimer not implemented")
}
func (UnimplementedTimersServer) DeleteTimer(context.Context, *DeleteTimerArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTimer not implemented")
}
func (UnimplementedTimersServer) DeleteTimers(context.Context, *DeleteTimersArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTimers not implemented")
}
func (UnimplementedTimersServer) mustEmbedUnimplementedTimersServer() {}
func (UnimplementedTimersServer) testEmbeddedByValue()                {}

// UnsafeTimersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TimersServer will
// result in compilation errors.
type UnsafeTimersServer interface {
	mustEmbedUnimplementedTimersServer()
}

func RegisterTimersServer(s grpc.ServiceRegistrar, srv TimersServer) {
	// If the following call pancis, it indicates UnimplementedTimersServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Timers_ServiceDesc, srv)
}

func _Timers_CreateTimer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Timer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimersServer).CreateTimer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Timers_CreateTimer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimersServer).CreateTimer(ctx, req.(*Timer))
	}
	return interceptor(ctx, in, info, handler)
}

func _Timers_ListTimers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTimersArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimersServer).ListTimers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Timers_ListTimers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimersServer).ListTimers(ctx, req.(*ListTimersArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _Timers_GetTimer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTimerArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimersServer).GetTimer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Timers_GetTimer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimersServer).GetTimer(ctx, req.(*GetTimerArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _Timers_DeleteTimer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteTimerArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimersServer).DeleteTimer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Timers_DeleteTimer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimersServer).DeleteTimer(ctx, req.(*DeleteTimerArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _Timers_DeleteTimers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteTimersArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TimersServer).DeleteTimers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Timers_DeleteTimers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TimersServer).DeleteTimers(ctx, req.(*DeleteTimersArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// Timers_ServiceDesc is the grpc.ServiceDesc for Timers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Timers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.Timers",
	HandlerType: (*TimersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTimer",
			Handler:    _Timers_CreateTimer_Handler,
		},
		{
			MethodName: "ListTimers",
			Handler:    _Timers_ListTimers_Handler,
		},
		{
			MethodName: "GetTimer",
			Handler:    _Timers_GetTimer_Handler,
		},
		{
			MethodName: "DeleteTimer",
			Handler:    _Timers_DeleteTimer_Handler,
		},
		{
			MethodName: "DeleteTimers",
			Handler:    _Timers_DeleteTimers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/v1/service.proto",
}

const (
	Workers_ListWorkers_FullMethodName = "/v1.Workers/ListWorkers"
)

// WorkersClient is the client API for Workers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkersClient interface {
	ListWorkers(ctx context.Context, in *ListWorkersArgs, opts ...grpc.CallOption) (*ListWorkersResponse, error)
}

type workersClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkersClient(cc grpc.ClientConnInterface) WorkersClient {
	return &workersClient{cc}
}

func (c *workersClient) ListWorkers(ctx context.Context, in *ListWorkersArgs, opts ...grpc.CallOption) (*ListWorkersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListWorkersResponse)
	err := c.cc.Invoke(ctx, Workers_ListWorkers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkersServer is the server API for Workers service.
// All implementations must embed UnimplementedWorkersServer
// for forward compatibility.
type WorkersServer interface {
	ListWorkers(context.Context, *ListWorkersArgs) (*ListWorkersResponse, error)
	mustEmbedUnimplementedWorkersServer()
}

// UnimplementedWorkersServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWorkersServer struct{}

func (UnimplementedWorkersServer) ListWorkers(context.Context, *ListWorkersArgs) (*ListWorkersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWorkers not implemented")
}
func (UnimplementedWorkersServer) mustEmbedUnimplementedWorkersServer() {}
func (UnimplementedWorkersServer) testEmbeddedByValue()                 {}

// UnsafeWorkersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkersServer will
// result in compilation errors.
type UnsafeWorkersServer interface {
	mustEmbedUnimplementedWorkersServer()
}

func RegisterWorkersServer(s grpc.ServiceRegistrar, srv WorkersServer) {
	// If the following call pancis, it indicates UnimplementedWorkersServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Workers_ServiceDesc, srv)
}

func _Workers_ListWorkers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListWorkersArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkersServer).ListWorkers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Workers_ListWorkers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkersServer).ListWorkers(ctx, req.(*ListWorkersArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// Workers_ServiceDesc is the grpc.ServiceDesc for Workers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Workers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.Workers",
	HandlerType: (*WorkersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListWorkers",
			Handler:    _Workers_ListWorkers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/v1/service.proto",
}
