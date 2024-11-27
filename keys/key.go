package keys

import (
	"crypto/rsa"
	"encoding/base64"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func InitPublicKey() (*rsa.PublicKey, error) {
	verifyBytes, err := base64.StdEncoding.DecodeString(os.Getenv("RSA_PUBLIC_KEY"))
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
}

func InitPrivateKey() (*rsa.PrivateKey, error) {
	signBytes, err := base64.StdEncoding.DecodeString(os.Getenv("RSA_PRIVATE_KEY"))
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(signBytes)
}
