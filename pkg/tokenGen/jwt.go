package tokenGen

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

type UserInfoToken struct {
	ID    string `json:"id" mapstructure:"user_id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type jWTUserInfoClaims struct {
	jwt.RegisteredClaims
	User *UserInfoToken `json:"user,omitempty"`
}

func NewToken(secret string, expirationAt time.Time, userInfo *UserInfoToken) (string, error) {
	expiresIn := expirationAt.Sub(time.Now())
	claims := jWTUserInfoClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		userInfo,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))

	return tokenString, err
}

func VerifyToken(secret string, token string) (*UserInfoToken, bool) {
	t, err := jwt.ParseWithClaims(
		token,
		&jWTUserInfoClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	userInfo, ok := t.Claims.(*jWTUserInfoClaims)
	if !ok {
		return nil, false
	}

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return userInfo.User, false
		}
		return nil, false
	}

	expTime, err := t.Claims.GetExpirationTime()
	fmt.Println("expTime", expTime)
	if err != nil {
		return nil, false
	}

	if !t.Valid || expTime.Before(time.Now()) {
		return nil, false
	}

	return userInfo.User, true
}
