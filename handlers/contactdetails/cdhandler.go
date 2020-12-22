package contactdetails

import (
	"net/http"
	"time"
	"io/ioutil"

	"../common"
	"../../respond"
	cd "../../contactdetails"
	db "../../store"
	cdProvider "../../providers/contactdetails"
)

type ContactDetailsGetter func(cif string) (cd.ContactDetails, error)

type ContactDetailsResponse struct {
	ContactDetails cd.ContactDetails
	LastConfirmed time.Time
}

type ContactDetailsHandler struct {
	scoreGetter common.ScoreGetter
	scorePutter common.ScorePutter
	provider cd.ContactDetailsProvider
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func NewHandler() ContactDetailsHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	return ContactDetailsHandler{
		scoreGetter: store.Get,
		scorePutter: store.Put,
		provider: cdProvider.NewProvider(),
		requestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}

func (h *ContactDetailsHandler) GetContactDetails(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	details, err := h.provider.GetContactDetails(cif)
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
	if scoreFound { lastConfirmed = score.LastUpdatedContactDetails }

	respond.WithJSON(w, http.StatusOK, ContactDetailsResponse {
		ContactDetails: details,
		LastConfirmed: lastConfirmed,
	})
}

func (h *ContactDetailsHandler) ConfirmContactDetails(w http.ResponseWriter, r *http.Request) {
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
			LastUpdatedContactDetails: time.Now(),
			Score: 100,
		}
		err = h.scorePutter(score)
		if(err != nil) {
			respond.WithError(w, http.StatusInternalServerError, err.Error());
			return
		}
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else if score.LastUpdatedContactDetails.AddDate(0, 1, 0).Before(time.Now()) {
		score.Score += 100
		score.LastUpdatedContactDetails = time.Now()
		err = h.scorePutter(score)
		if(err != nil) {
			respond.WithError(w, http.StatusInternalServerError, err.Error());
			return
		}
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else {
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 0, score.LastUpdatedContactDetails.AddDate(0, 1, 0) })
	}
}

func (h *ContactDetailsHandler) SaveMobileNumber(w http.ResponseWriter, r *http.Request) {
	h.saveContactDetail(w, r, h.provider.SaveMobileNumber)
}
func (h *ContactDetailsHandler) SaveHomeNumber(w http.ResponseWriter, r *http.Request) {
	h.saveContactDetail(w, r, h.provider.SaveHomeNumber)
}
func (h *ContactDetailsHandler) SaveEmailAddress(w http.ResponseWriter, r *http.Request) {
	h.saveContactDetail(w, r, h.provider.SaveEmailAddress)
}

func (h *ContactDetailsHandler) saveContactDetail(w http.ResponseWriter, r *http.Request, saveMethod func(string,string)error) {
	if(r.Method != http.MethodPut) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "PUT only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respond.WithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = saveMethod(cif, string(body))
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithOK(w)
}

