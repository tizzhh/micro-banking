package validate

import (
	"errors"
	"io"
	"log/slog"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

var (
	ErrEmptyBody            = errors.New("request body is empty")
	ErrDecodeFail           = errors.New("failed to decode request body")
	ErrValidationErrorsFail = errors.New("failed to get validation errors")
)

func ValidateRequest(log *slog.Logger, request any, body io.ReadCloser) error {
	const caller = "api.validate.ValidateRequest"
	log = sl.AddCaller(log, caller)
	log.Info("validating request")

	err := render.DecodeJSON(body, &request)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty", sl.Error(err))
		return ErrEmptyBody
	}
	if err != nil {
		log.Error("failed to decode request body", sl.Error(err))
		return ErrDecodeFail
	}

	log.Info("request body decoded", slog.Any("request", request))

	err = validator.New().Struct(request)
	if err != nil {
		err, ok := err.(validator.ValidationErrors)
		if !ok {
			log.Error("failed to get validation errors")
			return ErrValidationErrorsFail
		}
		log.Error("invalid request", sl.Error(err))
		return err
	}
	return nil
}
