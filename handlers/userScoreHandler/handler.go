package userscorehandler

import (
	"net/http"
	"../../respond"
	"../../store"
	"../common"
)

type ScoreGetter func(cif string) (record db.DynamicScoreRecord, ok bool, err error)

type UserScoreHandler struct {
	scoreGetter ScoreGetter
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func NewHandler() UserScoreHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	return UserScoreHandler{
		scoreGetter: store.Get,
		requestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}

func (h *UserScoreHandler) GetScoreRecord(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

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