package directdebits

import (
	"net/http"
	"time"

	"../../respond"
	"../../store"
	"../common"
	"../../payments"
	ddProvider "../../providers/directdebits"
)

type DirectDebitResponse struct {
	DirectDebitList []payments.Payment
	LastConfirmed time.Time
}

type DirectDebitHandler struct {
	scoreGetter common.ScoreGetter
	scorePutter common.ScorePutter
	paymentLister payments.PaymentLister
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func NewHandler() DirectDebitHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	provider := ddProvider.NewProvider()
	return DirectDebitHandler{
		scoreGetter: store.Get,
		scorePutter: store.Put,
		paymentLister: provider.GetDirectDebits,
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

	score, scoreFound, err := h.scoreGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}

	lastConfirmed := time.Time{}
	if scoreFound { lastConfirmed = score.LastUpdatedDirectDebits }

	respond.WithJSON(w, http.StatusOK, DirectDebitResponse {
		DirectDebitList: dds,
		LastConfirmed: lastConfirmed,
	})
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

	score, scoreFound, err := h.scoreGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}

	if !scoreFound {
		score = db.DynamicScoreRecord{
			CustomerCIF: cif,
			LastUpdatedDirectDebits: time.Now(),
			Score: 100,
		}
		h.scorePutter(score)
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else if score.LastUpdatedDirectDebits.AddDate(0, 1, 0).Before(time.Now()) {
		score.Score += 100
		score.LastUpdatedDirectDebits = time.Now()
		h.scorePutter(score)
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else {
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 0, score.LastUpdatedDirectDebits.AddDate(0, 1, 0) })
	}
}

func ListDummyDirectDebits(cif string) (dds []payments.Payment, err error){
	return []payments.Payment {
		payments.Build(1, 301, "Manchester City Council", time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 10875),
		payments.Build(2, 302, "Sky TV", time.Date(2021, 1, 14, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 3000),
		payments.Build(3, 303, "Vodafone", time.Date(2020, 12, 29, 0, 0, 0, 0, time.Local), payments.FrequencyMonthly, 2500),
	}, nil
}