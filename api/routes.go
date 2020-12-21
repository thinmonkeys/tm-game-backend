package api

import (
	helloWorldHandler "../handlers/helloworld"
	userScoreHandler "../handlers/userScoreHandler"
	directDebitHandler "../handlers/directdebits"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func New() (*chi.Mux, error) {
	hw := helloWorldHandler.NewHandler()
	us := userScoreHandler.NewHandler()
	dd := directDebitHandler.NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/helloWorld", hw.SayHello)
	r.Get("/getScore", us.GetScoreRecord)

	r.Get("/directdebits", dd.GetDirectDebits)
	return r, nil
}