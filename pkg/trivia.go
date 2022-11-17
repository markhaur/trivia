package pkg

import (
	"context"
	"errors"
	"time"
)

var (
	ErrFactNotFound      = errors.New("trivia not found")
	ErrFactAlreadyExists = errors.New("trivia already exists")
)

type Fact struct {
	ID        int64
	Question  string
	Answer    string
	CreatedAt time.Time
}

// FactRepository is the interface used to persists the Fact(s)
type FactRepository interface {
	Insert(context.Context, *Fact) error
	FindAll(context.Context) ([]Fact, error)
	FindByID(context.Context, int64) (*Fact, error)
	Update(context.Context, *Fact) error
	DeleteByID(context.Context, int64) error
}
