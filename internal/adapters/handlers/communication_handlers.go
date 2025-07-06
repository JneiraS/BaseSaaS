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

// CommunicationHandlers encapsule les dépendances pour les handlers de communication.
type CommunicationHandlers struct {
	emailService  *services.EmailService
	memberService *services.MemberService // Pour récupérer les adresses e-mail des membres
}

// NewCommunicationHandlers crée une nouvelle instance de CommunicationHandlers.
func NewCommunicationHandlers(emailService *services.EmailService, memberService *services.MemberService) *CommunicationHandlers {
	return &CommunicationHandlers{
		emailService:  emailService,
		memberService: memberService,
	}
}

// ShowEmailForm affiche le formulaire d'envoi d'e-mail.
func (h *CommunicationHandlers) ShowEmailForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "email_form.tmpl", gin.H{
		"title":      "Envoyer un e-mail aux membres",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
}

// SendEmailToMembers gère l'envoi d'e-mails aux membres.
func (h *CommunicationHandlers) SendEmailToMembers(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	subject := c.PostForm("subject")
	body := c.PostForm("body")

	// Récupérer tous les membres de l'utilisateur connecté
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

	if len(recipientEmails) == 0 {
		log.Printf("INFO: Aucun membre trouvé pour l'envoi d'e-mail.")
		session.AddFlash("Aucun membre trouvé pour envoyer l'e-mail.", "warning")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Envoyer l'e-mail
	if err := h.emailService.SendEmail(recipientEmails, subject, body); err != nil {
		log.Printf("ERREUR: Échec de l'envoi de l'e-mail: %v", err)
		session.AddFlash("Échec de l'envoi de l'e-mail: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	log.Printf("INFO: E-mail envoyé avec succès à %d membres.", len(recipientEmails))
	session.AddFlash(fmt.Sprintf("E-mail envoyé avec succès à %d membres.", len(recipientEmails)), "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/profile")
}
