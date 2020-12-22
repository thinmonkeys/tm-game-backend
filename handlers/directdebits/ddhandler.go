package directdebits

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"../../payments"
	ddProvider "../../providers/directdebits"
	"../../respond"
	"../common"
)

type DirectDebitResponse struct {
	DirectDebitList []payments.Payment
	LastConfirmed time.Time
	LastScored time.Time
	Badges []common.BadgeType
}

type DirectDebitHandler struct {
	common.ConfirmationHandler
	paymentLister payments.PaymentLister
	paymentUpdater payments.PaymentUpdater
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func NewHandler(confirmationHandler common.ConfirmationHandler) DirectDebitHandler {
	provider := ddProvider.NewProvider()
	return DirectDebitHandler{
		ConfirmationHandler: confirmationHandler,
		paymentLister: provider.GetDirectDebits,
		paymentUpdater: provider.SaveDirectDebit,
		requestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}

func (h *DirectDebitHandler) GetDirectDebits(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dds, err := h.paymentLister(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}

	response := DirectDebitResponse {
		DirectDebitList: dds,
		LastConfirmed: time.Time{},
		LastScored: time.Time{},
	}

	scoreCategory, scoreFound, err := h.ConfirmationHandler.CategoryGetter(cif, common.ScoreCategoryDirectDebits.Code)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	if scoreFound { 
		response.LastConfirmed = scoreCategory.LastConfirmed 
		response.LastScored = scoreCategory.LastScored
		response.Badges, err = h.ConfirmationHandler.GetBadgesByCategory(cif, common.ScoreCategoryDirectDebits)
		if err != nil {
			respond.WithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting badges: %s", err.Error()));
			return
		}
	}

	respond.WithJSON(w, http.StatusOK, response)
}

func (h *DirectDebitHandler) ConfirmDirectDebits(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodPost) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response, err := h.ConfirmationHandler.ConfirmCategory(cif, common.ScoreCategoryDirectDebits)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithJSON(w, http.StatusOK, response)
}

func (h *DirectDebitHandler) UpdateDirectDebit(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodPut) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "PUT only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	directDebit := payments.Payment{}
	err = json.NewDecoder(r.Body).Decode(&directDebit)
	if err != nil {
		respond.WithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.paymentUpdater(cif, directDebit)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithOK(w)
}

func ListDummyDirectDebits(cif string) (dds []payments.Payment, err error){
	return []payments.Payment {
		payments.Build(1, 301, "Manchester City Council", time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 10875),
		payments.Build(2, 302, "Sky TV", time.Date(2021, 1, 14, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 3000),
		payments.Build(3, 303, "Vodafone", time.Date(2020, 12, 29, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 2500),
	}, nil
}