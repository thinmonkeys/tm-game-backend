package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"../../payments"
	"../../respond"
)

type PaymentResponse struct {
	Payments []payments.Payment
	LastConfirmed time.Time
	LastScored time.Time
	Badges []BadgeType
}

type PaymentHandler struct {
	ConfirmationHandler
	Category ScoreCategory
	PaymentLister payments.PaymentLister
	PaymentUpdater payments.PaymentUpdater
	RequestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func (h *PaymentHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cif, err := h.RequestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	payments, err := h.PaymentLister(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}

	response := PaymentResponse {
		Payments: payments,
		LastConfirmed: time.Time{},
		LastScored: time.Time{},
	}

	scoreCategory, scoreFound, err := h.ConfirmationHandler.CategoryGetter(cif, h.Category.Code)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	if scoreFound { 
		response.LastConfirmed = scoreCategory.LastConfirmed 
		response.LastScored = scoreCategory.LastScored
		response.Badges, err = h.ConfirmationHandler.GetBadgesByCategory(cif, h.Category)
		if err != nil {
			respond.WithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting badges: %s", err.Error()));
			return
		}
	}

	respond.WithJSON(w, http.StatusOK, response)
}

func (h *PaymentHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodPut) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "PUT only")
		return
	}

	cif, err := h.RequestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	payment := payments.Payment{}
	err = json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		respond.WithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.PaymentUpdater(cif, payment)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithOK(w)
}

func (h *PaymentHandler) ConfirmPayments(w http.ResponseWriter, r *http.Request) {
	h.ConfirmationHandler.HandleConfirmRequest(w, r, h.Category, h.RequestAuthenticator)
}