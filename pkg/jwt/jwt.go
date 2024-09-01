package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
)

type JWT struct {
	tokenTTL time.Duration
}

func New(tokenTTL time.Duration) *JWT {
	return &JWT{tokenTTL: tokenTTL}
}

func (j *JWT) NewToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":   user.ID,
			"email": user.Email,
			"exp":   j.tokenTTL,
		},
	)

	secretKey := os.Getenv("SECRET-KEY")

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) CheckToken(token string) (*jwt.Token, error) {
	secretKey := os.Getenv("SECRET-KEY")

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}
