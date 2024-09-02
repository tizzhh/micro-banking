package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	errorResponseTemplate = `{"error":"%s"}`
)

var (
	log        = sl.Get()
	validation = validator.New()

	shortNameLen = 0
)

func TestRegister_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "testt"
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

func TestRegisterHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		firstName      string
		lastName       string
		age            uint32
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Register with empty password",
			email:          gofakeit.Email(),
			password:       "",
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            uint32(gofakeit.Number(18, 100)),
			expectedErr:    "field Password is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Register with empty email",
			email:          "",
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            uint32(gofakeit.Number(18, 100)),
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Register with empty first name",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      "",
			lastName:       gofakeit.LetterN(namesLen),
			age:            uint32(gofakeit.Number(18, 100)),
			expectedErr:    "field FirstName is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Register with empty last name",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       "",
			age:            uint32(gofakeit.Number(18, 100)),
			expectedErr:    "field LastName is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Register 0 age",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            0,
			expectedErr:    "field Age is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Invalid email",
			email:          "asjnbd",
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            20,
			expectedErr:    "field Email is not a valid email",
			expectedStatus: 400,
		},
		{
			name:           "Short password",
			email:          gofakeit.Email(),
			password:       gofakeit.LetterN(uint(shortNameLen)),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            20,
			expectedErr:    "field Password should be greater or equal to 5",
			expectedStatus: 400,
		},
		{
			name:           "Short first name",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(uint(shortNameLen)),
			lastName:       gofakeit.LetterN(namesLen),
			age:            20,
			expectedErr:    "field FirstName should be greater or equal to 5",
			expectedStatus: 400,
		},
		{
			name:           "Short last name name",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(uint(shortNameLen)),
			age:            20,
			expectedErr:    "field LastName should be greater or equal to 5",
			expectedStatus: 400,
		},
		{
			name:           "Small age",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            1,
			expectedErr:    "field Age should be greater or equal to 18",
			expectedStatus: 400,
		},
		{
			name:           "Large age",
			email:          gofakeit.Email(),
			password:       randomFakePassword(),
			firstName:      gofakeit.LetterN(namesLen),
			lastName:       gofakeit.LetterN(namesLen),
			age:            400,
			expectedErr:    "field Age is not valid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				registerRequestTemplate,
				tt.email,
				tt.password,
				tt.firstName,
				tt.lastName,
				tt.age,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/auth/register", bodyReader)
			require.NoError(t, err)

			mockClient := authMocks.NewAuthClient(t)
			auth := authApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(auth.NewUser())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))
		})
	}
}

func TestLogin_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "testt"
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

func TestLoginHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Login with empty password",
			email:          gofakeit.Email(),
			password:       "",
			expectedErr:    "field Password is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Login with empty email",
			email:          "",
			password:       randomFakePassword(),
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Login with empty email and password",
			email:          "",
			password:       "",
			expectedErr:    "field Email is a required field field Password is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Login with short password",
			email:          gofakeit.Email(),
			password:       "a",
			expectedErr:    "field Password should be greater or equal to 5",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reqBody := []byte(fmt.Sprintf(
				loginRequestTemplate,
				tt.email,
				tt.password,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/auth/login", bodyReader)
			require.NoError(t, err)

			mockClient := authMocks.NewAuthClient(t)
			auth := authApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(auth.LoginUser())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))
		})
	}
}

func TestLoginNonMatchingPassword_Fail(t *testing.T) {
	testingEmail := "test@gmail.com"
	testingPass := randomFakePassword()

	excpectedErr := "invalid credentials"

	reqBody := []byte(fmt.Sprintf(
		loginRequestTemplate,
		testingEmail,
		testingPass,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/auth/login", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"Login",
		context.Background(),
		testingEmail,
		testingPass,
	).Return("", status.Error(codes.InvalidArgument, "invalid credentials"))
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginUser())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		excpectedErr,
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

func TestUpdatePasswordHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		oldPassword    string
		newPassword    string
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Update password with empty old password",
			email:          gofakeit.Email(),
			oldPassword:    "",
			newPassword:    randomFakePassword(),
			expectedErr:    "field OldPassword is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Update password with empty email",
			email:          "",
			oldPassword:    randomFakePassword(),
			newPassword:    randomFakePassword(),
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Update password with empty new password",
			email:          gofakeit.Email(),
			oldPassword:    randomFakePassword(),
			newPassword:    "",
			expectedErr:    "field NewPassword is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Update password with empty email and password",
			email:          "",
			oldPassword:    "",
			newPassword:    randomFakePassword(),
			expectedErr:    "field Email is a required field field OldPassword is a required field",
			expectedStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqBody := []byte(fmt.Sprintf(
				updatePassRequestTemplate,
				tt.email,
				tt.newPassword,
				tt.oldPassword,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodPut, "/auth/change-password", bodyReader)
			require.NoError(t, err)

			mockClient := authMocks.NewAuthClient(t)
			auth := authApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(auth.UpdatePassword())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))
		})
	}
}

func TestUpdateNonMatchingOldPassword_Fail(t *testing.T) {
	testingEmail := "test@gmail.com"
	testingOldPass := randomFakePassword()
	testingNewPass := randomFakePassword()

	excpectedErr := "invalid credentials"

	reqBody := []byte(fmt.Sprintf(
		updatePassRequestTemplate,
		testingEmail,
		testingOldPass,
		testingNewPass,
	))
	bodyReader := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/auth/change-password", bodyReader)
	require.NoError(t, err)

	mockClient := authMocks.NewAuthClient(t)
	mockClient.On(
		"UpdatePassword",
		context.Background(),
		testingEmail,
		testingNewPass,
		testingOldPass,
	).Return(status.Error(codes.InvalidArgument, "invalid credentials"))
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.UpdatePassword())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		excpectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestDeletePassword_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserPassword := "admin"

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

func TestDeletePasswordHttp_FailCases(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "Delete user with empty old password",
			email:          gofakeit.Email(),
			password:       "",
			expectedErr:    "field Password is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Delete user with empty email",
			email:          "",
			password:       randomFakePassword(),
			expectedErr:    "field Email is a required field",
			expectedStatus: 400,
		},
		{
			name:           "Delete user with empty email and password",
			email:          "",
			password:       "",
			expectedErr:    "field Email is a required field field Password is a required field",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := []byte(fmt.Sprintf(
				deleteRequestTemplate,
				tt.email,
				tt.password,
			))
			bodyReader := bytes.NewBuffer(reqBody)

			req, err := http.NewRequest(http.MethodDelete, "/auth/unregister", bodyReader)
			require.NoError(t, err)

			mockClient := authMocks.NewAuthClient(t)
			auth := authApi.New(log, validation, mockClient)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(auth.DeleteUser())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, fmt.Sprintf(
				errorResponseTemplate,
				tt.expectedErr,
			), strings.TrimRight(rr.Body.String(), "\n"))

		})
	}

}

func TestDeletePasswordNonMatchingPassword_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserPassword := "admin"

	excpectedErr := "user not found"

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
	).Return(status.Error(codes.NotFound, "user not found"))
	auth := authApi.New(log, validation, mockClient)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.DeleteUser())

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, fmt.Sprintf(
		errorResponseTemplate,
		excpectedErr,
	), strings.TrimRight(rr.Body.String(), "\n"))
}

func TestUser_HappyPath(t *testing.T) {
	testUserEmail := "test-user0@gmail.com"
	testUserName := "testt"
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
