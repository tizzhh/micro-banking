package mail

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/tizzhh/micro-banking/internal/config"
)

type App struct {
	log *slog.Logger
}

func New(log *slog.Logger) *App {
	return &App{log: log}
}

func (a *App) SendMail(message string, to []string) error {
	const caller = "mail.SendMail"

	cfg := config.Get()
	auth := smtp.PlainAuth("", cfg.Mail.From, cfg.Mail.ApiKey, cfg.Mail.SmtpHost)

	err := smtp.SendMail(fmt.Sprintf("%s:%d", cfg.Mail.SmtpHost, cfg.Mail.SmtpPort), auth, cfg.Mail.From, to, []byte(message))
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}
	fmt.Println("Email Sent Successfully!")
	return nil
}
