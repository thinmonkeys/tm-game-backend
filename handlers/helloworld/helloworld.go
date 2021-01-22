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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	io.WriteString(w, `<html>
    <head>
        <title>Hello World</title>
    </head>
    <body>
        This is a page presented inside the thinkmoney app but actually hosted in a separate AWS website.
<br/><br/>
        <a href="index.html">Refresh</a>
    </body>
</html>`)
}