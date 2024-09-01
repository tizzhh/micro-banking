package common

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/internal/api/response"
	"github.com/tizzhh/micro-banking/internal/api/validate"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleValidationErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, validate.ErrEmptyBody) {
		response.RespondWithError(w, r, "request body is empty", http.StatusBadRequest)
	} else if errors.Is(err, validate.ErrDecodeFail) {
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
	} else if errors.Is(err, validate.ErrValidationErrorsFail) {
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
	} else {
		response.ValidationError(w, r, err.(validator.ValidationErrors))
	}
}

func HandleGrpcError(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	const caller = "bank.auth.handler.handleGrpcError"
	log = sl.AddCaller(log, caller)

	grpcErr, ok := status.FromError(err)
	if !ok {
		log.Error("could not parse error", sl.Error(err))
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
		return
	}
	switch grpcErr.Code() {
	case codes.Internal:
		response.RespondWithError(w, r, grpcErr.Message(), http.StatusInternalServerError)
	case codes.InvalidArgument:
		response.RespondWithError(w, r, grpcErr.Message(), http.StatusBadRequest)
	case codes.AlreadyExists:
		response.RespondWithError(w, r, grpcErr.Message(), http.StatusBadRequest)
	case codes.FailedPrecondition:
		response.RespondWithError(w, r, grpcErr.Message(), http.StatusBadRequest)
	default:
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
	}
}
