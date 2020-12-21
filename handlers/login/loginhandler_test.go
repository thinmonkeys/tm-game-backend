package login

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"../common"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	testCases := []struct {
		label string
		request *http.Request
		mockAuthenticator LoginAuthenticator
		expectedResponseCode int
		expectedResponseBody string
		expectToken bool
		testTime time.Time
		signingKey string
		tokenExpiryTime time.Duration
		expectedCIFKey string
		expectedExpiryDate time.Time
	} {
		{ "Wrong password",
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{ "username":"RandomUser", "password":"SomePassword" }`)),
			func(r LoginRequest) (CustomerCIF string, err error) { 
				assert.Equal(t, "RandomUser", r.Username, "Username passed to Authenticator")
				assert.Equal(t, "SomePassword", r.Password, "Password passed to Authenticator")
				return "", nil 
			},
			http.StatusUnauthorized,
			`{"isSuccess":false}`,
			false,
			time.Date(2020, time.November, 18, 12, 42, 15, 0, time.Local), 
			"thinmonkeysSignature",
			time.Minute * time.Duration(30),
			"",
			time.Time{}, 
			},
		{ "Error in auth",
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{ "username":"Fred123", "password":"Password6" }`)),
			func(r LoginRequest) (CustomerCIF string, err error) { 
				assert.Equal(t, "Fred123", r.Username, "Username passed to Authenticator")
				assert.Equal(t, "Password6", r.Password, "Password passed to Authenticator")
				return "", errors.New("Something went wrong") 
			},
			http.StatusInternalServerError,
			`{"error":"Something went wrong","status":500}`,
			false,
			time.Date(2020, time.November, 18, 12, 42, 15, 0, time.Local), 
			"thinmonkeysSignature",
			time.Minute * time.Duration(30),
			"",
			time.Time{}, 
			},
		{ "Correct password",			
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{ "username":"IanTest666", "password":"Floccinaucinihilipilification17" }`)),
			func(r LoginRequest) (CustomerCIF string, err error) { 
				assert.Equal(t, "IanTest666", r.Username, "Username passed to Authenticator")
				assert.Equal(t, "Floccinaucinihilipilification17", r.Password, "Password passed to Authenticator")
				return "4006001202", nil
			},
			http.StatusOK,
			`{"isSuccess":true,"customerCIF":"4006001202","authToken":"([A-Za-z0-9\-_]+.[A-Za-z0-9\-_]+.[A-Za-z0-9\-_]+)"}`,
			true,
			time.Date(2020, time.November, 18, 12, 42, 15, 0, time.Local), 
			"thinmonkeysSignature",
			time.Minute * time.Duration(30),
			"4006001202",
			time.Date(2020, time.November, 18, 13, 12, 15, 0, time.Local), 
			},
	}

	for _,tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			testHandler := LoginHandler { 
				loginAuthenticator: tc.mockAuthenticator,
				tokenSettings: common.TokenSettings {
					ExpiryDuration: tc.tokenExpiryTime,
					SigningKey: tc.signingKey,
					Issuer: "thinmonkeys",
				},
				timeProvider: func() time.Time { return tc.testTime },
			}
			w := httptest.NewRecorder()
			testHandler.Login(w, tc.request)
			result := w.Result()

			assert.Equal(t, tc.expectedResponseCode, result.StatusCode, "Response code")
			body,err := ioutil.ReadAll(result.Body)
			assert.Nil(t, err, "Unhandled error reading result")
			assert.Regexp(t, tc.expectedResponseBody + "\n", string(body), "Response")	
			
			if tc.expectToken {
				re := regexp.MustCompile(tc.expectedResponseBody);
				submatches := re.FindStringSubmatch(string(body))
				token := submatches[1]
				fmt.Printf(token)
				parts := strings.Split(token, ".")
				assert.Equal(t, 3, len(parts), "parts of token")

				header := assertAndUnwrapTokenPart(t, parts[0], "header", 2)
				assertStringJSONFragment(t, header, "alg", "HS256")
				assertStringJSONFragment(t, header, "typ", "JWT")

				body := assertAndUnwrapTokenPart(t, parts[1], "body", 3)

				var exp int64
				err = json.Unmarshal(body["exp"], &exp)
				assert.Nil(t, err, "Error decoding expiration date")
				expDate := time.Unix(exp, 0)
				assert.Equal(t, tc.expectedExpiryDate, expDate, "expiration date")

				assertStringJSONFragment(t, body, "iss", "thinmonkeys")
				assertStringJSONFragment(t, body, "sub", tc.expectedCIFKey)

				jwt.TimeFunc = func()(time.Time) { return tc.testTime }
				parsed,err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
					}
					return []byte("thinmonkeysSignature"), nil
				})
				assert.NotNil(t, parsed, "Parser returned a nil token")
				assert.Nil(t, err, "Error parsing token")
			}
		})
	}
}

func assertStringJSONFragment(t *testing.T, jsonMap map[string]json.RawMessage, key string, expectedValue string) {
	var stringValue string
	err := json.Unmarshal(jsonMap[key], &stringValue)
	assert.Nil(t, err, "Error decoding %s as string", key)
	assert.Equal(t, expectedValue, stringValue, key)
}

func assertAndUnwrapTokenPart(t *testing.T, tokenPart string, description string, expectedLength int) (jsonMap map[string]json.RawMessage) {
	bodyJSON, err := base64.RawURLEncoding.DecodeString(tokenPart)
	assert.Nil(t, err, "Error decoding %s base64", description)
	err = json.Unmarshal(bodyJSON, &jsonMap)
	assert.Nil(t, err, "Error unmarshalling %s JSON", description)
	assert.Equal(t, expectedLength, len(jsonMap), "%s elements", description)
	return
}