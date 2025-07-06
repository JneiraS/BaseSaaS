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

// PollHandlers encapsulates the dependencies for poll-related HTTP handlers.
// It holds a reference to the PollService, which contains the business logic for polls.
type PollHandlers struct {
	pollService *services.PollService
}

// NewPollHandlers creates a new instance of PollHandlers.
// It takes a PollService as a dependency, adhering to the dependency inversion principle.
func NewPollHandlers(pollService *services.PollService) *PollHandlers {
	return &PollHandlers{pollService: pollService}
}

// ListPolls displays a list of all polls.
// It retrieves polls from the PollService and renders them using the "polls.tmpl" template.
func (h *PollHandlers) ListPolls(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve all polls from the service.
	polls, err := h.pollService.GetAllPolls()
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des sondages: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des sondages."})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the polls list page.
	c.HTML(http.StatusOK, "polls.tmpl", gin.H{
		"title":      "Sondages",
		"navbar":     navbar,
		"user":       user,
		"polls":      polls,
		"csrf_token": csrfToken,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ListPolls: %v", err)
	}
}

// ShowCreatePollForm displays the form for creating a new poll.
// It provides an empty poll model for the form.
func (h *PollHandlers) ShowCreatePollForm(c *gin.Context) {
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

	// Render the poll creation form.
	c.HTML(http.StatusOK, "poll_form.tmpl", gin.H{
		"title":      "Créer un nouveau sondage",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"poll":       models.Poll{}, // Empty poll for the form
	})
	// Save session changes if any.
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ShowCreatePollForm: %v", err)
	}
}

// CreatePoll handles the submission of the new poll creation form.
// It binds the form data, extracts options, sets the UserID, and calls the service to create the poll.
func (h *PollHandlers) CreatePoll(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newPoll models.Poll
	// Bind the form data to the newPoll struct. If binding fails, add a flash message and redirect.
	if err := c.ShouldBind(&newPoll); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire de sondage: %v", err)
		session.AddFlash("Données de sondage invalides: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls/new")
		return
	}

	newPoll.UserID = user.ID // Assign the current user's ID to the new poll.

	// Retrieve poll options from the form (sent as separate fields).
	options := c.PostFormArray("options")
	for _, optText := range options {
		if strings.TrimSpace(optText) != "" {
			newPoll.Options = append(newPoll.Options, models.Option{Text: optText})
		}
	}

	// Call the service to create the poll. Handle any errors during creation.
	if err := h.pollService.CreatePoll(&newPoll); err != nil {
		log.Printf("ERREUR: Erreur lors de la création du sondage: %v", err)
		session.AddFlash("Erreur lors de la création du sondage: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls/new")
		return
	}

	// Add a success flash message and redirect to the polls list page.
	session.AddFlash("Sondage créé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/polls")
}

// ShowPollDetails displays the details of a specific poll and allows users to vote.
// It retrieves the poll, checks if the user has voted, and fetches poll results.
func (h *PollHandlers) ShowPollDetails(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the poll ID from the URL parameter.
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de sondage invalide"})
		return
	}

	// Retrieve the poll from the service.
	poll, err := h.pollService.GetPollByID(uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Sondage non trouvé: %v", err)
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Sondage non trouvé"})
		return
	}

	// Check if the user has already voted for this poll.
	hasVoted, err := h.pollService.HasUserVoted(user.ID, uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la vérification du vote: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la vérification du vote."})
		return
	}

	// Retrieve the poll results.
	results, err := h.pollService.GetPollResults(uint(pollID))
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des résultats du sondage: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des résultats du sondage."})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the poll details page.
	c.HTML(http.StatusOK, "poll_details.tmpl", gin.H{
		"title":      poll.Question,
		"navbar":     navbar,
		"user":       user,
		"poll":       poll,
		"has_voted":  hasVoted,
		"results":    results,
		"csrf_token": csrfToken,
	})
	// Save session changes if any.
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de session dans ShowPollDetails: %v", err)
	}
}

// VoteOnPoll handles the submission of a user's vote for a poll option.
// It validates the poll and option IDs, checks if the user has already voted,
// and records the vote via the PollService.
func (h *PollHandlers) VoteOnPoll(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the poll ID from the URL parameter.
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		session.AddFlash("ID de sondage invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	// Parse the selected option ID from the form data.
	optionID, err := strconv.ParseUint(c.PostForm("option_id"), 10, 64)
	if err != nil {
		session.AddFlash("Option de vote invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
		return
	}

	// Call the service to record the vote. Handle any errors (e.g., already voted, invalid option).
	if err := h.pollService.Vote(uint(optionID), user.ID, uint(pollID)); err != nil {
		log.Printf("ERREUR: Échec du vote: %v", err)
		session.AddFlash("Échec du vote: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
		return
	}

	// Add a success flash message and redirect to the poll details page.
	session.AddFlash("Votre vote a été enregistré avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/polls/%d", pollID))
}

// DeletePoll handles the deletion of a poll.
// It retrieves the poll by ID, ensures it belongs to the authenticated user, and calls the service to delete it.
func (h *PollHandlers) DeletePoll(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the poll ID from the URL parameter.
	pollID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		session.AddFlash("ID de sondage invalide.", "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	// Verify that the user is the owner of the poll before allowing deletion.
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

	// Call the service to delete the poll. Handle any errors during deletion.
	if err := h.pollService.DeletePoll(uint(pollID)); err != nil {
		log.Printf("ERREUR: Échec de la suppression du sondage: %v", err)
		session.AddFlash("Échec de la suppression du sondage: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.Redirect(http.StatusFound, "/polls")
		return
	}

	// Add a success flash message and redirect to the polls list page.
	session.AddFlash("Sondage supprimé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/polls")
}
