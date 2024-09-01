package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tizzhh/micro-banking/internal/config"
	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
)

type JWT struct {
	tokenTTL  time.Duration
	secretKey string
}

func New(tokenTTL time.Duration) *JWT {
	cfg := config.Get()
	return &JWT{
		tokenTTL:  tokenTTL,
		secretKey: cfg.SecretKey,
	}
}

func (j *JWT) NewToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":   user.ID,
			"email": user.Email,
			"exp":   j.tokenTTL,
		},
	)

	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) CheckToken(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}
