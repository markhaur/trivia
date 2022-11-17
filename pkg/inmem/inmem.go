package inmem

import (
	"context"
	"sync"

	"github.com/markhaur/trivia/pkg"
)

type triviaRepository struct {
	sync.RWMutex
	trivialist []pkg.Fact
	exists     map[int64]bool
	counter    int64
}

func NewFactRepository() pkg.FactRepository {
	return &triviaRepository{trivialist: []pkg.Fact{}, exists: make(map[int64]bool)}
}

func (tr *triviaRepository) Insert(_ context.Context, newFact *pkg.Fact) error {
	tr.Lock()
	defer tr.Unlock()

	if newFact.ID > 0 {
		if tr.exists[newFact.ID] {
			return pkg.ErrFactAlreadyExists
		}

		tr.exists[newFact.ID] = true
		tr.trivialist = append(tr.trivialist, *newFact)
		return nil
	}

	tr.counter++
	for tr.exists[tr.counter] {
		tr.counter++
	}
	tr.exists[tr.counter] = true
	newFact.ID = tr.counter
	tr.trivialist = append(tr.trivialist, *newFact)
	return nil
}

func (tr *triviaRepository) FindAll(_ context.Context) ([]pkg.Fact, error) {
	tr.RLock()
	defer tr.RUnlock()

	return tr.trivialist, nil
}

func (tr *triviaRepository) FindByID(_ context.Context, id int64) (*pkg.Fact, error) {
	tr.RLock()
	defer tr.RUnlock()

	for _, t := range tr.trivialist {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, pkg.ErrFactNotFound
}

func (tr *triviaRepository) Update(_ context.Context, updatedFact *pkg.Fact) error {
	tr.Lock()
	defer tr.Unlock()

	for i, t := range tr.trivialist {
		if t.ID == updatedFact.ID {
			tr.trivialist[i] = *updatedFact
			return nil
		}
	}
	return pkg.ErrFactNotFound
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
	return pkg.ErrFactNotFound
}
