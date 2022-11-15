package trivia

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/markhaur/trivia/pkg"
	"github.com/matryer/way"
)

func NewServer(service Service, logger log.Logger) http.Handler {
	s := server{service: service}
}

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
			// TODO: need to validate by @jarri-abidi
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		trivia, err := s.service.Save(r.Context(), pkg.Trivia{Name: req.Name})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := make(response, 0, len(list))
		for _, v := range list {
			resp = append(resp, trivia{ID: v.ID, Name: v.Name, CreatedAt: v.CreatedAt})
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *server) handleRemoveTrivia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id", 10, 64))
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
		id, err := strconv.ParseInt(way.Param("id"), 10, 64)
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
