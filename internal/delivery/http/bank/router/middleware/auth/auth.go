package auth

import (
	"bytes"
	"errors"

	"io"
	"log/slog"
	"net/http"

	"strings"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tizzhh/micro-banking/internal/api/response"
	"github.com/tizzhh/micro-banking/internal/api/validate"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type PermissionsChecker interface {
	CheckPermissions(token string) (*jwt.Token, error)
}

type EmailRequest struct {
	Email string `json:"email"`
}

var (
	ErrMissingIDUrlParam = errors.New("missing id in url params")
	ErrInvalidToken      = errors.New("invalid token")
	ErrMissingUIDToken   = errors.New("missing uid in token")
	ErrInvalidUID        = errors.New("invalid uid")
)

var (
	ErrMissingEmailToken = errors.New("missing email in token")
)

func AuthenticateUser(log *slog.Logger, permissionsChecker PermissionsChecker) func(next http.Handler) http.Handler {
	const caller = "bank.middleware.auth.AuthenticateUser"
	log = sl.AddCaller(log, caller)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			log.Info("authenticating request", slog.String("auth header", reqToken))
			splitToken := strings.Split(reqToken, "Bearer")
			if len(splitToken) != 2 {
				log.Warn("invalid token")
				response.RespondWithError(w, r, "invalid token", http.StatusUnauthorized)
				return
			}
			reqToken = strings.TrimSpace(splitToken[1])

			log.Info("token in req", slog.String("token", reqToken))

			err := checkEmail(reqToken, r, permissionsChecker)
			if err != nil {
				log.Info("user does not have permissions", slog.String("token", reqToken), sl.Error(err))
				handleIncorrectEmail(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func checkEmail(token string, req *http.Request, permissionsChecker PermissionsChecker) error {
	var emailInRequest EmailRequest

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body.Close()
	// reset body
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	err = render.DecodeJSON(bytes.NewBuffer(bodyBytes), &emailInRequest)
	if errors.Is(err, io.EOF) {
		return validate.ErrEmptyBody
	}
	if err != nil {
		return validate.ErrDecodeFail
	}

	parsedToken, err := permissionsChecker.CheckPermissions(token)
	if err != nil {
		return ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidToken
	}

	idFromToken, ok := claims["email"].(string)
	if !ok {
		return ErrMissingEmailToken
	}

	if idFromToken != emailInRequest.Email {
		return ErrInvalidToken
	}

	return nil
}

func handleIncorrectEmail(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, validate.ErrEmptyBody) {
		response.RespondWithError(w, r, validate.ErrEmptyBody.Error(), http.StatusBadRequest)
	} else if errors.Is(err, validate.ErrDecodeFail) {
		response.RespondWithError(w, r, validate.ErrDecodeFail.Error(), http.StatusBadRequest)
	} else if errors.Is(err, ErrInvalidToken) {
		response.RespondWithError(w, r, ErrInvalidToken.Error(), http.StatusBadRequest)
	} else if errors.Is(err, ErrMissingEmailToken) {
		response.RespondWithError(w, r, ErrMissingEmailToken.Error(), http.StatusBadRequest)
	} else {
		response.RespondWithError(w, r, "internal error", http.StatusInternalServerError)
	}
}
