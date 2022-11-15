package trivia

import (
	"context"
	"fmt"

	"github.com/markhaur/trivia/pkg"
)

// Service is an application service that lets us interact with a list of trivias
type Service interface {
	Save(context.Context, pkg.Trivia) (*pkg.Trivia, error)
	List(context.Context) ([]pkg.Trivia, error)
	Update(context.Context, pkg.Trivia) (*pkg.Trivia, bool, error)
	Remove(context.Context, int64) error
}

// Middleware describes a Service Middleware
type Middleware func(Service) Service

type service struct {
	repository pkg.TriviaRepository
}

func NewService(repository pkg.TriviaRepository) Service {
	return &service{repository: repository}
}

func (s *service) Save(ctx context.Context, trivia pkg.Trivia) (*pkg.Trivia, error) {
	if err := s.repository.Insert(ctx, &trivia); err != nil {
		return nil, fmt.Errorf("could not save trivia: %v", err)
	}
	return &trivia, nil
}

func (s *service) List(ctx context.Context) ([]pkg.Trivia, error) {
	list, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list all trivias: %v", err)
	}
	return list, nil
}

func (s *service) Update(ctx context.Context, trivia pkg.Trivia) (*pkg.Trivia, bool, error) {
	err := s.repository.Update(ctx, &trivia)
	if err == pkg.ErrTriviaNotFound {
		err = s.repository.Insert(ctx, &trivia)
		if err != nil {
			return nil, false, fmt.Errorf("could not create trivia: %v", err)
		}
		return &trivia, true, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("could not update task: %v", err)
	}
	return &trivia, false, nil
}

func (s *service) Remove(ctx context.Context, id int64) error {
	if err := s.repository.DeleteByID(ctx, id); err != nil {
		if err == pkg.ErrTriviaNotFound {
			return err
		}
		return fmt.Errorf("could not remove trivia: %v", err)
	}
	return nil
}
