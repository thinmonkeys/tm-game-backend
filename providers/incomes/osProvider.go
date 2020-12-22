package incomes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"../../payments"
	"../common"
)

type IncomeProvider struct {
	connection common.ConnectionSettings
	accountCache common.CustomerAccountCache
}

func NewProvider() IncomeProvider {
	connection := common.DefaultConnectionSettings()
	return IncomeProvider {
		connection: connection,
		accountCache: common.NewCache(&connection),
	}
}

func (ip IncomeProvider) GetIncomes(cif string) ([]payments.Payment, error) {
	//osIncomes, err := ip.getOutsystemsIncomes(cif)
	//if err != nil { return nil, err }

	results := []payments.Payment{}
	/*for _,osIncome := range osIncomes {
		dueDate, err := time.Parse(common.DateOnlyFormat, osIncome.PaymentDate)
		if err != nil { return nil, fmt.Errorf("Error decoding date value '%s' as date: %s", osIncome.PaymentDate, err.Error()) }
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
	}*/

	return results, nil
}

func (ip IncomeProvider) SaveIncome(cif string, payment payments.Payment) (err error) { 
	accountID, err := ip.accountCache.GetPrimaryAccountId(cif)
	if err != nil || accountID == "" { return }

	return
}

type osIncome struct {
}


func (ip IncomeProvider) getOutsystemsIncomes(cif string) (osIncomes []osIncome, err error) {
	osIncomes = []osIncome{}
	accountID, err := ip.accountCache.GetPrimaryAccountId(cif)
	if err != nil { return }

	response, err := ip.connection.RunRequest(http.MethodGet, fmt.Sprintf("/incomes/%s/%s", cif, accountID), nil)
	if err != nil { return }
	
	err = json.NewDecoder(response.Body).Decode(&osIncomes)
	if err != nil { err = fmt.Errorf("Error decoding JSON response: %s", err.Error()) }
	return
}
