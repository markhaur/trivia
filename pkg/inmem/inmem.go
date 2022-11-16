package inmem

import (
	"context"
	"sync"

	"github.com/markhaur/trivia/pkg"
)

type triviaRepository struct {
	sync.RWMutex
	trivialist []pkg.Trivia
	exists     map[int64]bool
	counter    int64
}

func NewTriviaRepository() pkg.TriviaRepository {
	return &triviaRepository{trivialist: []pkg.Trivia{}}
}

func (tr *triviaRepository) Insert(_ context.Context, newTrivia *pkg.Trivia) error {
	tr.Lock()
	defer tr.Unlock()

	if newTrivia.ID > 0 {
		if tr.exists[newTrivia.ID] {
			return pkg.ErrTriviaAlreadyExists
		}

		tr.exists[newTrivia.ID] = true
		tr.trivialist = append(tr.trivialist, *newTrivia)
		return nil
	}

	tr.counter++
	for tr.exists[tr.counter] {
		tr.counter++
	}
	tr.exists[tr.counter] = true
	newTrivia.ID = tr.counter
	tr.trivialist = append(tr.trivialist, *newTrivia)
	return nil
}

func (tr *triviaRepository) FindAll(_ context.Context) ([]pkg.Trivia, error) {
	tr.RLock()
	defer tr.RUnlock()

	return tr.trivialist, nil
}

func (tr *triviaRepository) FindByID(_ context.Context, id int64) (*pkg.Trivia, error) {
	tr.RLock()
	defer tr.RUnlock()

	for _, t := range tr.trivialist {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, pkg.ErrTriviaNotFound
}

func (tr *triviaRepository) Update(_ context.Context, updatedTrivia *pkg.Trivia) error {
	tr.Lock()
	defer tr.Unlock()

	for i, t := range tr.trivialist {
		if t.ID == updatedTrivia.ID {
			tr.trivialist[i] = *updatedTrivia
			return nil
		}
	}
	return pkg.ErrTriviaNotFound
}

func (tr *triviaRepository) DeleteByID(_ context.Context, id int64) error {
	tr.Lock()
	defer tr.Unlock()

	for i, t := range tr.trivialist {
		if t.ID == id {
			tr.trivialist = append(tr.trivialist[:i], tr.trivialist[i+1:]...)
			return nil
		}
	}
	return pkg.ErrTriviaNotFound
}
