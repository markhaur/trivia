package trivia

import (
	"context"
	"log"
	"time"

	"github.com/markhaur/trivia/pkg"
)

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(s Service) Service { return &loggingMiddleware{logger, s} }
}

type loggingMiddleware struct {
	logger log.Logger
	Service
}

func (s *loggingMiddleware) Save(ctx context.Context, trivia pkg.Trivia) (_ *pkg.Trivia, err error) {
	defer func(begin time.Time) {
		s.logger.Printf(
			"method", "save",
			"name", trivia.Name,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.Save(ctx, trivia)
}

func (s *loggingMiddleware) List(ctx context.Context) (_ []pkg.Trivia, err error) {
	defer func(begin time.Time) {
		s.logger.Printf(
			"method", "list",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.List(ctx)
}

func (s *loggingMiddleware) Remove(ctx context.Context, id int64) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf(
			"method", "remove",
			"id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.Remove(ctx, id)
}

func (s *loggingMiddleware) Update(ctx context.Context, trivia pkg.Trivia) (err error) {
	defer func(begin time.Time) {
		s.logger.Printf(
			"method", "update",
			"name", trivia.Name,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.Update(ctx, trivia)
}
