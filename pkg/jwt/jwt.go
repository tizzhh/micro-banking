package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tizzhh/micro-banking/internal/domain/auth/models"
)

func NewToken(user models.User, tokenTTL time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":   user.ID,
			"email": user.Email,
			"exp":   tokenTTL,
		},
	)

	secretKey := os.Getenv("SECRET-KEY")

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}
