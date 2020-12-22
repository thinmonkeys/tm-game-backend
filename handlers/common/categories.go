package common

type ScoreCategory struct {
	Code string
	Name string
}

var (
	ScoreCategoryDirectDebits = ScoreCategory { "DD", "Direct Debits" }
	ScoreCategoryStandingOrders = ScoreCategory { "SO", "Standing Orders" }
	ScoreCategoryIncomes = ScoreCategory { "IN", "Incomes" }
	ScoreCategoryContactDetails = ScoreCategory { "CD", "Contact Details" }
)