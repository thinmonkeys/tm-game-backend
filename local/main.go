package main

import (
	"net/http"
	"../api"
)

func main() {
	r, err := api.New()
	if err != nil {
		panic(err)
	}
	if err = http.ListenAndServe(":3001", r); err != nil {
		panic(err)
	}
}
