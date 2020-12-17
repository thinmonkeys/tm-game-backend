package api

import (
	helloWorldHandler "../handlers/helloworld"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func New() (*chi.Mux, error) {
	hw := helloWorldHandler.NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/helloWorld", hw.SayHello)
	return r, nil
}