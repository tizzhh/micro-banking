package auth

import (
	"context"
	"errors"

	"github.com/bufbuild/protovalidate-go"
	authv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/auth"
	auth "github.com/tizzhh/micro-banking/internal/services/auth/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverApi struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

type Auth interface {
	Register(ctx context.Context, email string, password string, firstName string, lastName string, age int32) (int64, error)
	Login(ctx context.Context, email string, password string) (string, error)
	UpdatePassword(ctx context.Context, email string, oldPassword string, newPassword string) error
	Unregister(ctx context.Context, email string, password string) error
}

func (s *serverApi) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	newUserId, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetFirstName(), req.GetLastName(), req.GetAge())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RegisterResponse{
		UserId: newUserId,
		Email:  req.GetEmail(),
	}, nil
}

func (s *serverApi) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverApi) UpdatePassword(ctx context.Context, req *authv1.UpdatePasswordRequest) (*authv1.UpdatePasswordResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.UpdatePassword(ctx, req.GetEmail(), req.GetOldPassword(), req.NewPassword)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.UpdatePasswordResponse{
		Email: req.GetEmail(),
	}, nil
}

func (s *serverApi) Unregister(ctx context.Context, req *authv1.UnregisterRequest) (*authv1.UnregisterResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if err = validator.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.Unregister(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.UnregisterResponse{
		Email: req.GetEmail(),
	}, nil
}
