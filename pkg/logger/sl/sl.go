package sl

import (
	"log/slog"
	"os"
	"sync"

	"github.com/tizzhh/micro-banking/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

var loggerSingleton *slog.Logger
var once sync.Once

func Get() *slog.Logger {
	once.Do(func() {
		loggerSingleton = setupLogger()
	})
	return loggerSingleton
}

func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func AddCaller(log *slog.Logger, caller string) *slog.Logger {
	return log.With(slog.String("caller", caller))
}

func AddRequestId(log *slog.Logger, reqID string) *slog.Logger {
	return log.With(slog.String("request_id", reqID))
}

func setupLogger() *slog.Logger {
	cfg := config.Get()
	switch cfg.Env {
	case envLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
