package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/internal/api/response"
	"github.com/tizzhh/micro-banking/internal/api/validate"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/common"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type AuthAPI struct {
	log        *slog.Logger
	validator  *validator.Validate
	authClient AuthClient
}

type AuthClient interface {
	Register(ctx context.Context, email string, password string, firstName string, lastName string, age uint32) (uint64, error)
	Login(ctx context.Context, email string, password string) (string, error)
	UpdatePassword(ctx context.Context, email string, oldPassword string, newPassword string) error
	Unregister(ctx context.Context, email string, password string) error
	User(ctx context.Context, email string) (UserResponse, error)
}

func New(log *slog.Logger, validator *validator.Validate, authClient AuthClient) *AuthAPI {
	return &AuthAPI{
		log:        log,
		validator:  validator,
		authClient: authClient,
	}
}

func (aa *AuthAPI) NewUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.auth.handler.NewUser"
		log := sl.AddRequestId(sl.AddCaller(aa.log, caller), middleware.GetReqID(r.Context()))
		log.Info("registering New user")

		var registerRequest RegisterRequest

		err := validate.ValidateRequest(aa.log, &registerRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		newUserID, err := aa.authClient.Register(
			r.Context(),
			registerRequest.Email,
			registerRequest.Password,
			registerRequest.FirstName,
			registerRequest.LastName,
			registerRequest.Age,
		)
		if err != nil {
			log.Error("failed to register user", sl.Error(err))
			common.HandleGrpcError(aa.log, w, r, err)
			return
		}

		log.Info("user created")

		render.JSON(w, r, UserResponse{
			ID:        newUserID,
			Email:     registerRequest.Email,
			FirstName: registerRequest.FirstName,
			LastName:  registerRequest.LastName,
			Age:       registerRequest.Age,
		})
	}
}

func (aa *AuthAPI) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.auth.handler.LoginUser"
		log := sl.AddRequestId(sl.AddCaller(aa.log, caller), middleware.GetReqID(r.Context()))
		log.Info("loging the user in")

		var loginRequest LoginRequest

		err := validate.ValidateRequest(aa.log, &loginRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		token, err := aa.authClient.Login(
			r.Context(),
			loginRequest.Email,
			loginRequest.Password,
		)
		if err != nil {
			log.Error("failed to login user", sl.Error(err))
			common.HandleGrpcError(aa.log, w, r, err)
			return
		}

		log.Info("user created")

		render.JSON(w, r, LoginResponse{
			Token: token,
		})
	}
}

func (aa *AuthAPI) UpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.auth.handler.UpdatePassword"
		log := sl.AddRequestId(sl.AddCaller(aa.log, caller), middleware.GetReqID(r.Context()))
		log.Info("updating user password")

		var loginRequest UpdatePasswordRequest

		err := validate.ValidateRequest(aa.log, &loginRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		err = aa.authClient.UpdatePassword(
			r.Context(),
			loginRequest.Email,
			loginRequest.OldPassword,
			loginRequest.NewPassword,
		)
		if err != nil {
			log.Error("failed to update password", sl.Error(err))
			common.HandleGrpcError(aa.log, w, r, err)
			return
		}

		log.Info("password updated")

		response.ReponsdWithOK(w, r, "Password updated successfully", http.StatusOK)
	}
}

func (aa *AuthAPI) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.auth.handler.DeleteUser"
		log := sl.AddRequestId(sl.AddCaller(aa.log, caller), middleware.GetReqID(r.Context()))
		log.Info("deleting user")

		var loginRequest DeleteUserRequest

		err := validate.ValidateRequest(aa.log, &loginRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		err = aa.authClient.Unregister(
			r.Context(),
			loginRequest.Email,
			loginRequest.Password,
		)
		if err != nil {
			log.Error("failed to delete user", sl.Error(err))
			common.HandleGrpcError(aa.log, w, r, err)
			return
		}

		log.Info("user deleted")

		response.ReponsdWithOK(w, r, "User deleted successfully", http.StatusNoContent)
	}
}

func (aa *AuthAPI) User() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const caller = "bank.auth.handler.User"
		log := sl.AddRequestId(sl.AddCaller(aa.log, caller), middleware.GetReqID(r.Context()))
		log.Info("getting user")

		var loginRequest UserRequest

		err := validate.ValidateRequest(aa.log, &loginRequest, r.Body)
		if err != nil {
			log.Error("validation err", sl.Error(err))
			common.HandleValidationErr(w, r, err)
			return
		}

		user, err := aa.authClient.User(
			r.Context(),
			loginRequest.Email,
		)
		if err != nil {
			log.Error("failed to delete user", sl.Error(err))
			common.HandleGrpcError(aa.log, w, r, err)
			return
		}

		log.Info("user deleted")

		render.JSON(w, r, UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Balance:   user.Balance,
			Age:       user.Age,
		})
	}
}
