package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	authApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/auth"
	authMocks "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/auth/mocks"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

const (
	registerRequestTemplate = `{"email": "%s",
		"password": "%s",
		"first_name": "%s",
		"last_name": "%s",
		"age": %d
	}`
	registerResponseTemplate = `{"user_id":%d,"email":"%s","first_name":"%s","last_name":"%s","balance":%d,"age":%d}`

	loginRequestTemplate  = `{"email": "%s","password": "%s"}`
	loginResponseTemplate = `{"token":"%s"}`

	updatePassRequestTemplate  = `{"email": "%s","new_password": "%s","old_password": "%s"}`
	updatePassResponseTemplate = `{"message":"%s"}`

	deleteRequestTemplate  = `{"email": "%s","password": "%s"}`
	deleteResponseTemplate = `{"message":"%s"}`

	userRequestTemplate  = `{"email": "%s"}`
	userResponseTemplate = `{"user_id":%d,"email":"%s","first_name":"%s","last_name":"%s","balance":%d,"age":%d}`
)

var (
	log        = sl.Get()
	validation = validator.New()
)

func TestRegister_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "test"
	var testUserAge uint32 = 18
	testUserBalance := 0
	var testUserId uint64 = 0

	reqBody := []byte(fmt.Sprintf(
		registerRequestTemplate,
		testUserEmail,
		testUserName,
		testUserName,
		testUserName,
		testUserAge,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/auth/register", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"Register",
		context.Background(),
		testUserEmail,
		testUserName,
		testUserName,
		testUserName,
		testUserAge,
	).Return(uint64(0), nil)
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.NewUser())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		registerResponseTemplate,
		testUserId,
		testUserEmail,
		testUserName,
		testUserName,
		testUserBalance,
		testUserAge,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestLogin_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "test"
	testToken := "token"

	reqBody := []byte(fmt.Sprintf(
		loginRequestTemplate,
		testUserEmail,
		testUserName,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/auth/login", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"Login",
		context.Background(),
		testUserEmail,
		testUserName,
	).Return(testToken, nil)
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginUser())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		loginResponseTemplate,
		testToken,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestUpdatePassword_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserOldPassword := "test"
	testUserNewPassword := "test1"

	testResponse := "Password updated successfully"

	reqBody := []byte(fmt.Sprintf(
		updatePassRequestTemplate,
		testUserEmail,
		testUserOldPassword,
		testUserNewPassword,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPut, "/auth/change-password", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"UpdatePassword",
		context.Background(),
		testUserEmail,
		testUserNewPassword,
		testUserOldPassword,
	).Return(nil)
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.UpdatePassword())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		updatePassResponseTemplate,
		testResponse,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestDeletePassword_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserPassword := "test"

	testResponse := "User deleted successfully"

	reqBody := []byte(fmt.Sprintf(
		deleteRequestTemplate,
		testUserEmail,
		testUserPassword,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodDelete, "/auth/unregister", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"Unregister",
		context.Background(),
		testUserEmail,
		testUserPassword,
	).Return(nil)
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.DeleteUser())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		updatePassResponseTemplate,
		testResponse,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestUser_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "test"
	var testUserAge uint32 = 18
	testUserBalance := 0
	var testUserId uint64 = 0

	reqBody := []byte(fmt.Sprintf(
		userRequestTemplate,
		testUserEmail,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodGet, "/auth/user", bodyReader)
	require.NoError(t, err)

	exceptedResponse := authApi.UserResponse{
		ID:        testUserId,
		Email:     testUserEmail,
		FirstName: testUserName,
		LastName:  testUserName,
		Balance:   uint64(testUserBalance),
		Age:       testUserAge,
	}

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"User",
		context.Background(),
		testUserEmail,
	).Return(exceptedResponse, nil)
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.User())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		userResponseTemplate,
		testUserId,
		testUserEmail,
		testUserName,
		testUserName,
		testUserBalance,
		testUserAge,
	), strings.TrimRight(rr.Body.String(), "\n"))
}
