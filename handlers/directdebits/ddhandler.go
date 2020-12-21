package directdebits

import (
	"net/http"
	"time"

	"../../respond"
	"../../store"
)

type ScoreGetter func(cif string) (db.DynamicScoreRecord, bool, error)
type ScorePutter func(record db.DynamicScoreRecord) (error)
type DirectDebitLister func(cif string) ([]DirectDebit, error)

type DirectDebitResponse struct {
	DirectDebitList []DirectDebit
	LastConfirmed time.Time
}

type DirectDebitHandler struct {
	scoreGetter ScoreGetter
	scorePutter ScorePutter
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

func ListDummyDirectDebits(cif string) (dds []DirectDebit, err error){
	return []DirectDebit {
		BuildDirectDebit(1, "3040506070", 301, "Manchester City Council", 10875, 401, "Monthly on the 1st", time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)),
		BuildDirectDebit(1, "3040506070", 302, "Sky TV", 3000, 414, "Monthly on the 14th", time.Date(2021, 1, 14, 0, 0, 0, 0, time.Local)),
		BuildDirectDebit(1, "3040506070", 303, "Vodafone", 2500, 429, "Monthly on the 29th", time.Date(2020, 12, 29, 0, 0, 0, 0, time.Local)),
	}, nil
}