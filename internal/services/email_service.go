package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/JneiraS/BaseSasS/internal/config"
)

// EmailService manages the sending of emails.
// It holds a reference to the application's configuration for SMTP settings.
type EmailService struct {
	cfg *config.Config
}

// NewEmailService creates a new instance of EmailService.
// It takes a Config struct as a dependency.
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendEmail sends a simple email using the configured SMTP settings.
// It constructs the email content, including an HTML template, and sends it to the specified recipients.
func (s *EmailService) SendEmail(to []string, subject, body string) error {
	from := s.cfg.EmailSender
	password := s.cfg.SMTPPassword

	// Load the HTML email template.
	tmpl, err := template.ParseFiles("templates/email_template.tmpl")
	if err != nil {
		return fmt.Errorf("erreur lors du chargement du template d'e-mail: %w", err)
	}

	// Prepare data to be injected into the HTML template.
	data := struct {
		Subject string
		Body    string
	}{
		Subject: subject,
		Body:    body,
	}

	// Execute the template and write the result into a buffer.
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return fmt.Errorf("erreur lors de l'exécution du template d'e-mail: %w", err)
	}

	// Use the generated HTML content as the email body.
	htmlBody := tpl.String()

	// Construct the email message with headers and HTML content.
	msg := []byte("To: " + to[0] + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		htmlBody)

	// Set up SMTP authentication.
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, password, s.cfg.SMTPHost)

	// Construct the SMTP server address.
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Send the email.
	return smtp.SendMail(addr, auth, from, to, msg)
}
