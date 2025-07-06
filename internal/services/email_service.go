package services

import (
	"bytes"
	"fmt"
	"html/template"
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

	// Charger le template HTML
	tmpl, err := template.ParseFiles("templates/email_template.tmpl")
	if err != nil {
		return fmt.Errorf("erreur lors du chargement du template d'e-mail: %w", err)
	}

	// Préparer les données pour le template
	data := struct {
		Subject string
		Body    string
	}{
		Subject: subject,
		Body:    body,
	}

	// Exécuter le template et écrire le résultat dans un buffer
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return fmt.Errorf("erreur lors de l'exécution du template d'e-mail: %w", err)
	}

	// Utiliser le contenu HTML généré comme corps de l'e-mail
	htmlBody := tpl.String()

	msg := []byte("To: " + to[0] + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		htmlBody)

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, password, s.cfg.SMTPHost)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	return smtp.SendMail(addr, auth, from, to, msg)
}
