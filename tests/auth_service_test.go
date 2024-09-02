package tests

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/auth"
	suite "github.com/tizzhh/micro-banking/tests/suite/auth"
)

const (
	jwtSecret = "test-secret"

	namesLen = 5

	passwordDefaultLen = 10

	emptyBalance = uint64(0)
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()
	firstName := gofakeit.LetterN(namesLen)
	lastName := gofakeit.LetterN(namesLen)
	age := uint32(gofakeit.Number(18, 100))

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())
	assert.Equal(t, email, respReg.GetEmail())

	respLogin, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), uint64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))

	assert.Equal(t, float64(st.Cfg.TokenTTL), claims["exp"].(float64))

	respUser, err := st.AuthClient.User(ctx, &authv1.UserRequest{
		Email: email,
	})
	require.NoError(t, err)
	assert.Equal(t, email, respUser.GetEmail())
	assert.Equal(t, firstName, respUser.GetFirstName())
	assert.Equal(t, lastName, respUser.GetLastName())
	assert.Equal(t, age, respUser.GetAge())
	assert.Equal(t, emptyBalance, respUser.GetBalance())
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()
	firstName := gofakeit.LetterN(namesLen)
	lastName := gofakeit.LetterN(namesLen)
	age := gofakeit.Number(18, 100)

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       uint32(age),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       uint32(age),
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		firstName   string
		lastName    string
		age         uint32
		expectedErr string
	}{
		{
			name:        "Register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			firstName:   gofakeit.LetterN(namesLen),
			lastName:    gofakeit.LetterN(namesLen),
			age:         uint32(gofakeit.Number(18, 100)),
			expectedErr: "password: value length must be at least 5 characters",
		},
		{
			name:        "Register with empty email",
			email:       "",
			password:    randomFakePassword(),
			firstName:   gofakeit.LetterN(namesLen),
			lastName:    gofakeit.LetterN(namesLen),
			age:         uint32(gofakeit.Number(18, 100)),
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Register with empty first name",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			firstName:   "",
			lastName:    gofakeit.LetterN(namesLen),
			age:         uint32(gofakeit.Number(18, 100)),
			expectedErr: "first_name: value length must be at least 2 characters",
		},
		{
			name:        "Register with empty last name",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			firstName:   gofakeit.LetterN(namesLen),
			lastName:    "",
			age:         uint32(gofakeit.Number(18, 100)),
			expectedErr: "last_name: value length must be at least 2 characters",
		},
		{
			name:        "Register 0 age",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			firstName:   gofakeit.LetterN(namesLen),
			lastName:    gofakeit.LetterN(namesLen),
			age:         0,
			expectedErr: "age: value must be greater than or equal to 18 and less than 150",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
				Email:     tt.email,
				Password:  tt.password,
				FirstName: tt.firstName,
				LastName:  tt.lastName,
				Age:       tt.age,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Login with empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password: value length must be at least 5 characters",
		},
		{
			name:        "Login with empty email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Login with empty email and password",
			email:       "",
			password:    "",
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Login with non-matching password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			expectedErr: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestUpdatePasswordDelete_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()
	firstName := gofakeit.LetterN(namesLen)
	lastName := gofakeit.LetterN(namesLen)
	age := gofakeit.Number(18, 100)

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       uint32(age),
	})

	require.NoError(t, err)

	respLogin, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	newPassord := randomFakePassword()
	respUpdate, err := st.AuthClient.UpdatePassword(ctx, &authv1.UpdatePasswordRequest{
		Email:       email,
		OldPassword: password,
		NewPassword: newPassord,
	})

	require.NoError(t, err)
	assert.Equal(t, email, respUpdate.GetEmail())

	respDelete, err := st.AuthClient.Unregister(ctx, &authv1.UnregisterRequest{
		Email:    email,
		Password: newPassord,
	})
	require.NoError(t, err)
	assert.Equal(t, email, respDelete.GetEmail())

	_, err = st.AuthClient.User(ctx, &authv1.UserRequest{
		Email: email,
	})
	assert.ErrorContains(t, err, "user not found")
}

func TestUpdatePassword_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	password := randomFakePassword()
	firstName := gofakeit.LetterN(namesLen)
	lastName := gofakeit.LetterN(namesLen)
	age := gofakeit.Number(18, 100)

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       uint32(age),
	})

	require.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		oldPassword string
		newPassword string
		expectedErr string
	}{
		{
			name:        "Update password with empty old password",
			email:       gofakeit.Email(),
			oldPassword: "",
			newPassword: randomFakePassword(),
			expectedErr: "password: value length must be at least 5 characters",
		},
		{
			name:        "Update password with empty email",
			email:       "",
			oldPassword: randomFakePassword(),
			newPassword: randomFakePassword(),
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Update password with empty new password",
			email:       gofakeit.Email(),
			oldPassword: randomFakePassword(),
			newPassword: "",
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Update password with empty email and password",
			email:       "",
			oldPassword: "",
			newPassword: randomFakePassword(),
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Update password with non-matching password",
			email:       gofakeit.Email(),
			oldPassword: randomFakePassword(),
			newPassword: randomFakePassword(),
			expectedErr: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.AuthClient.UpdatePassword(ctx, &authv1.UpdatePasswordRequest{
				Email:       tt.email,
				OldPassword: tt.oldPassword,
				NewPassword: tt.newPassword,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestDelete_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	password := randomFakePassword()
	firstName := gofakeit.LetterN(namesLen)
	lastName := gofakeit.LetterN(namesLen)
	age := gofakeit.Number(18, 100)

	_, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Age:       uint32(age),
	})

	require.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Delete user with empty old password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password: value length must be at least 5 characters",
		},
		{
			name:        "Delete user with empty email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Delete user with empty email and password",
			email:       "",
			password:    "",
			expectedErr: "email: value is empty, which is not a valid email address",
		},
		{
			name:        "Delete user with non-matching password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			expectedErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := st.AuthClient.Unregister(ctx, &authv1.UnregisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passwordDefaultLen)
}
