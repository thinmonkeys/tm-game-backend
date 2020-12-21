package directdebits

import (
	"net/http"
	"time"

	"../../respond"
	"../../store"
	"../common"
)

type DirectDebitLister func(cif string) ([]DirectDebit, error)

type DirectDebitResponse struct {
	DirectDebitList []DirectDebit
	LastConfirmed time.Time
}

type DirectDebitHandler struct {
	scoreGetter common.ScoreGetter
	scorePutter common.ScorePutter
	directDebitLister DirectDebitLister
}

func NewHandler() DirectDebitHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	return DirectDebitHandler{
		scoreGetter: store.Get,
		scorePutter: store.Put,
		directDebitLister: ListDummyDirectDebits,
	}
}

func (h *DirectDebitHandler) GetDirectDebits(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cifs := r.URL.Query()["cif"]
	if len(cifs) == 0 {
		respond.WithError(w, http.StatusBadRequest, "Please provide the Customer CIF in the querystring (?cif=)")
		return
	}
	cif := cifs[0]

	dds, err := h.directDebitLister(cif)
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

	cifs := r.URL.Query()["cif"]
	if len(cifs) == 0 {
		respond.WithError(w, http.StatusBadRequest, "Please provide the Customer CIF in the querystring (?cif=)")
		return
	}
	cif := cifs[0]

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

func ListDummyDirectDebits(cif string) (dds []DirectDebit, err error){
	return []DirectDebit {
		BuildDirectDebit(1, "3040506070", 301, "Manchester City Council", 10875, 401, "Monthly on the 1st", time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)),
		BuildDirectDebit(1, "3040506070", 302, "Sky TV", 3000, 414, "Monthly on the 14th", time.Date(2021, 1, 14, 0, 0, 0, 0, time.Local)),
		BuildDirectDebit(1, "3040506070", 303, "Vodafone", 2500, 429, "Monthly on the 29th", time.Date(2020, 12, 29, 0, 0, 0, 0, time.Local)),
	}, nil
}