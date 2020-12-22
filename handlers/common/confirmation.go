package common

import (
	"time"

	db "../../store"
)

type ConfirmationResponse struct {
	PointsGained int
	NextPointsEligible time.Time
}

// NOT READY FOR USE YET - hardcoded to Direct Debit type. Needs updating once the new UserCategoryScore table or whatever its called is ready
type ConfirmationHandler struct {
	scoreGetter ScoreGetter
	scorePutter ScorePutter
}

func (h *ConfirmationHandler) ConfirmCategory(cif string, category ScoreCategory) (resp ConfirmationResponse, err error) {
	resp = ConfirmationResponse{}
	score, scoreFound, err := h.scoreGetter(cif)
	if err != nil { return }

	if !scoreFound {
		score = db.DynamicScoreRecord{
			CustomerCIF: cif,
			LastUpdatedDirectDebits: time.Now(),
			Score: 100,
		}
		err = h.scorePutter(score)
		if err != nil { return }
		return ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) }, nil
	} else if score.LastUpdatedDirectDebits.AddDate(0, 1, 0).Before(time.Now()) {
		score.Score += 100
		score.LastUpdatedDirectDebits = time.Now()
		err = h.scorePutter(score)
		if err != nil { return }
		return ConfirmationResponse { 100, time.Now().AddDate(0, 1, 0) }, nil
	} else {
		return ConfirmationResponse { 0, score.LastUpdatedDirectDebits.AddDate(0, 1, 0) }, nil
	}
}