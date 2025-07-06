package services

import (
	"fmt"
	"net/smtp"

	"github.com/JneiraS/BaseSasS/internal/config"
)

// EmailService gère l'envoi d'e-mails.
type EmailService struct {
	cfg *config.Config
}

// NewEmailService crée une nouvelle instance de EmailService.
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendEmail envoie un e-mail simple.
func (s *EmailService) SendEmail(to []string, subject, body string) error {
	from := s.cfg.EmailSender
	password := s.cfg.SMTPPassword

	msg := []byte("To: " + to[0] + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, password, s.cfg.SMTPHost)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	return smtp.SendMail(addr, auth, from, to, msg)
}
