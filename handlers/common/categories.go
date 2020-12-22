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
	ScoreCategoryAll = ScoreCategory { "ALL", "Contact Details" }
	AllScoreCategories []ScoreCategory
	ScoreCategoryLookup map[string]ScoreCategory
)

func init() {
	AllScoreCategories = []ScoreCategory { 
		ScoreCategoryDirectDebits,
		ScoreCategoryStandingOrders,
		ScoreCategoryIncomes,
		ScoreCategoryContactDetails,
		ScoreCategoryAll,
	}
	ScoreCategoryLookup = map[string]ScoreCategory{}
	for _,sc := range AllScoreCategories {
		ScoreCategoryLookup[sc.Code] = sc
	}
}