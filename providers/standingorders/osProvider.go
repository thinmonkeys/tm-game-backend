package standingorders

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"../../payments"
	"../common"
)

type StandingOrderProvider struct {
	connection common.ConnectionSettings
	accountCache common.CustomerAccountCache
}

func NewProvider() StandingOrderProvider {
	connection := common.DefaultConnectionSettings()
	return StandingOrderProvider {
		connection: connection,
		accountCache: common.NewCache(&connection),
	}
}

func (sop StandingOrderProvider) GetStandingOrders(cif string) ([]payments.Payment, error) {
	osSOs, err := sop.getOutsystemsStandingOrders(cif)
	if err != nil { return nil, err }

	results := []payments.Payment{}
	for _,osSO := range osSOs {
		dueDate, err := time.Parse(common.DateOnlyFormat, osSO.PaymentDate)
		if err != nil { return nil, fmt.Errorf("Error decoding date value '%s' as date: %s", osSO.PaymentDate, err.Error()) }
		id, err := strconv.Atoi(osSO.PaymentID)
		if err != nil { return nil, fmt.Errorf("Error decoding ID value '%s' as int64: %s", osSO.PaymentID, err.Error()) }
		payeeId, err := strconv.Atoi(osSO.Payee.PayeeID)
		if err != nil { return nil, fmt.Errorf("Error decoding Payee ID value '%s' as int64: %s", osSO.Payee.PayeeID, err.Error()) }
		results = append(results, payments.Payment {
			ID: id,
			RecipientID: payeeId,
			RecipientName: osSO.Payee.Name,
			DueDate: dueDate,
			Frequency: osFrequencyMapFromFiserv[osSO.PaymentFrequency],
			AmountPence: int(math.Round(osSO.Amount * 100)),
		})
	}

	return results, nil
}

func (sop StandingOrderProvider) SaveStandingOrder(cif string, payment payments.Payment) (err error) { 
	accountID, err := sop.accountCache.GetPrimaryAccountId(cif)
	if err != nil { return }

	osSOs, err := sop.getOutsystemsStandingOrders(cif)
	if err != nil { return err }

	id := fmt.Sprintf("%022d", payment.ID)
	found := false
	var osSO osStandingOrder
	for _,so := range osSOs {
		if so.PaymentID == id {
			osSO = so
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Standing order %d not found", payment.ID)
	}

	formattedDate := payment.DueDate.Format(common.DateOnlyFormat)
	changed := false
	if osSO.PaymentDate != formattedDate || osFrequencyMapFromFiserv[osSO.PaymentFrequency] != payment.Frequency {
		osSO.PaymentFrequency = osFrequencyMapToFiserv[payment.Frequency]	
		osSO.PaymentDate = formattedDate
		changed = true
	}
	if int(math.Round(osSO.Amount * 100)) != payment.AmountPence {
		osSO.Amount = float64(payment.AmountPence) * float64(0.01)
		changed = true
	}

	if changed {
		_,err = sop.connection.RunRequest(http.MethodPut, fmt.Sprintf("/standingorders/%s/%s", cif, accountID), osSO)
	}
	return
}

type osStandingOrder struct {
	PaymentID string
	Amount float64
	PaymentDate string
	PaymentReference string
	EndDate string
	PaymentFrequency string
	PaymentFrequencyDescription string
	Payee osPayee
}

type osPayee struct {
	PayeeID string
	Name string
	SortCode string
	AccountNumber string
	Reference string
	Nickname string
}

func (sop StandingOrderProvider) getOutsystemsStandingOrders(cif string) (osSOs []osStandingOrder, err error) {
	osSOs = []osStandingOrder{}
	accountID, err := sop.accountCache.GetPrimaryAccountId(cif)
	if err != nil { return }

	response, err := sop.connection.RunRequest(http.MethodGet, fmt.Sprintf("/standingorders/%s/%s", cif, accountID), nil)
	if err != nil { return }
	
	err = json.NewDecoder(response.Body).Decode(&osSOs)
	if err != nil { err = fmt.Errorf("Error decoding JSON response: %s", err.Error()) }
	return
}

var osFrequencyMapFromFiserv map[string]payments.Frequency = map[string]payments.Frequency{
	"Daily": payments.FrequencyDaily,
	"Weekly": payments.FrequencyWeekly,
	"BiWeekly": payments.FrequencyFortnightly,
	"TwiceMonthly": payments.FrequencyFortnightly,
	"Monthly": payments.FrequencyMonthly,
	"FourWeeks": payments.FrequencyMonthly,
	"BiMonthly": payments.FrequencyMonthly,
	"FirstOfMonth": payments.FrequencyMonthly,
	"Quarterly": payments.FrequencyQuarterly,
	"SemiAnnually": payments.FrequencyQuarterly,
	"Annual": payments.FrequencyAnnually,
	"EndOfMonth": payments.FrequencyMonthly,
}

var osFrequencyMapToFiserv map[payments.Frequency]string = map[payments.Frequency]string{
	payments.FrequencyWeekly: "Weekly",
	payments.FrequencyFortnightly: "BiWeekly",
	payments.FrequencyMonthly: "Monthly",
	payments.FrequencyQuarterly: "Quarterly",
	payments.FrequencyAnnually: "Annual",
}
