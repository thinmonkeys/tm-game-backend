package respond

import (
	"encoding/json"
	"net/http"
	"os"
)

// WithJSON writes a JSON response.
func WithJSON(w http.ResponseWriter, status int, of interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	e := json.NewEncoder(w)
	err := e.Encode(of)
	if err != nil {
		http.Error(w, "error encoding error", http.StatusInternalServerError)
	}
}

// Error returned by the API.
type Error struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// WithError responds with a JSON error.
func WithError(w http.ResponseWriter, status int, msg string) {
	if os.Getenv("STAGE") == "prod" {
		WithJSON(w, status, Error{
			Error:  "Internal Server Error",
			Status: status,
		})
	} else {
		WithJSON(w, status, Error{
			Error:  msg,
			Status: status,
		})
	}
}

// WithOK responds with a JSON OK.
func WithOK(w http.ResponseWriter) {
	WithJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
	})
}
