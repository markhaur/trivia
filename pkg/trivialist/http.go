package trivialist

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

	var handleSaveTrivia http.Handler
	handleSaveTrivia = s.handleSaveTrivia()
	handleSaveTrivia = httpLoggingMiddleware(logger, "handleSaveTrivia")(handleSaveTrivia)

	var handleListTrivia http.Handler
	handleListTrivia = s.handleListTrivia()
	handleListTrivia = httpLoggingMiddleware(logger, "handleListTrivia")(handleListTrivia)

	var handleRemoveTrivia http.Handler
	handleRemoveTrivia = s.handleRemoveTrivia()
	handleRemoveTrivia = httpLoggingMiddleware(logger, "handleRemoveTrivia")(handleRemoveTrivia)

	var handleUpdateTrivia http.Handler
	handleUpdateTrivia = s.handleUpdateTrivia()
	handleUpdateTrivia = httpLoggingMiddleware(logger, "handleUpdateTrivia")(handleUpdateTrivia)

	router := way.NewRouter()

	router.Handle("POST", "/trivialist/v1/trivia", handleSaveTrivia)
	router.Handle("GET", "/trivialist/v1/trivia", handleListTrivia)
	router.Handle("DELETE", "/trivialist/v1/trivia/:id", handleRemoveTrivia)
	router.Handle("PUT", "trivialist/v1/trivia/:id", handleUpdateTrivia)

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, ErrResourceNotFound) })

	return router
}

const (
	contentTypeKey   = "Content-Type"
	contentTypeValue = "application/json; charset=utf-8"
)

var (
	ErrNonNumericTriviaID = errors.New("trivia id must be numeric")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrMethodNotAllowed   = errors.New("method not allowed")
)

type ErrInvalidRequestBody struct{ err error }

func (e ErrInvalidRequestBody) Error() string { return fmt.Sprintf("invalid request body: %v", e.err) }

type server struct {
	service Service
}

func (s *server) handleSaveTrivia() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err})
			return
		}

		trivia, err := s.service.Save(r.Context(), pkg.Trivia{Name: req.Name})
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: trivia.ID, Name: trivia.Name, CreatedAt: trivia.CreatedAt})
	}
}

func (s *server) handleListTrivia() http.HandlerFunc {
	type trivia struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}
	type response []trivia
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.service.List(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}

		resp := make(response, 0, len(list))
		for _, v := range list {
			resp = append(resp, trivia{ID: v.ID, Name: v.Name, CreatedAt: v.CreatedAt})
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *server) handleRemoveTrivia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericTriviaID)
			return
		}

		if err := s.service.Remove(r.Context(), id); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) handleUpdateTrivia() http.HandlerFunc {
	type request struct {
		Name      string    `json:"Name"`
		CreatedAt time.Time `json:"createdAt"`
	}
	type response struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"CreatedAt"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericTriviaID)
			return
		}

		var req request
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err: err})
			return
		}

		trivia, isCreated, err := s.service.Update(r.Context(), pkg.Trivia{ID: id, Name: req.Name, CreatedAt: req.CreatedAt})
		if err != nil {
			writeError(w, err)
			return
		}

		if isCreated {
			w.WriteHeader(http.StatusCreated)
		}

		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: trivia.ID, Name: trivia.Name, CreatedAt: trivia.CreatedAt})
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch err {
	case ErrResourceNotFound, pkg.ErrTriviaNotFound:
		w.WriteHeader(http.StatusNotFound)
	case pkg.ErrTriviaAlreadyExists:
		w.WriteHeader(http.StatusConflict)
	case ErrNonNumericTriviaID:
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
