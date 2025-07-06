package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// PollHandlers encapsule les dépendances pour les handlers de sondages.
type PollHandlers struct {
	pollService *services.PollService
}

// NewPollHandlers crée une nouvelle instance de PollHandlers.
func NewPollHandlers(pollService *services.PollService) *PollHandlers {
	return &PollHandlers{pollService: pollService}
}

// ListPolls affiche la liste de tous les sondages.
func (h *PollHandlers) ListPolls(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	polls, err := h.pollService.GetAllPolls()
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des sondages: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des sondages."})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "polls.tmpl", gin.H{
		"title":      "Sondages",
		"navbar":     navbar,
		"user":       user,
		"polls":      polls,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ListPolls: %v", err)
	}
}

// ShowCreatePollForm affiche le formulaire de création d'un nouveau sondage.
func (h *PollHandlers) ShowCreatePollForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "poll_form.tmpl", gin.H{
		"title":      "Créer un nouveau sondage",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"poll":       models.Poll{}, // Sondage vide pour le formulaire
	})
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ShowCreatePollForm: %v", err)
	}
}

// CreatePoll gère la soumission du formulaire de création de sondage.
func (h *PollHandlers) CreatePoll(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newPoll models.Poll
	if err := c.ShouldBind(&newPoll); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire de sondage: %v", err)
		session.AddFlash("Données de sondage invalides: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls/new")
		return
	}

	newPoll.UserID = user.ID

	// Récupérer les options du formulaire (elles sont envoyées sous forme de champs séparés)
	options := c.PostFormArray("options")
	for _, optText := range options {
		if strings.TrimSpace(optText) != "" {
			newPoll.Options = append(newPoll.Options, models.Option{Text: optText})
		}
	}

	if err := h.pollService.CreatePoll(&newPoll); err != nil {
		log.Printf("ERREUR: Erreur lors de la création du sondage: %v", err)
		session.AddFlash("Erreur lors de la création du sondage: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls/new")
		return
	}

	session.AddFlash("Sondage créé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/polls")
}

// ShowPollDetails affiche les détails d'un sondage et permet de voter.
func (h *PollHandlers) ShowPollDetails(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de sondage invalide"})
		return
	}

	poll, err := h.pollService.GetPollByID(uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Sondage non trouvé: %v", err)
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Sondage non trouvé"})
		return
	}

	// Vérifier si l'utilisateur a déjà voté pour ce sondage
	hasVoted, err := h.pollService.HasUserVoted(user.ID, uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la vérification du vote: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la vérification du vote."})
		return
	}

	// Récupérer les résultats du sondage
	results, err := h.pollService.GetPollResults(uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des résultats du sondage: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des résultats du sondage."})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "poll_details.tmpl", gin.H{
		"title":      poll.Question,
		"navbar":     navbar,
		"user":       user,
		"poll":       poll,
		"has_voted":  hasVoted,
		"results":    results,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ShowPollDetails: %v", err)
	}
}

// VoteOnPoll gère la soumission d'un vote.
func (h *PollHandlers) VoteOnPoll(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		session.AddFlash("ID de sondage invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	optionID, err := strconv.ParseUint(c.PostForm("option_id"), 10, 64)
	if err != nil {
		session.AddFlash("Option de vote invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
		return
	}

	if err := h.pollService.Vote(uint(optionID), user.ID, uint(pollID)); err != nil {
		log.Printf("ERREUR: Échec du vote: %v", err)
		session.AddFlash("Échec du vote: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
		return
	}

	session.AddFlash("Votre vote a été enregistré avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
}

// DeletePoll gère la suppression d'un sondage.
func (h *PollHandlers) DeletePoll(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		session.AddFlash("ID de sondage invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	// Vérifier que l'utilisateur est le propriétaire du sondage
	poll, err := h.pollService.GetPollByID(uint(pollID))
	if err != nil {
		session.AddFlash("Sondage non trouvé.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	if poll.UserID != user.ID {
		session.AddFlash("Accès non autorisé à la suppression de ce sondage.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	if err := h.pollService.DeletePoll(uint(pollID)); err != nil {
		log.Printf("ERREUR: Échec de la suppression du sondage: %v", err)
		session.AddFlash("Échec de la suppression du sondage: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	session.AddFlash("Sondage supprimé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/polls")
}
