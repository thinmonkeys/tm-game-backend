package common

type BadgeType struct {
	Code string
	Name string
	Category ScoreCategory
	Level int
}

var (
	BadgeTypeDirectDebitLevel1 = BadgeType { "DD1", "Direct Debit Checker", ScoreCategoryDirectDebits, 1 }
	BadgeTypeDirectDebitLevel2 = BadgeType { "DD2", "Direct Debit Pro", ScoreCategoryDirectDebits, 2 }
	BadgeTypeDirectDebitLevel3 = BadgeType { "DD3", "Direct Debit Wizard", ScoreCategoryDirectDebits, 3  }
	BadgeTypeStandingOrdersLevel1 = BadgeType { "SO1", "Standing Order Checker", ScoreCategoryStandingOrders, 1 }
	BadgeTypeStandingOrdersLevel2 = BadgeType { "SO2", "Standing Order Pro", ScoreCategoryStandingOrders, 2 }
	BadgeTypeStandingOrdersLevel3 = BadgeType { "SO3", "Standing Order Wizard", ScoreCategoryStandingOrders, 3 }
	BadgeTypeIncomesLevel1 = BadgeType { "IN1", "Income Checker", ScoreCategoryIncomes, 1 }
	BadgeTypeIncomesLevel2 = BadgeType { "IN2", "Income Pro", ScoreCategoryIncomes, 2 }
	BadgeTypeIncomesLevel3 = BadgeType { "IN3", "Income Wizard", ScoreCategoryIncomes, 3 }
	BadgeTypeContactDetails1 = BadgeType { "CD1", "Contact Details Checker", ScoreCategoryContactDetails, 1 }
	BadgeTypeContactDetails2 = BadgeType { "CD2", "Contact Details Pro", ScoreCategoryContactDetails, 2 }
	BadgeTypeContactDetails3 = BadgeType { "CD3", "Contact Details Wizard", ScoreCategoryContactDetails, 3 }
	BadgeTypeAllDetails1 = BadgeType { "ALL1", "Account Checker", ScoreCategoryAll, 1 }
	BadgeTypeAllDetails2 = BadgeType { "ALL2", "Account Pro", ScoreCategoryAll, 2 }
	BadgeTypeAllDetails3 = BadgeType { "ALL3", "Account Guru", ScoreCategoryAll, 3 }
	AllBadgeTypes []BadgeType
	BadgeTypeLookup map[string]BadgeType
)

func init() {
	AllBadgeTypes = []BadgeType { 
		BadgeTypeDirectDebitLevel1,
		BadgeTypeDirectDebitLevel2,
		BadgeTypeDirectDebitLevel3,
		BadgeTypeStandingOrdersLevel1,
		BadgeTypeStandingOrdersLevel2,
		BadgeTypeStandingOrdersLevel3,
		BadgeTypeIncomesLevel1,
		BadgeTypeIncomesLevel2,
		BadgeTypeIncomesLevel3,
		BadgeTypeContactDetails1,
		BadgeTypeContactDetails2,
		BadgeTypeContactDetails3,
		BadgeTypeAllDetails1,
		BadgeTypeAllDetails2,
		BadgeTypeAllDetails3,
	}
	BadgeTypeLookup = map[string]BadgeType{}
	for _,bt := range AllBadgeTypes {
		BadgeTypeLookup[bt.Code] = bt
	}
}

func (h *ConfirmationHandler) GetBadgesByCategory(cif string, cat ScoreCategory) ([]BadgeType, error) {
	return h.getBadges(cif, func(bt BadgeType)bool { return bt.Category == cat })
}

func (h *ConfirmationHandler) GetAllBadges(cif string) ([]BadgeType, error) {
	return h.getBadges(cif, func(bt BadgeType)bool { return true })
}

func (h *ConfirmationHandler) getBadges(cif string, predicate func(BadgeType)bool) ([]BadgeType, error) {
	allBadges, err := h.BadgeGetter(cif)
	if err != nil { return nil, err }

	matchingBadges := []BadgeType{}
	for _,b := range allBadges {
		badgeType,ok := BadgeTypeLookup[b.BadgeCode]
		if ok && predicate(badgeType) {
			matchingBadges = append(matchingBadges, badgeType)
		}
	}
	return matchingBadges, nil
}

func GetBadgeType(cat ScoreCategory, level int) (BadgeType, bool) {
	for _,b := range AllBadgeTypes {
		if (b.Category == cat && b.Level == level) { return b, true }
	}
	return BadgeType{}, false
}