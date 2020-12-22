package contactdetails

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cd "../../contactdetails"
	db "../../store"
	"../common"
	"github.com/stretchr/testify/assert"
)

var jsonDateFormat string = "2006-01-02T15:04:05Z"

func TestConfirmDirectDebits(t *testing.T) {
	testTime := time.Now()
	testCases := []struct {
		label string
		cifKey string
		currentScoreRecord *db.DynamicScoreRecord
		currentHistoryRecord *db.ScoreHistoryRecord
		expectedNewScoreRecord *db.DynamicScoreRecord
		expectedNewHistoryRecord *db.ScoreHistoryRecord
		expectedResponseCode int
		expectedResponseText string
	} {
		{ "Score date not updated if within a month",
			"4006001200", 
			&db.DynamicScoreRecord{ CustomerCIF: "4006001200", Score: 236 },
			&db.ScoreHistoryRecord{ CustomerCIF: "4006001200", LastConfirmed: testTime.AddDate(0, 0, -3), LastScored: testTime.AddDate(0, -1, 1), TimesConfirmed: 4, TimesScored: 2 },
			nil,
			&db.ScoreHistoryRecord{ CustomerCIF: "4006001200", LastConfirmed: testTime, LastScored: testTime.AddDate(0, -1, 1), TimesConfirmed: 5, TimesScored: 2 },
			200,
			`{"PointsGained":0,"NextPointsEligible":TIME_PLACEHOLDER,"NewBadges":[]}`,
		},
		{ "Record updated if outside a month",
			"4006079876", 
			&db.DynamicScoreRecord{ CustomerCIF: "4006079876", Score: 236 },
			&db.ScoreHistoryRecord{ CustomerCIF: "4006079876", LastConfirmed: testTime.AddDate(0, 0, -3), LastScored: testTime.AddDate(0, -1, -1), TimesConfirmed: 4, TimesScored: 2 },
			&db.DynamicScoreRecord{ CustomerCIF: "4006079876", Score: 336 },
			&db.ScoreHistoryRecord{ CustomerCIF: "4006079876", LastConfirmed: testTime, LastScored: testTime, TimesConfirmed: 5, TimesScored: 3 },
			200,
			`{"PointsGained":100,"NextPointsEligible":TIME_PLACEHOLDER,"NewBadges":[]}`,
		},
		{ "Record created if none exists",
			"4009998887", 
			nil,
			nil,
			&db.DynamicScoreRecord{ CustomerCIF: "4009998887", Score: 100 },
			&db.ScoreHistoryRecord{ CustomerCIF: "4009998887", LastConfirmed: testTime, LastScored: testTime, TimesConfirmed: 1, TimesScored: 1 },
			200,
			`{"PointsGained":100,"NextPointsEligible":TIME_PLACEHOLDER,"NewBadges":[]}`,
		},
	}

	for _,tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			mockScoreGetter := func(cif string) (db.DynamicScoreRecord, bool, error){
				assert.Equal(t, tc.cifKey, cif, "Should supply the CIF key to the Get query")
				if tc.currentScoreRecord != nil {
					return *tc.currentScoreRecord, true, nil
				} else {
					return db.DynamicScoreRecord{}, false, nil
				}
			}
			mockHistoryGetter := func(cif string, cat string) (db.ScoreHistoryRecord, bool, error){
				assert.Equal(t, tc.cifKey, cif, "Should supply the CIF key to the Get query")
				assert.Equal(t, "CD", cat, "Should supply the ContactDetails category code to the Get query")
				if tc.currentHistoryRecord != nil {
					return *tc.currentHistoryRecord, true, nil
				} else {
					return db.ScoreHistoryRecord{}, false, nil
				}
			}
			mockBadgeGetter := func(cif string) ([]db.BadgeHistoryRecord, error) { 
				return []db.BadgeHistoryRecord{	db.BadgeHistoryRecord{ BadgeCode: "CD1" }, db.BadgeHistoryRecord{ BadgeCode: "CD2" } }, nil
			}			
			mockHistoryGetAll := func(cif string) ([]db.ScoreHistoryRecord, error) { return []db.ScoreHistoryRecord{}, nil }
			var savedScoreRecord *db.DynamicScoreRecord
			mockScorePutter := func(record db.DynamicScoreRecord) error {
				savedScoreRecord = &record
				return nil
			}
			var savedHistoryRecord *db.ScoreHistoryRecord
			mockHistoryPutter := func(record db.ScoreHistoryRecord) error {
				savedHistoryRecord = &record
				return nil
			}
			mockBadgePutter := func(rec db.BadgeHistoryRecord) error { return nil }

			testHandler := ContactDetailsHandler { 
				ConfirmationHandler: common.ConfirmationHandler {
					ScoreGetter: mockScoreGetter,
					ScorePutter: mockScorePutter,
					CategoryGetter: mockHistoryGetter,
					CategoryPutter: mockHistoryPutter,
					CategoryGetAll: mockHistoryGetAll,
					BadgeGetter: mockBadgeGetter,
					BadgePutter: mockBadgePutter,
				},
				provider: mockContactDetailsProvider{},
				requestAuthenticator: func(*http.Request) (string, error) { return tc.cifKey, nil },
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/contactdetails?cif=" + tc.cifKey, nil)
			testHandler.ConfirmContactDetails(w, r)
			result := w.Result()

			var expectedNextEligible time.Time
			if tc.expectedNewScoreRecord == nil {
				assert.Nil(t, savedScoreRecord, "No save should be performed on Score")
			} else {
				assert.NotNil(t, savedScoreRecord, "A save should be performed on Score")
				assert.Equal(t, tc.expectedNewScoreRecord.CustomerCIF, savedScoreRecord.CustomerCIF, "Saved record CIF key")
				assert.Equal(t, tc.expectedNewScoreRecord.Score, savedScoreRecord.Score, "Saved record score")
			}
			if tc.expectedNewHistoryRecord == nil {
				assert.Nil(t, savedHistoryRecord, "No save should be performed on History")
				expectedNextEligible = tc.currentHistoryRecord.LastScored.AddDate(0,1,0)
			} else {
				assert.NotNil(t, savedHistoryRecord, "A save should be performed on History")
				assert.Equal(t, tc.expectedNewHistoryRecord.CustomerCIF, savedHistoryRecord.CustomerCIF, "Saved record CIF key")
				assert.Equal(t, tc.expectedNewHistoryRecord.TimesConfirmed, savedHistoryRecord.TimesConfirmed, "Saved record confirm count")
				assert.Equal(t, tc.expectedNewHistoryRecord.TimesScored, savedHistoryRecord.TimesScored, "Saved record score count")
				assert.WithinDuration(t, tc.expectedNewHistoryRecord.LastConfirmed, savedHistoryRecord.LastConfirmed, time.Millisecond * time.Duration(100), "Saved record last confirm date")
				assert.WithinDuration(t, tc.expectedNewHistoryRecord.LastScored, savedHistoryRecord.LastScored, time.Millisecond * time.Duration(100), "Saved record last scored date")
				expectedNextEligible = savedHistoryRecord.LastScored.AddDate(0,1,0)
			}
			assert.Equal(t, tc.expectedResponseCode, result.StatusCode, "Response code")
			body,err := ioutil.ReadAll(result.Body)
			assert.Nil(t, err, "Unhandled error reading result")
			jsonDate,_ := json.Marshal(expectedNextEligible)
			assert.Equal(t, strings.Replace(tc.expectedResponseText, "TIME_PLACEHOLDER", string(jsonDate), 1) + "\n", string(body), "Response body")
		})
	}
}

type mockContactDetailsProvider struct {}

func (mockContactDetailsProvider) GetContactDetails(cif string) (details cd.ContactDetails, err error){
	return cd.BuildContactDetails (
		cif,
		"Freda",
		"Flintstone",
		"07777123456",
		"01617731234",
		"freda@theflintstones.co.uk",
		cd.BuildAddress("Building 1", "Think Park", "", "Mosley Road", "Trafford Park", "Manchester", "Lancashire", "M17 1FQ"),
	), nil
}

func (mockContactDetailsProvider) SaveEmailAddress(cif string, newEmailAddress string) error { return nil }
func (mockContactDetailsProvider) SaveMobileNumber(cif string, newMobileNumber string) error { return nil }
func (mockContactDetailsProvider) SaveHomeNumber(cif string, newHomeNumber string) error { return nil }
func (mockContactDetailsProvider) SaveAddress(cif string, newAddress cd.Address) error { return nil }
