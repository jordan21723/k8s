package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strings"
)

func CreateJWT(claims jwt.Claims, signedString, signingMethodName string) (string, error) {
	signingMethod, err := GetJWTSignMethod(signingMethodName)
	if err != nil {
		return "", err
	}
	unSignedToken := jwt.NewWithClaims(signingMethod, claims)
	if signedToken, err := unSignedToken.SignedString([]byte(signedString)); err != nil {
		return "", err
	} else {
		return signedToken, nil
	}
}

func GetJWTSignMethod(signingMethodName string) (jwt.SigningMethod, error) {
	validMethodList := []string{
		"HS384",
		"HS256",
		"HS512",
		"ES256",
		"ECDSA",
		"ES512",
		"ES384",
		"PS256",
		"PS384",
		"PS512",
		"RS256",
		"RS384",
		"RS512",
	}
	signingMethod := jwt.GetSigningMethod(signingMethodName)
	if signingMethod == nil {
		return nil, errors.New(fmt.Sprintf("Signing Method %s is not valid method. Set signing method to one of following %s", signingMethodName, strings.Join(validMethodList, ",")))
	}
	return signingMethod, nil
}

func ValidateJWT(token string, claims jwt.Claims, signedString string) error {
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(signedString), nil
	})
	if err != nil {
		return err
	}
	return nil
}
