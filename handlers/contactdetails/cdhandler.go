package contactdetails

import (
	"net/http"
	"time"

	"../../respond"
	db "../../store"
	"../common"
)

type ContactDetailsGetter func(cif string) (ContactDetails, error)

type ContactDetailsResponse struct {
	ContactDetails ContactDetails
	LastConfirmed time.Time
}

type ContactDetailsHandler struct {
	scoreGetter common.ScoreGetter
	scorePutter common.ScorePutter
	contactDetailsGetter ContactDetailsGetter
}

func NewHandler() ContactDetailsHandler {
	store, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	return ContactDetailsHandler{
		scoreGetter: store.Get,
		scorePutter: store.Put,
		contactDetailsGetter: GetDummyContactDetails,
	}
}

func (h *ContactDetailsHandler) GetContactDetails(w http.ResponseWriter, r *http.Request) {
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

	details, err := h.contactDetailsGetter(cif)
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
			LastUpdatedContactDetails: time.Now(),
			Score: 100,
		}
		h.scorePutter(score)
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else if score.LastUpdatedContactDetails.AddDate(0, 1, 0).Before(time.Now()) {
		score.Score += 100
		score.LastUpdatedContactDetails = time.Now()
		h.scorePutter(score)
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) })
	} else {
		respond.WithJSON(w, http.StatusOK, common.ConfirmationResponse { 0, score.LastUpdatedContactDetails.AddDate(0, 1, 0) })
	}
}

func GetDummyContactDetails(cif string) (details ContactDetails, err error){
	return BuildContactDetails (
		cif,
		"Freda",
		"Flintstone",
		"07777123456",
		"01617731234",
		"freda@theflintstones.co.uk",
		BuildAddress("Building 1", "Think Park", "", "Mosley Road", "Trafford Park", "Manchester", "Lancashire", "M17 1FQ"),
	), nil
}