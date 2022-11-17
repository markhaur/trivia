package factlist

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/markhaur/trivia/pkg"
	"github.com/matryer/way"
)

func NewServer(service Service, logger log.Logger) http.Handler {
	s := server{service: service}

	var handleSaveFact http.Handler
	handleSaveFact = s.handleSaveFact()
	handleSaveFact = httpLoggingMiddleware(logger, "handleSaveFact")(handleSaveFact)

	var handleListFact http.Handler
	handleListFact = s.handleListFact()
	handleListFact = httpLoggingMiddleware(logger, "handleListFact")(handleListFact)

	var handleRemoveFact http.Handler
	handleRemoveFact = s.handleRemoveFact()
	handleRemoveFact = httpLoggingMiddleware(logger, "handleRemoveFact")(handleRemoveFact)

	var handleUpdateFact http.Handler
	handleUpdateFact = s.handleUpdateFact()
	handleUpdateFact = httpLoggingMiddleware(logger, "handleUpdateFact")(handleUpdateFact)

	router := way.NewRouter()

	router.Handle("POST", "/factlist/v1/fact", handleSaveFact)
	router.Handle("GET", "/factlist/v1/fact", handleListFact)
	router.Handle("DELETE", "/factlist/v1/fact/:id", handleRemoveFact)
	router.Handle("PUT", "factlist/v1/fact/:id", handleUpdateFact)

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, ErrResourceNotFound) })

	return router
}

const (
	contentTypeKey   = "Content-Type"
	contentTypeValue = "application/json; charset=utf-8"
)

var (
	ErrNonNumericFactID = errors.New("fact id must be numeric")
	ErrResourceNotFound = errors.New("resource not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

type ErrInvalidRequestBody struct{ err error }

func (e ErrInvalidRequestBody) Error() string { return fmt.Sprintf("invalid request body: %v", e.err) }

type server struct {
	service Service
}

func (s *server) handleSaveFact() http.HandlerFunc {
	type request struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	type response struct {
		ID        int64     `json:"id"`
		Question  string    `json:"question"`
		Answer    string    `json:"answer"`
		CreatedAt time.Time `json:"createdAt"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err})
			return
		}

		fact, err := s.service.Save(r.Context(), pkg.Fact{Question: req.Question, Answer: req.Answer})
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: fact.ID, Question: fact.Question, Answer: fact.Answer, CreatedAt: fact.CreatedAt})
	}
}

func (s *server) handleListFact() http.HandlerFunc {
	type fact struct {
		ID        int64     `json:"id"`
		Question  string    `json:"question"`
		Answer    string    `json:"answer"`
		CreatedAt time.Time `json:"createdAt"`
	}
	type response []fact
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.service.List(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}

		resp := make(response, 0, len(list))
		for _, v := range list {
			resp = append(resp, fact{ID: v.ID, Question: v.Question, Answer: v.Answer, CreatedAt: v.CreatedAt})
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *server) handleRemoveFact() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericFactID)
			return
		}

		if err := s.service.Remove(r.Context(), id); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) handleUpdateFact() http.HandlerFunc {
	type request struct {
		Question  string    `json:"Question"`
		Answer    string    `json:"answer"`
		CreatedAt time.Time `json:"createdAt"`
	}
	type response struct {
		ID        int64     `json:"id"`
		Question  string    `json:"question"`
		Answer    string    `json:"answer"`
		CreatedAt time.Time `json:"CreatedAt"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericFactID)
			return
		}

		var req request
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err: err})
			return
		}

		fact, isCreated, err := s.service.Update(r.Context(), pkg.Fact{ID: id, Question: req.Question, Answer: req.Answer, CreatedAt: req.CreatedAt})
		if err != nil {
			writeError(w, err)
			return
		}

		if isCreated {
			w.WriteHeader(http.StatusCreated)
		}

		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: fact.ID, Question: fact.Question, Answer: fact.Answer, CreatedAt: fact.CreatedAt})
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch err {
	case ErrResourceNotFound, pkg.ErrFactNotFound:
		w.WriteHeader(http.StatusNotFound)
	case pkg.ErrFactAlreadyExists:
		w.WriteHeader(http.StatusConflict)
	case ErrNonNumericFactID:
		w.WriteHeader(http.StatusBadRequest)
	case ErrMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		switch err.(type) {
		case ErrInvalidRequestBody:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func httpLoggingMiddleware(logger log.Logger, operation string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			lrw := &loggingResponseWriter{w, http.StatusOK}
			next.ServeHTTP(lrw, r)
			logger.Log(
				"operation", operation,
				"method", r.Method,
				"path", r.URL.Path,
				"took", time.Since(begin),
				"status", lrw.statusCode,
			)
		})
	}
}
