package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"../common"
	"../../respond"
	"github.com/dgrijalva/jwt-go"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	IsSuccess bool `json:"isSuccess"`
	CustomerCIF string `json:"customerCIF,omitempty"`
	AuthToken string `json:"authToken,omitempty"`
}

type LoginAuthenticator func (LoginRequest) (CustomerCIF string, err error)

type LoginHandler struct {
	loginAuthenticator LoginAuthenticator
	tokenSettings common.TokenSettings
	timeProvider func()(time.Time)
}

func NewHandler() LoginHandler {
	return LoginHandler {
		loginAuthenticator: dummyAuthenticator,
		tokenSettings: common.DefaultTokenSettings(),
		timeProvider: time.Now,
	}
}


func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	request, err, errorCode := parseRequest(r)
	if err != nil {
		respond.WithError(w, errorCode, err.Error())
		return
	}

	cif, err := h.loginAuthenticator(request)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if cif == "" {
		respond.WithJSON(w, http.StatusUnauthorized, LoginResponse {
			IsSuccess: false,
		})
		return
	}

	claims := jwt.StandardClaims{
			ExpiresAt: h.timeProvider().Add(h.tokenSettings.ExpiryDuration).Unix(),
			Issuer:    h.tokenSettings.Issuer,
			Subject: 	cif,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.tokenSettings.SigningKey))
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithJSON(w, http.StatusOK, LoginResponse {
		IsSuccess: true,
		CustomerCIF: cif,
		AuthToken: signedToken,
	})
}

func parseRequest(r *http.Request) (request LoginRequest, err error, errorCode int) {
	switch r.Method {
	case http.MethodPost:
		e := json.NewDecoder(r.Body).Decode(&request)
		if e != nil { return request, fmt.Errorf("Error parsing JSON request: %s", e), http.StatusBadRequest }
		return request, nil, http.StatusOK

	default:
		return request, fmt.Errorf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed
	}
}

func dummyAuthenticator (request LoginRequest) (CustomerCIF string, err error){
	if len(request.Username) == 10 && request.Password == "Password1" {
		return request.Username, nil
	} else {
		return "", nil
	}
}