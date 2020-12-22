package contactdetails

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	cd "../../contactdetails"
	cdProvider "../../providers/contactdetails"
	"../../respond"
	"../common"
)

type ContactDetailsGetter func(cif string) (cd.ContactDetails, error)

type ContactDetailsResponse struct {
	ContactDetails cd.ContactDetails
	LastConfirmed time.Time
	LastScored time.Time
	Badges []common.BadgeType
}

type ContactDetailsHandler struct {
	common.ConfirmationHandler
	provider cd.ContactDetailsProvider
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

func NewHandler(confirmationHandler common.ConfirmationHandler) ContactDetailsHandler {
	return ContactDetailsHandler{
		ConfirmationHandler: confirmationHandler,
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

	response := ContactDetailsResponse {
		ContactDetails: details,
		LastConfirmed: time.Time{},
		LastScored: time.Time{},
	}

	scoreCategory, scoreFound, err := h.ConfirmationHandler.CategoryGetter(cif, common.ScoreCategoryContactDetails.Code)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	if scoreFound { 
		response.LastConfirmed = scoreCategory.LastConfirmed 
		response.LastScored = scoreCategory.LastScored
		response.Badges, err = h.ConfirmationHandler.GetBadgesByCategory(cif, common.ScoreCategoryContactDetails)
		if err != nil {
			respond.WithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting badges: %s", err.Error()));
			return
		}
	}

	respond.WithJSON(w, http.StatusOK, response)
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

	response, err := h.ConfirmationHandler.ConfirmCategory(cif, common.ScoreCategoryContactDetails)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithJSON(w, http.StatusOK, response)
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
func (h *ContactDetailsHandler) SaveAddress(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodPut) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "PUT only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	newAddress := cd.Address{}
	err = json.NewDecoder(r.Body).Decode(&newAddress)
	if err != nil {
		respond.WithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.provider.SaveAddress(cif, newAddress)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithOK(w)
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

