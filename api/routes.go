package api

import (
	helloWorldHandler "../handlers/helloworld"
	userScoreHandler "../handlers/userScoreHandler"
	directDebitHandler "../handlers/directdebits"
	contactDetailsHandler "../handlers/contactdetails"
	loginHandler "../handlers/login"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func New() (*chi.Mux, error) {
	hw := helloWorldHandler.NewHandler()
	login := loginHandler.NewHandler()
	us := userScoreHandler.NewHandler()
	dd := directDebitHandler.NewHandler()
	cd := contactDetailsHandler.NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/helloWorld", hw.SayHello)
	r.Post("/login", login.Login)
	r.Get("/getScore", us.GetScoreRecord)

	r.Get("/directdebits", dd.GetDirectDebits)	
	r.Post("/directdebits", dd.ConfirmDirectDebits)
	r.Put("/directdebits", dd.UpdateDirectDebit)

	r.Get("/contactdetails", cd.GetContactDetails)	
	r.Post("/contactdetails", cd.ConfirmContactDetails)
	r.Put("/contactdetails/mobile", cd.SaveMobileNumber)
	r.Put("/contactdetails/home", cd.SaveHomeNumber)
	r.Put("/contactdetails/email", cd.SaveEmailAddress)
	r.Put("/contactdetails/address", cd.SaveAddress)

	return r, nil
}