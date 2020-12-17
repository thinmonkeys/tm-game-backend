package helloworld

import (
	"fmt"
	"net/http"
	"../../respond"
)

type HelloWorldHandler struct {
}

func NewHandler() HelloWorldHandler {
	return HelloWorldHandler{}
}

func (h *HelloWorldHandler) SayHello(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	names := r.URL.Query()["name"]
	if len(names) == 0 {
		respond.WithError(w, http.StatusBadRequest, "Please provide your name in the querystring")
		return
	}

	respond.WithJSON(w, http.StatusOK, fmt.Sprintf("Hello %s!", names[0]))
}