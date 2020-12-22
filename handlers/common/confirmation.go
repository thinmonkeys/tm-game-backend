package common

import (
	"net/http"
	"time"

	"../../respond"
	db "../../store"
)

type ConfirmationResponse struct {
	PointsGained int
	NextPointsEligible time.Time
	NewBadges []BadgeType
}

type ConfirmationHandler struct {
	ScoreGetter ScoreGetter
	ScorePutter ScorePutter
	CategoryGetter CategoryScoreGetter
	CategoryGetAll CategoryScoreGetAll
	CategoryPutter CategoryScorePutter
	BadgeGetter BadgeGetter
	BadgePutter BadgePutter
}


func DefaultConfirmationHandler() ConfirmationHandler {
	scoreStore,err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	categoryStore,err := db.DefaultScoreHistoryStore()
	if(err != nil) { panic(err) }
	badgeStore,err := db.DefaultBadgeHistoryStore()
	if(err != nil) { panic(err) }
	return ConfirmationHandler{
		ScoreGetter: scoreStore.Get,
		ScorePutter: scoreStore.Put,
		CategoryGetter: categoryStore.Get,
		CategoryGetAll: categoryStore.GetAll,
		CategoryPutter: categoryStore.Put,
		BadgeGetter: badgeStore.Get,
		BadgePutter: badgeStore.Put,
	}
}

func (h *ConfirmationHandler) HandleConfirmRequest(w http.ResponseWriter, r *http.Request, category ScoreCategory, authenticator RequestAuthenticatorFunc) {
	if(r.Method != http.MethodPost) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	cif, err := authenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response, err := h.ConfirmCategory(cif, category)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.WithJSON(w, http.StatusOK, response)
}

func (h *ConfirmationHandler) ConfirmCategory(cif string, category ScoreCategory) (resp ConfirmationResponse, err error) {
	resp = ConfirmationResponse{}
	score, scoreFound, err := h.ScoreGetter(cif)
	if err != nil { return }

	categoryRecord, categoryFound, err := h.CategoryGetter(cif, category.Code)
	if err != nil { return }

	if !categoryFound {
		categoryRecord = db.ScoreHistoryRecord {
			CustomerCIF: cif,
			CategoryCode: category.Code,
		}
	}

	pointsGained := 0
	if !scoreFound {
		score = db.DynamicScoreRecord{
			CustomerCIF: cif,
			Score: 0,
		}
	}

	if categoryRecord.LastScored.AddDate(0, 1, 0).Before(time.Now()) {
		pointsGained = 100
		score.Score += pointsGained
		categoryRecord.LastScored = time.Now()
		categoryRecord.TimesScored++
	}

	categoryRecord.LastConfirmed = time.Now()
	categoryRecord.TimesConfirmed++
	err = h.CategoryPutter(categoryRecord)
	if err != nil { return }

	newBadges := []BadgeType{}
	if pointsGained > 0	{
		err = h.ScorePutter(score)
		if err != nil { return }

		newBadges, err = h.handleBadges(cif, category, categoryRecord.TimesScored)
		if err != nil { return }
	}
	
	return ConfirmationResponse { pointsGained, categoryRecord.LastScored.AddDate(0, 1, 0), newBadges }, nil
}

func (h *ConfirmationHandler) handleBadges(cif string, category ScoreCategory, scoreCount int) ([]BadgeType, error) {
	categoryBadges,err := h.GetBadgesByCategory(cif, category)
	if err != nil { return nil, err }

	newBadges := []BadgeType{}

	badgeLevel1,_ := GetBadgeType(category, 1)
	if scoreCount >= 1 && !hasBadge(categoryBadges,badgeLevel1) {
		newBadges = append(newBadges, badgeLevel1)
	}
	badgeLevel2,_ := GetBadgeType(category, 2)
	if scoreCount >= 3 && !hasBadge(categoryBadges,badgeLevel2) {
		newBadges = append(newBadges, badgeLevel2)
	}
	badgeLevel3,_ := GetBadgeType(category, 3)
	if scoreCount >= 6 && !hasBadge(categoryBadges,badgeLevel3) {
		newBadges = append(newBadges, badgeLevel3)
	}

	if len(newBadges) == 0 {
		return newBadges, nil
	}

	allCategoryBadges,err := h.GetBadgesByCategory(cif, ScoreCategoryAll)
	if err != nil { return nil, err }

	minScored := 0
	allCategoryRecords, err := h.CategoryGetAll(cif)
	if err != nil { return nil, err }

	if len(allCategoryRecords) == 4 {
		minScored = 1000
		for _,rec := range allCategoryRecords {
			if(minScored > rec.TimesScored) {
				minScored = rec.TimesScored	
			}
		}
	}
	badgeLevel1,_ = GetBadgeType(ScoreCategoryAll, 1)
	if minScored >= 1 && !hasBadge(allCategoryBadges,badgeLevel1) {
		newBadges = append(newBadges, badgeLevel1)
	}
	badgeLevel2,_ = GetBadgeType(ScoreCategoryAll, 2)
	if minScored >= 3 && !hasBadge(allCategoryBadges,badgeLevel2) {
		newBadges = append(newBadges, badgeLevel2)
	}
	badgeLevel3,_ = GetBadgeType(ScoreCategoryAll, 3)
	if minScored >= 6 && !hasBadge(allCategoryBadges,badgeLevel3) {
		newBadges = append(newBadges, badgeLevel3)
	}

	for _,badge := range newBadges {
		badgeRecord := db.BadgeHistoryRecord { 
			CustomerCIF: cif,
			BadgeCode: badge.Code,
			DateAwarded: time.Now(),
		}
		err := h.BadgePutter(badgeRecord)
		if err != nil { return nil, err }
	}

	return newBadges, nil
}

func hasBadge(ownedBadges []BadgeType, badgeType BadgeType) bool {
	for _,badge := range ownedBadges {
		if badge.Code == badgeType.Code {
			return true;
		}
	}
	return false;
}