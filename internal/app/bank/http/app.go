package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/auth"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/router"
	"github.com/tizzhh/micro-banking/internal/services/bank"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type App struct {
	log             *slog.Logger
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	router          *chi.Mux
	port            int
	shutdownTimeout time.Duration
	server          *http.Server
}

func New(
	log *slog.Logger,
	readTimeout, writeTimeout, idleTimeout time.Duration,
	authClient auth.AuthClient,
	currencyClient currency.CurrencyClient,
	port int,
	shutdownTimeout time.Duration,
	tokenTTL time.Duration,
	bank *bank.Bank,
) *App {
	router := router.New(log, validator.New(), authClient, currencyClient, tokenTTL, bank)
	return &App{
		log:             log,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		idleTimeout:     idleTimeout,
		router:          router,
		port:            port,
		shutdownTimeout: shutdownTimeout,
	}
}

func (a *App) MustRun() {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.port),
		Handler:      a.router,
		ReadTimeout:  a.readTimeout,
		WriteTimeout: a.writeTimeout,
		IdleTimeout:  a.idleTimeout,
	}
	a.server = server

	a.log.Info("starting bank server")

	if err := server.ListenAndServe(); err != nil {
		a.log.Error("shutdown", sl.Error(err))
	}
}

func (a *App) Stop() {
	a.log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failed to stop server", sl.Error(err))
		return
	}
}
