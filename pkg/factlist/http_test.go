package factlist_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/markhaur/trivia/pkg"
	"github.com/markhaur/trivia/pkg/factlist"
	"github.com/markhaur/trivia/pkg/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTasks(t *testing.T) {
	tt := []struct {
		Name        string
		FactsInRepo []pkg.Fact
		Expected    string
	}{
		{
			Name:        "Returns 200 and empty list if no facts exist",
			FactsInRepo: []pkg.Fact{},
			Expected:    `[]`,
		},
		{
			Name: "Returns 200 and 3 facts if 3 facts exists",
			FactsInRepo: []pkg.Fact{
				{ID: 1, Question: "what is my github username?", Answer: "markhaur"},
				{ID: 2, Question: "what's your favourite language?", Answer: "Go"},
				{ID: 3, Question: "what's the name of current project?", Answer: "trivia"},
			},
			Expected: `[
				{"id": 1, "question": "what is my github username?", "answer": "markhaur", "createdAt":"0001-01-01T00:00:00Z"},
				{"id": 2, "question": "what's your favourite language?", "answer": "Go", "createdAt":"0001-01-01T00:00:00Z"},
				{"id": 3, "question": "what's the name of current project?", "answer": "trivia", "createdAt":"0001-01-01T00:00:00Z"}
			]`,
		},
		{
			Name: "Returns 200 and 1 fact if 1 fact exists",
			FactsInRepo: []pkg.Fact{
				{ID: 1, Question: "what is my github username?", Answer: "markhaur"},
			},
			Expected: `[
				{"id": 1, "question": "what is my github username?", "answer": "markhaur", "createdAt":"0001-01-01T00:00:00Z"}
			]`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = factlist.NewService(inmem.NewFactRepository())
				handler = factlist.NewServer(svc, log.NewNopLogger())
			)

			for _, fact := range tc.FactsInRepo {
				_, err := svc.Save(context.TODO(), fact)
				require.NoError(err, "could not save fact")
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/factlist/v1/fact", nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(http.StatusOK, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.Expected, rec.Body.String(), "unexpected http response body")
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tt := []struct {
		Name             string
		ReqBody          string
		FactID           string
		ExpectedQuestion string
		ExpectedAnswer   string
		ExpectedCode     int
		ExpectedRspBody  string
	}{
		{
			Name:             "Returns 200 and updates fact for valid request",
			ReqBody:          `{"question": "what is your username?", "answer": "markhaur"}`,
			FactID:           "1",
			ExpectedQuestion: "what is your username?",
			ExpectedAnswer:   "markhaur",
			ExpectedCode:     http.StatusOK,
			ExpectedRspBody:  `{"id": 1, "question": "what is your username?", "answer": "markhaur", "createdAt":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:             "Returns 201 and creates fact for valid request if it doesn't exist",
			ReqBody:          `{"question": "what is your username?", "answer": "markhaur"}`,
			FactID:           "1337",
			ExpectedQuestion: "what is your username?",
			ExpectedAnswer:   "markhaur",
			ExpectedCode:     http.StatusCreated,
			ExpectedRspBody:  `{"id": 1337, "question": "what is your username?", "answer": "markhaur", "createdAt":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:             "Returns 400 and error msg for non-numeric id",
			ReqBody:          `{"question": "what is your username?", "answer": "markhaur"}`,
			FactID:           ":p",
			ExpectedQuestion: "what is your username?",
			ExpectedAnswer:   "markhaur",
			ExpectedCode:     http.StatusBadRequest,
			ExpectedRspBody:  `{"error": "fact id must be numeric"}`,
		},
		{
			Name:             "Returns 400 and error msg for invalid json",
			ReqBody:          `<?!{"question": "what is your username?", "answer": "markhaur"}`,
			FactID:           "1",
			ExpectedQuestion: "what is your username?",
			ExpectedAnswer:   "markhaur",
			ExpectedCode:     http.StatusBadRequest,
			ExpectedRspBody:  `{"error": "invalid request body: invalid character '<' looking for beginning of value"}`,
		},
		{
			Name:             "Returns 400 and error msg for blank question",
			ReqBody:          `{"question": "	", "answer": "empty"}`,
			FactID:           "1",
			ExpectedQuestion: "what is your username?",
			ExpectedAnswer:   "markhaur",
			ExpectedCode:     http.StatusBadRequest,
			ExpectedRspBody:  `{"error": "invalid request body: invalid character '\\t' in string literal"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = factlist.NewService(inmem.NewFactRepository())
				handler = factlist.NewServer(svc, log.NewNopLogger())
			)

			newFact := pkg.Fact{Question: "what is your username?", Answer: "markhaur"}
			_, err := svc.Save(context.TODO(), newFact)
			require.NoError(err, "could not save fact")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/factlist/v1/fact/%s", tc.FactID)
			req, err := http.NewRequest("PUT", url, strings.NewReader(tc.ReqBody))
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), "unexpected http response body")

			list, err := svc.List(context.TODO())
			require.NoError(err, "could not list facts")

			if tc.ExpectedCode != 200 && tc.ExpectedCode != 201 {
				return
			}

			for _, fact := range list {
				if strconv.FormatInt(fact.ID, 10) == tc.FactID {
					assert.Equal(tc.ExpectedQuestion, fact.Question, "expected Question to be updated")
					assert.Equal(tc.ExpectedAnswer, fact.Answer, "expected Answer to be updated")
					return
				}
			}
			t.Error("could not find fact after calling handler")
		})
	}
}
