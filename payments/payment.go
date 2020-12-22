package payments

import "time"

type Payment struct {
	ID            int
	RecipientID   int
	RecipientName string
	DueDate       time.Time
	Frequency     Frequency
	AmountPence   int
}

func Build(id int, recipientId int, recipient string, dueDate time.Time, freq Frequency, amountPence int) Payment {
	return Payment { id, recipientId, recipient, dueDate, freq, amountPence } 
}

type PaymentLister func(cif string) ([]Payment, error)

type Frequency string

const (
	FrequencyDaily       Frequency = "Daily"
	FrequencyWeekly      Frequency = "Weekly"
	FrequencyMonthly     Frequency = "Monthly"
	FrequencyFortnightly Frequency = "Fortnightly"
	FrequencyQuarterly   Frequency = "Quarterly"
	FrequencyAnnually    Frequency = "Annually"
)