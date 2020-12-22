package userscorehandler

import (
	"fmt"
	"net/http"
	"time"

	"../../respond"
	"../../store"
	"../common"
)

type UserScoreHandler struct {
	scoreGetter common.ScoreGetter
	allScoreGetter common.AllScoreGetter
	categoryGetter common.CategoryScoreGetAll
	badgeGetter common.BadgeGetter
	requestAuthenticator func(r *http.Request) (cifKey string, err error) 
}

type UserScoreResponse struct {
	CustomerCIF string
	Score int
	Position int
	IsJointPosition bool
	PointsBehindNext int
	Categories []UserCategoryScore
	Badges []common.BadgeType
}

type UserCategoryScore struct {
	Category common.ScoreCategory
	LastConfirmedDateTime time.Time
	LastScoredDateTime time.Time
	ConfirmationCount int
	ScoreCount int
}

func NewHandler() UserScoreHandler {
	scoreStore, err := db.DefaultDynamicScoreStore()
	if(err != nil) { panic(err) }
	categoryStore,err := db.DefaultScoreHistoryStore()
	if(err != nil) { panic(err) }
	badgeStore,err := db.DefaultBadgeHistoryStore()
	if(err != nil) { panic(err) }

	return UserScoreHandler{
		scoreGetter: scoreStore.Get,
		allScoreGetter: scoreStore.GetAllScores,
		categoryGetter: categoryStore.GetAll,
		badgeGetter: badgeStore.Get,
		requestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}

func (h *UserScoreHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if(r.Method != http.MethodGet) { 
		respond.WithError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	cif, err := h.requestAuthenticator(r)
	if err != nil {
		respond.WithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	record, _, err := h.scoreGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	allScores, err := h.allScoreGetter()
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, err.Error());
		return
	}
	allCategories, err := h.categoryGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting category scores for %s: %s", cif, err.Error()));
		return
	}
	allBadges, err := h.badgeGetter(cif)
	if err != nil {
		respond.WithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting badges for %s: %s", cif, err.Error()));
		return
	}	

	response := UserScoreResponse {
		CustomerCIF: cif,
		Score: record.Score,
		Position: 1,
		Categories: []UserCategoryScore {},
	}

	for _,cat := range allCategories {
		response.Categories = append(response.Categories, UserCategoryScore {
			Category: common.ScoreCategoryLookup[cat.CategoryCode],
			LastConfirmedDateTime: cat.LastConfirmed,
			LastScoredDateTime: cat.LastScored,
			ConfirmationCount: cat.TimesConfirmed,
			ScoreCount: cat.TimesScored,
		})
	}

	for _,badge := range allBadges {
		response.Badges = append(response.Badges, common.BadgeTypeLookup[badge.BadgeCode])
	}

	joints := 0
	for _,s := range allScores {
		if s > response.Score {
			response.Position ++
		} else if s == response.Score {
			joints++
			response.IsJointPosition = (joints > 1)
		}
	}

	respond.WithJSON(w, http.StatusOK, response)
}