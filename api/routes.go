package api

import (
	userScoreHandler "../handlers/userScoreHandler"
	directDebitHandler "../handlers/directdebits"
	standingOrderHandler "../handlers/standingorders"
	incomeHandler "../handlers/incomes"
	contactDetailsHandler "../handlers/contactdetails"
	commonHandler "../handlers/common"
	loginHandler "../handlers/login"
	helloHandler "../handlers/helloworld"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func New() (*chi.Mux, error) {
	login := loginHandler.NewHandler()
	us := userScoreHandler.NewHandler()
	ch := commonHandler.DefaultConfirmationHandler()
	dd := directDebitHandler.NewHandler(ch)
	so := standingOrderHandler.NewHandler(ch)
	inc := incomeHandler.NewHandler(ch)
	cd := contactDetailsHandler.NewHandler(ch)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/login", login.Login)
	r.Get("/score", us.GetScore)

	hw := helloHandler.NewHandler()
	r.Get("/helloworld", hw.SayHello)

	r.Get("/directdebits", dd.GetDirectDebits)	
	r.Post("/directdebits", dd.ConfirmDirectDebits)
	r.Put("/directdebits", dd.UpdateDirectDebit)

	r.Get("/standingorders", so.GetPayments)	
	r.Post("/standingorders", so.ConfirmPayments)
	r.Put("/standingorders", so.UpdatePayment)

	r.Get("/incomes", inc.GetPayments)	
	r.Post("/incomes", inc.ConfirmPayments)
	r.Put("/incomes", inc.UpdatePayment)

	r.Get("/contactdetails", cd.GetContactDetails)	
	r.Post("/contactdetails", cd.ConfirmContactDetails)
	r.Put("/contactdetails/mobile", cd.SaveMobileNumber)
	r.Put("/contactdetails/home", cd.SaveHomeNumber)
	r.Put("/contactdetails/email", cd.SaveEmailAddress)
	r.Put("/contactdetails/address", cd.SaveAddress)

	return r, nil
}