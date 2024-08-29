package auth

import (
	"context"
	"github.com/tizzhh/micro-banking/protos/gen/go/auth"
	"google.golang.org/grpc"
)

type serverApi struct {
	authv1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	authv1.RegisterAuthServer(gRPC, &serverApi{})
}

type Auth interface {
	Register(ctx context.Context, email string, password string, firstName string, lastName string, age int32) (int64, string, error)
	Login(ctx context.Context, email string, passwod string) (string, error)
	UpdatePassword(ctx context.Context, email string, oldPassword string, newPassword string) (string, error)
	Unregister(ctx context.Context, email string) (string, error)
}

func (s *serverApi) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	panic("unimplemented")
}

func (s *serverApi) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	panic("unimplemented")
}

func (s *serverApi) UpdatePassword(ctx context.Context, req *authv1.UpdatePasswordRequest) (*authv1.UpdatePasswordResponse, error) {
	panic("unimplemented")
}

func (s *serverApi) Unregister(ctx context.Context, req *authv1.UnregisterRequest) (*authv1.UnregisterResponse, error) {
	panic("unimplemented")
}
