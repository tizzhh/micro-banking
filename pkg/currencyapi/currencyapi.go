package currencyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/tizzhh/micro-banking/internal/config"
	currencyapihttp "github.com/tizzhh/micro-banking/pkg/currencyapi/domain/http"
	"github.com/tizzhh/micro-banking/pkg/logger/sl"
)

type Api struct {
	log        *slog.Logger
	HttpClient http.Client
}

func New(log *slog.Logger, timeout time.Duration) *Api {
	return &Api{
		log:        log,
		HttpClient: http.Client{Timeout: timeout},
	}
}

const (
	urlTemplate = "%s?apikey=%s&currencies=%s"
)

func (a *Api) QueryRates(ctx context.Context, currencyCode string) (float32, error) {
	const caller = "currencyapi.QueryRates"

	log := sl.AddCaller(a.log, caller)

	log.Info("querying rate", slog.String("currency", currencyCode))

	cfg := config.Get()

	queryUrl := fmt.Sprintf(urlTemplate, cfg.CurrencyApi.URL, cfg.CurrencyApi.ApiKey, currencyCode)
	resp, err := a.HttpClient.Get(queryUrl)
	if err != nil {
		log.Error("failed to query rates", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read response body", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	var response currencyapihttp.Response
	if err = json.Unmarshal(resBody, &response); err != nil {
		log.Error("failed to unmarshal response body", sl.Error(err))
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	rates, exists := response.Data.Currencies[currencyCode]
	if !exists {
		log.Error("currency code missing in response body")
		return 0, fmt.Errorf("%s: %w", caller, err)
	}

	log.Info("queried rates", slog.String("currency", currencyCode), slog.String("last_updated", response.Meta.LastUpdated))

	return rates.Value, nil
}
