package directdebits

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"../../payments"
	"../common"
)


type DirectDebitProvider struct {
	connection common.ConnectionSettings
	accountCache common.CustomerAccountCache
}

func NewProvider() DirectDebitProvider {
	connection := common.DefaultConnectionSettings()
	return DirectDebitProvider {
		connection: connection,
		accountCache: common.NewCache(&connection),
	}
}

type osDirectDebit struct {
	Amount float64 //           : 55.0000
	PaymentCategoryId int // : 7
	PaymentTypeId int //   : 18
	CompanyId int //       : 1
	Description string //      : British Red Cross
	DirectDebitID int //    : 10008987
	DueDate string //         : 2017-11-12
	FinalPaymentDate string // : 0001-01-01
	Frequency osFrequency //        : @{DueDay=12; FrequencyID=6}
	NoOfPayments int //     : 12
	Reference string //      : TOMTEST00000000002
	Status string  //         : NotClaiming
}

const dateOnlyFormat string = "2006-01-02"

type osFrequency struct {
	DueDay int //12
	FrequencyID int // 6
	DayOfTheWeek int 
}

var osFrequencyMapFromId map[int]payments.Frequency = map[int]payments.Frequency{
	1: payments.FrequencyWeekly,
	2: payments.FrequencyFortnightly,
	3: payments.FrequencyMonthly,
	4: payments.FrequencyMonthly,
	5: payments.FrequencyMonthly,
	6: payments.FrequencyMonthly,
	7: payments.FrequencyQuarterly,
	8: payments.FrequencyAnnually,
	9: payments.FrequencyMonthly,
}

var osFrequencyMapToId map[payments.Frequency]int = map[payments.Frequency]int{
	payments.FrequencyWeekly: 1,
	payments.FrequencyFortnightly: 2,
	payments.FrequencyMonthly: 6,
	payments.FrequencyQuarterly: 7,
	payments.FrequencyAnnually: 8,
}

func (ddp DirectDebitProvider) GetDirectDebits(cif string) ([]payments.Payment, error) {
	osDDs, err := ddp.getOutsystemsDirectDebits(cif)
	if err != nil { return nil, err }

	results := []payments.Payment{}
	for _,osDD := range osDDs {
		dueDate, err := time.Parse(dateOnlyFormat, osDD.DueDate)
		if err != nil { return nil, fmt.Errorf("Error decoding date value '%s' as date: %s", osDD.DueDate, err.Error()) }
		results = append(results, payments.Payment {
			ID: osDD.DirectDebitID,
			RecipientID: osDD.CompanyId,
			RecipientName: osDD.Description,
			DueDate: dueDate,
			Frequency: osFrequencyMapFromId[osDD.Frequency.FrequencyID],
			AmountPence: int(math.Round(osDD.Amount * 100)),
		})
	}

	return results, nil
}

func (ddp DirectDebitProvider) SaveDirectDebit(cif string, payment payments.Payment) (err error) {
	accountID, err := ddp.accountCache.GetPrimaryAccountId(cif)
	if err != nil { return }

	osDDs, err := ddp.getOutsystemsDirectDebits(cif)
	if err != nil { return err }

	found := false
	var osDD osDirectDebit
	for _,dd := range osDDs {
		if dd.DirectDebitID == payment.ID {
			osDD = dd
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Direct debit %d not found", payment.ID)
	}

	changed := osDD.DueDate != payment.DueDate.Format(dateOnlyFormat) || int(math.Round(osDD.Amount * 100)) != payment.AmountPence
	if osFrequencyMapFromId[osDD.Frequency.FrequencyID] != payment.Frequency {
		osDD.Frequency = mapFrequencyToOutSystems(payment.Frequency, payment.DueDate)
		changed = true
	}
	formattedDate := payment.DueDate.Format(dateOnlyFormat)
	if formattedDate != osDD.DueDate {
		osDD.DueDate = formattedDate
		changed = true
	}
	if int(math.Round(osDD.Amount * 100)) != payment.AmountPence {
		osDD.Amount = float64(payment.AmountPence) * float64(0.01)
		changed = true
	}

	if changed {
		_,err = ddp.connection.RunRequest(http.MethodPut, fmt.Sprintf("/directdebits/%s/%s", cif, accountID), osDD)
	}
	return
}

func (ddp DirectDebitProvider) getOutsystemsDirectDebits(cif string) (osDDs []osDirectDebit, err error) {
	osDDs = []osDirectDebit{}
	accountID, err := ddp.accountCache.GetPrimaryAccountId(cif)
	if err != nil { return }

	response, err := ddp.connection.RunRequest(http.MethodGet, fmt.Sprintf("/directdebits/%s/%s", cif, accountID), nil)
	if err != nil { return }
	
	err = json.NewDecoder(response.Body).Decode(&osDDs)
	if err != nil { err = fmt.Errorf("Error decoding JSON response: %s", err.Error()) }
	return
}

func mapFrequencyToOutSystems(freq payments.Frequency, dueDate time.Time) osFrequency {
	switch freq {
	case payments.FrequencyWeekly, payments.FrequencyFortnightly:
		return osFrequency{
			FrequencyID: osFrequencyMapToId[freq],
			DayOfTheWeek: (int(dueDate.Weekday()) + 6) % 7 + 1,  // golang Weekday gives 0(Sun)-6(Sat), want 1(Mon)-7(Sun)
		}
	default:
		return osFrequency{
			FrequencyID: osFrequencyMapToId[freq],
			DueDay: dueDate.Day(),
		}
	}
}