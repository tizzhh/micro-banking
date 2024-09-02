package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Message string `json:"message"`
}

type Error struct {
	Error string `json:"error"`
}

func ReponsdWithOK(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, Response{Message: message})
}

func RespondWithError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, Error{Error: message})
}
func ValidationError(w http.ResponseWriter, r *http.Request, errs validator.ValidationErrors) {
	errMsgs := make([]string, 0, len(errs))
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid email", err.Field()))
		case "alpha":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s should only cosist of alphabetic characters", err.Field()))
		case "gte":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s should be greater or equal to %v", err.Field(), err.Param()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	render.JSON(w, r, Error{
		Error: strings.Join(errMsgs, " "),
	})
}
