package trivia

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTriviaNotFound      = errors.New("trivia not found")
	ErrTriviaAlreadyExists = errors.New("trivia already exists")
)

type Trivia struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

// TriviaRepository is the interface used to persists the Trivia(s)
type TriviaRepository interface {
	Insert(context.Context, *Trivia) error
	FindAll(context.Context) ([]Trivia, error)
	FindByID(context.Context, int64) (*Trivia, error)
	Update(context.Context, *Trivia) error
	DeleteByID(context.Context, int64) error
}
