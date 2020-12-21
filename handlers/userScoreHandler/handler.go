package userscorehandler

import (
	"net/http"
	"../../respond"
	"../../store"
)

type ScoreGetter func(cif string) (record db.DynamicScoreRecord, ok bool, err error)

type UserScoreHandler struct {
	scoreGetter ScoreGetter
}

func NewHandler() UserScoreHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	return UserScoreHandler{
		scoreGetter: store.Get,
	}
}

func (h *UserScoreHandler) GetScoreRecord(w http.ResponseWriter, r *http.Request) {
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

	record, ok, err := h.scoreGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	if !ok {
		respond.WithError(w, http.StatusNotFound, "No record found for customer " + cif)
		return
	}
	respond.WithJSON(w, http.StatusOK, record)
}