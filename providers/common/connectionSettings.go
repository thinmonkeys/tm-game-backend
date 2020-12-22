package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)


type CallHTTP func(*http.Request) (*http.Response, error)

type ConnectionSettings struct {
	ApiBaseUrl string
	ApiKey     string
	CallHTTP   CallHTTP
}

func DefaultConnectionSettings() ConnectionSettings {
	return ConnectionSettings{
		ApiBaseUrl: "https://thinkmoney-dev.outsystemsenterprise.com/thinmonkeys_api/rest",
		ApiKey:     "th1nm0nkeys!",
		CallHTTP:   http.DefaultClient.Do,
	}
}

func (connection *ConnectionSettings) RunRequest(method string, relativeUrl string, requestBody interface{}) (*http.Response, error) {
	var body io.Reader = nil
	if requestBody != nil {
		jsonBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("Error marshalling request as JSON: %s", err.Error())
		}
		body = bytes.NewReader(jsonBytes)
	}

	httpReq, err := http.NewRequest(method, connection.ApiBaseUrl + relativeUrl, body)
	if err != nil {
		return nil, fmt.Errorf("Error generating HTTP request: %s", err.Error())
	}
	httpReq.Header.Add("x-api-key", connection.ApiKey);
	if requestBody != nil {
		httpReq.Header.Add("Content-Type", "application/json");
	}
	return connection.CallHTTP(httpReq);
}