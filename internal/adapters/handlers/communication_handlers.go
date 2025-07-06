package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CommunicationHandlers encapsulates the dependencies for communication-related HTTP handlers.
// It holds references to the EmailService for sending emails and MemberService for retrieving member email addresses.
type CommunicationHandlers struct {
	emailService  *services.EmailService
	memberService *services.MemberService // To retrieve member email addresses
}

// NewCommunicationHandlers creates a new instance of CommunicationHandlers.
// It takes EmailService and MemberService as dependencies.
func NewCommunicationHandlers(emailService *services.EmailService, memberService *services.MemberService) *CommunicationHandlers {
	return &CommunicationHandlers{
		emailService:  emailService,
		memberService: memberService,
	}
}

// ShowEmailForm displays the email sending form.
// It retrieves the authenticated user from the session and renders the "email_form.tmpl" template.
func (h *CommunicationHandlers) ShowEmailForm(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the email form page.
	c.HTML(http.StatusOK, "email_form.tmpl", gin.H{
		"title":      "Envoyer un e-mail aux membres",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
}

// SendEmailToMembers handles the submission of the email sending form.
// It retrieves the subject and body from the form, fetches all member emails for the authenticated user,
// and sends the email using the EmailService.
func (h *CommunicationHandlers) SendEmailToMembers(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get email subject and body from the form submission.
	subject := c.PostForm("subject")
	body := c.PostForm("body")

	// Retrieve all members for the authenticated user to get their email addresses.
	members, err := h.memberService.GetMembersByUserID(user.ID)
	if err != nil {
		log.Printf("ERREUR: Impossible de récupérer les membres pour l'envoi d'e-mail: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des membres."})
		return
	}

	var recipientEmails []string
	for _, member := range members {
		recipientEmails = append(recipientEmails, member.Email)
	}

	// If no members are found, add a warning flash message and redirect.
	if len(recipientEmails) == 0 {
		log.Printf("INFO: Aucun membre trouvé pour l'envoi d'e-mail.")
		session.AddFlash("Aucun membre trouvé pour envoyer l'e-mail.", "warning")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Send the email using the EmailService.
	if err := h.emailService.SendEmail(recipientEmails, subject, body); err != nil {
		log.Printf("ERREUR: Échec de l'envoi de l'e-mail: %v", err)
		session.AddFlash("Échec de l'envoi de l'e-mail: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Add a success flash message and redirect upon successful email sending.
	log.Printf("INFO: E-mail envoyé avec succès à %d membres.", len(recipientEmails))
	session.AddFlash(fmt.Sprintf("E-mail envoyé avec succès à %d membres.", len(recipientEmails)), "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/profile")
}
