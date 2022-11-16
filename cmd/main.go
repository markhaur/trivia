package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/log"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/markhaur/trivia/pkg/inmem"
	"github.com/markhaur/trivia/pkg/trivialist"
)

func main() {
	logger := log.NewJSONLogger(os.Stderr)
	defer logger.Log("msg", "terminated")

	path, found := os.LookupEnv("TRIVIA_CONFIG_PATH")
	if found {
		if err := godotenv.Load(path); err != nil {
			logger.Log("msg", "could not load .env file", "path", path, "err", err)
		}
	}

	var config struct {
		ServerAddress           string        `envconfig:"SERVER_ADDRESS" default:"localhost:8082"`
		ServerWriteTimeout      time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
		ServerReadTimeout       time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
		ServerIdleTimeout       time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
		GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT" default:"30s"`
		DBSource                string        `envconfig:"DB_SOURCE"`
		DBConnectTimeout        time.Duration `envconfig:"DB_CONNECT_TIMEOUT"`
	}
	if err := envconfig.Process("TRIVIAAPP", &config); err != nil {
		logger.Log("msg", "could not load env vars", "err", err)
		os.Exit(1)
	}

	trivias := inmem.NewTriviaRepository()

	var service trivialist.Service
	service = trivialist.NewService(trivias)
	service = trivialist.LoggingMiddleware(logger)(service)

	mux := http.NewServeMux()
	mux.Handle("/trivialist/v1/", trivialist.NewServer(service, logger))

	server := &http.Server{
		Addr:         config.ServerAddress,
		WriteTimeout: config.ServerWriteTimeout,
		ReadTimeout:  config.ServerReadTimeout,
		IdleTimeout:  config.ServerIdleTimeout,
		Handler:      mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		logger.Log("transport", "http", "address", config.ServerAddress, "msg", "listening")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log("transport", "http", "address", config.ServerAddress, "msg", "failed", "err", err)
			sig <- os.Interrupt
		}
	}()

	logger.Log("received", <-sig, "msg", "terminating")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown http server", "err", err)
	}

}
