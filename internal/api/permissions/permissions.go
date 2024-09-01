package permissions

import "github.com/golang-jwt/jwt/v5"

type JWTChecker interface {
	CheckToken(token string) (*jwt.Token, error)
}

type PermissionsChecker struct {
	jwt JWTChecker
}

func New(jwt JWTChecker) *PermissionsChecker {
	return &PermissionsChecker{
		jwt: jwt,
	}
}

func (pc *PermissionsChecker) CheckPermissions(token string) (*jwt.Token, error) {
	parsedToken, err := pc.jwt.CheckToken(token)
	return parsedToken, err
}
