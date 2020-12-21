package directdebits

import "time"

type DirectDebit struct {
	ID                   int
	AccountNumber        string
	CreditorID           int
	CreditorName         string
	AmountPence          int
	FrequencyID          int
	FrequencyDescription string
	NextDueDate          time.Time
}

func BuildDirectDebit(id int, accNo string, credID int, credName string, amountPence int, freqID int, freqDesc string, nextDue time.Time) DirectDebit {
	return DirectDebit{id, accNo, credID, credName, amountPence, freqID, freqDesc, nextDue }
}