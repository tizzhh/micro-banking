package grpc

import (
	"context"
	"fmt"

	authv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/auth"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/auth"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

func (c *Client) Register(ctx context.Context, email string, password string, firstName string, lastName string, age uint32) (uint64, error) {
	const caller = "clients.auth.grpc.Register"
	log := sl.AddCaller(c.log, caller)
	log.Info("registerting user")
	resp, err := c.api.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
	})
	if err != nil {
		log.Error("failed to register user", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	return resp.GetUserId(), nil
}

func (c *Client) Login(ctx context.Context, email string, password string) (string, error) {
	const caller = "clients.auth.grpc.Login"
	log := sl.AddCaller(c.log, caller)
	log.Info("getting token for user")
	resp, err := c.api.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		log.Error("failed to get token", sl.Error(err))
		return "", fmt.Errorf("%s: %w", caller, err)
	}

	return resp.GetToken(), nil
}

func (c *Client) UpdatePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
	const caller = "clients.auth.grpc.UpdatePassword"
	log := sl.AddCaller(c.log, caller)
	log.Info("updating user's password")
	_, err := c.api.UpdatePassword(ctx, &authv1.UpdatePasswordRequest{
		Email:       email,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	if err != nil {
		log.Error("failed to update password", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}
	return nil
}

func (c *Client) Unregister(ctx context.Context, email string, password string) error {
	const caller = "clients.auth.grpc.Unregister"
	log := sl.AddCaller(c.log, caller)
	log.Info("deleting user")
	_, err := c.api.Unregister(ctx, &authv1.UnregisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		log.Error("failed to delete user", sl.Error(err))
		return fmt.Errorf("%s: %w", caller, err)
	}
	return nil
}

func (c *Client) User(ctx context.Context, email string) (auth.UserResponse, error) {
	const caller = "clients.auth.grpc.User"
	log := sl.AddCaller(c.log, caller)
	log.Info("getting user")
	resp, err := c.api.User(ctx, &authv1.UserRequest{
		Email: email,
	})
	if err != nil {
		log.Error("failed to get user", sl.Error(err))
		return auth.UserResponse{}, fmt.Errorf("%s: %w", caller, err)
	}
	return auth.UserResponse{
		Email:     resp.GetEmail(),
		FirstName: resp.GetFirstName(),
		LastName:  resp.GetLastName(),
		Balance:   resp.GetBalance(),
		Age:       resp.GetAge(),
	}, nil
}
