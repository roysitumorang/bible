package middleware

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/bible/keys"
)

func BearerVerify(tokenString string) (*jwt.RegisteredClaims, error) {
	var claimsStruct jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(_ *jwt.Token) (interface{}, error) {
			return keys.InitPublicKey()
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid JWT")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("invalid JWT")
	}
	if claims.Issuer != os.Getenv("JWT_ISSUER") {
		return nil, errors.New("iss is invalid")
	}
	if len(claims.Audience) == 0 || claims.Audience[0] != os.Getenv("GOOGLE_API_CLIENT_ID") {
		return nil, errors.New("aud is invalid")
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("JWT is expired")
	}
	return claims, nil
}
