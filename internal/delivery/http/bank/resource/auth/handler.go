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

//go:generate go run github.com/vektra/mockery/v2 --name=AuthClient
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

// NewUser godoc
// @Summary Register a new user
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param RegisterRequest body RegisterRequest true "Register Request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/register [post]
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

// Login godoc
// @Summary Login a user
// @Description Login a user and get token
// @Tags auth
// @Accept json
// @Produce json
// @Param LoginRequest body LoginRequest true "Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/login [post]
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

// UpdatePassword godoc
// @Summary Update user's password
// @Description Update user's password with new password
// @Tags auth
// @Accept json
// @Produce json
// @Param UpdatePasswordRequest body UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/change-password [put]
// @Security BearerAuth
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

// DeleteUser godoc
// @Summary Unregister user
// @Description Unregister user from the service.
// This action deletes corresponding data from the db
// @Tags auth
// @Accept json
// @Produce json
// @Param DeleteUserRequest body DeleteUserRequest true "Delete user Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/unregister [delete]
// @Security BearerAuth
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

// User godoc
// @Summary Returns user
// @Description Returns user
// @Tags auth
// @Accept json
// @Produce json
// @Param UserRequest body UserRequest true "User Request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /auth/user [get]
// @Security BearerAuth
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
