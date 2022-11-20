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
	"github.com/markhaur/trivia/pkg/factlist"
	"github.com/markhaur/trivia/pkg/inmem"
)

func main() {
	logger := log.NewJSONLogger(os.Stderr)
	defer logger.Log("msg", "terminated")

	path, found := os.LookupEnv("TRIVIAAPP_CONFIG_PATH")
	if found {
		if err := godotenv.Load(path); err != nil {
			logger.Log("msg", "could not load .env file", "path", path, "err", err)
		}
	}

	var config struct {
		ServerAddress           string        `envconfig:"TRIVIA_SERVER_ADDRESS" default:"localhost:8082"`
		ServerWriteTimeout      time.Duration `envconfig:"TRIVIA_SERVER_WRITE_TIMEOUT" default:"15s"`
		ServerReadTimeout       time.Duration `envconfig:"TRIVIA_SERVER_READ_TIMEOUT" default:"15s"`
		ServerIdleTimeout       time.Duration `envconfig:"TRIVIA_SERVER_IDLE_TIMEOUT" default:"60s"`
		GracefulShutdownTimeout time.Duration `envconfig:"TRIVIA_GRACEFUL_SHUTDOWN_TIMEOUT" default:"30s"`
		DBSource                string        `envconfig:"TRIVIA_DB_SOURCE"`
		DBConnectTimeout        time.Duration `envconfig:"TRIVIA_DB_CONNECT_TIMEOUT"`
	}
	if err := envconfig.Process("TRIVIAAPP", &config); err != nil {
		logger.Log("msg", "could not load env vars", "err", err)
		os.Exit(1)
	}

	trivias := inmem.NewFactRepository()

	var service factlist.Service
	service = factlist.NewService(trivias)
	service = factlist.LoggingMiddleware(logger)(service)

	mux := http.NewServeMux()
	mux.Handle("/factlist/v1/", factlist.NewServer(service, logger))

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
