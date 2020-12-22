package common

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenSettings struct {
	SigningKey string
	Issuer string
	ExpiryDuration time.Duration
}

type RequestAuthenticatorFunc func(r *http.Request) (cifKey string, err error) 

type RequestAuthenticator struct {
	tokenSettings TokenSettings
}

func DefaultTokenSettings() TokenSettings {
	return TokenSettings {
		SigningKey: "thinmonkeysSecretSignature",
		Issuer: "thinmonkeys",
		ExpiryDuration: time.Duration(30) * time.Minute,
	}
}

func DefaultRequestAuthenticator() RequestAuthenticator {
	return RequestAuthenticator{
		tokenSettings: DefaultTokenSettings(),
	}
}

const (
	errorMessageMissingHeader string = "Missing x-auth-token header"
)

func (auth RequestAuthenticator) AuthenticateRequest(r *http.Request) (cifKey string, err error) {
	token := r.Header.Get("x-auth-token")
	if token == "" {
		return "", errors.New(errorMessageMissingHeader)
	}

	parsedToken, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
	
		return []byte(auth.tokenSettings.SigningKey), nil
	})
	
	if claims, ok := parsedToken.Claims.(*jwt.StandardClaims); ok && parsedToken.Valid {
		if claims.Issuer == auth.tokenSettings.Issuer {
			return claims.Subject, nil
		} else {
			return "", jwt.NewValidationError("Invalid issuer " + claims.Issuer, jwt.ValidationErrorIssuer)
		}
	}

	return "", err
}

func (auth RequestAuthenticator) AuthenticateRequestAllowingQueryOverride(r *http.Request) (cifKey string, err error) {
	cifs := r.URL.Query()["cif"]
	if len(cifs) == 1 {
		return cifs[0], nil
	}
	
	cifKey, err = auth.AuthenticateRequest(r)
	if err != nil && err.Error() == errorMessageMissingHeader {
		err = errors.New("Please either provide the Customer CIF in the querystring (?cif=), or provide an auth token in the x-auth-token header")
	}
	return
}
