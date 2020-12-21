package contactdetails

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "../../store"
	"github.com/stretchr/testify/assert"
)

var jsonDateFormat string = "2006-01-02T15:04:05Z"

func TestConfirmDirectDebits(t *testing.T) {
	testTime := time.Now()
	testCases := []struct {
		label string
		cifKey string
		currentRecord *db.DynamicScoreRecord
		expectedNewRecord *db.DynamicScoreRecord
		expectedResponseCode int
		expectedResponseText string
	} {
		{ "Record not updated if within a month",
			"4006001200", 
			&db.DynamicScoreRecord{ CustomerCIF: "4006001200", LastUpdatedContactDetails: testTime.AddDate(0,-1,1), Score: 236 },
			nil,
			200,
			`{"PointsGained":0,"NextPointsEligible":TIME_PLACEHOLDER}`,
		},
		{ "Record updated if outside a month",
			"4006079876", 
			&db.DynamicScoreRecord{ CustomerCIF: "4006079876", LastUpdatedContactDetails: testTime.AddDate(0,-1,-1), Score: 236 },
			&db.DynamicScoreRecord{ CustomerCIF: "4006079876", LastUpdatedContactDetails: testTime, Score: 336 },
			200,
			`{"PointsGained":100,"NextPointsEligible":TIME_PLACEHOLDER}`,
		},
		{ "Record created if none exists",
			"4009998887", 
			nil,
			&db.DynamicScoreRecord{ CustomerCIF: "4009998887", LastUpdatedContactDetails: testTime, Score: 100 },
			200,
			`{"PointsGained":100,"NextPointsEligible":TIME_PLACEHOLDER}`,
		},
	}

	for _,tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			mockGetter := func(cif string) (db.DynamicScoreRecord, bool, error){
				assert.Equal(t, tc.cifKey, cif, "Should supply the CIF key to the Get query")
				if tc.currentRecord != nil {
					return *tc.currentRecord, true, nil
				} else {
					return db.DynamicScoreRecord{}, false, nil
				}
			}
			var savedRecord *db.DynamicScoreRecord
			mockPutter := func(record db.DynamicScoreRecord) error {
				savedRecord = &record
				return nil
			}

			testHandler := ContactDetailsHandler { 
				scoreGetter: mockGetter,
				scorePutter: mockPutter,
				contactDetailsGetter: GetDummyContactDetails,
				requestAuthenticator: func(*http.Request) (string, error) { return tc.cifKey, nil },
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/contactdetails?cif=" + tc.cifKey, nil)
			testHandler.ConfirmContactDetails(w, r)
			result := w.Result()

			var expectedNextEligible time.Time
			if tc.expectedNewRecord == nil {
				assert.Nil(t, savedRecord, "No save should be performed")
				expectedNextEligible = tc.currentRecord.LastUpdatedContactDetails.AddDate(0,1,0)
			} else {
				assert.NotNil(t, savedRecord, "A save should be performed")
				assert.Equal(t, tc.expectedNewRecord.CustomerCIF, savedRecord.CustomerCIF, "Saved record CIF key")
				assert.Equal(t, tc.expectedNewRecord.Score, savedRecord.Score, "Saved record score")
				assert.WithinDuration(t, tc.expectedNewRecord.LastUpdatedContactDetails, savedRecord.LastUpdatedContactDetails, time.Millisecond * time.Duration(100), "Saved record last updated date")
				expectedNextEligible = savedRecord.LastUpdatedContactDetails.AddDate(0,1,0)
			}
			assert.Equal(t, tc.expectedResponseCode, result.StatusCode, "Response code")
			body,err := ioutil.ReadAll(result.Body)
			assert.Nil(t, err, "Unhandled error reading result")
			jsonDate,_ := json.Marshal(expectedNextEligible)
			assert.Equal(t, strings.Replace(tc.expectedResponseText, "TIME_PLACEHOLDER", string(jsonDate), 1) + "\n", string(body), "Response body")
		})
	}
}