package api

import (
	helloWorldHandler "../handlers/helloworld"
	userScoreHandler "../handlers/userScoreHandler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func New() (*chi.Mux, error) {
	hw := helloWorldHandler.NewHandler()
	us := userScoreHandler.NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/helloWorld", hw.SayHello)
	r.Get("/getScore", us.GetScoreRecord)
	return r, nil
}