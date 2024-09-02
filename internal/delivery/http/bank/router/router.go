package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/tizzhh/micro-banking/internal/api/permissions"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/auth"
	bankApi "github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/bank"
	"github.com/tizzhh/micro-banking/internal/delivery/http/bank/resource/currency"
	authentication "github.com/tizzhh/micro-banking/internal/delivery/http/bank/router/middleware/auth"
	mwLogger "github.com/tizzhh/micro-banking/internal/delivery/http/bank/router/middleware/logger"
	"github.com/tizzhh/micro-banking/internal/services/bank"
	"github.com/tizzhh/micro-banking/pkg/jwt"
)

func New(
	log *slog.Logger,
	validator *validator.Validate,
	authClient auth.AuthClient,
	currencyClient currency.CurrencyClient,
	tokenTTL time.Duration,
	bank *bank.Bank,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	authApi := auth.New(log, validator, authClient)

	jwt := jwt.New(tokenTTL)
	permissionChecker := permissions.New(jwt)

	router.Get("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))).ServeHTTP)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/docs/swagger.json"),
	))

	router.Route("/v1/auth", func(r chi.Router) {
		r.Method(http.MethodPost, "/register", authApi.NewUser())
		r.Method(http.MethodPost, "/login", authApi.LoginUser())

		r.Group(func(r chi.Router) {
			r.Use(authentication.AuthenticateUser(log, permissionChecker))

			r.Method(http.MethodPut, "/change-password", authApi.UpdatePassword())
			r.Method(http.MethodDelete, "/unregister", authApi.DeleteUser())
			r.Method(http.MethodGet, "/user", authApi.User())
		})
	})

	currencyApi := currency.New(log, validator, currencyClient)
	bankApi := bankApi.New(log, validator, bank)

	router.Route("/v1/bank", func(r chi.Router) {
		r.Use(authentication.AuthenticateUser(log, permissionChecker))

		r.Method(http.MethodGet, "/my-wallet", currencyApi.MyWallet())
		r.Method(http.MethodPost, "/deposit", bankApi.Deposit())
		r.Method(http.MethodPost, "/withdraw", bankApi.Withdraw())

		r.Route("/currency", func(r chi.Router) {
			r.Method(http.MethodPost, "/buy", currencyApi.BuyCurrency())
			r.Method(http.MethodPost, "/sell", currencyApi.SellCurrency())
		})
	})

	router.Route("/v1", func(r chi.Router) {
		r.Method(http.MethodGet, "/liveness", bankApi.Liveness())
	})

	return router
}
