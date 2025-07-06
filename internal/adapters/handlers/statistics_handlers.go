package handlers

import (
	"net/http"
	"time"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// StatisticsHandlers encapsulates the dependencies for statistics-related HTTP handlers.
// It holds references to various service layers to fetch statistical data.
type StatisticsHandlers struct {
	memberService   *services.MemberService
	financeService  *services.FinanceService
	eventService    *services.EventService
	documentService *services.DocumentService
}

// NewStatisticsHandlers creates a new instance of StatisticsHandlers.
// It takes various service interfaces as dependencies, adhering to the dependency inversion principle.
func NewStatisticsHandlers(memberService *services.MemberService, financeService *services.FinanceService, eventService *services.EventService, documentService *services.DocumentService) *StatisticsHandlers {
	return &StatisticsHandlers{
		memberService:   memberService,
		financeService:  financeService,
		eventService:    eventService,
		documentService: documentService,
	}
}

// ShowDashboard displays the main dashboard page.
// It retrieves the authenticated user from the session and passes necessary data to the template.
func (h *StatisticsHandlers) ShowDashboard(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the dashboard template with user information, navbar, and a timestamp for cache busting.
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"title":      "Tableau de Bord",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"CurrentTimestamp": time.Now().Unix(),
	})
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowDashboard: %v", err)
	}
}

// GetMemberStats returns statistics related to members in JSON format.
// It fetches total member count and members grouped by status for the authenticated user.
func (h *StatisticsHandlers) GetMemberStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	// Fetch total members count.
	totalMembers, err := h.memberService.GetTotalMembersCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total de membres"})
		return
	}

	// Fetch members count by status.
	membersByStatus, err := h.memberService.GetMembersCountByStatus(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des membres par statut"})
		return
	}

	// Return member statistics as JSON.
	c.JSON(http.StatusOK, gin.H{
		"total_members":     totalMembers,
		"members_by_status": membersByStatus,
	})
}

// GetFinanceStats returns financial statistics in JSON format.
// It fetches total income, total expenses, and calculates net balance for the authenticated user.
func (h *StatisticsHandlers) GetFinanceStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	// Fetch total income.
	totalIncome, err := h.financeService.GetTotalIncome(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du total des revenus"})
		return
	}

	// Fetch total expenses.
	totalExpenses, err := h.financeService.GetTotalExpenses(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du total des dépenses"})
		return
	}

	// Return financial statistics as JSON.
	c.JSON(http.StatusOK, gin.H{
		"total_income":   totalIncome,
		"total_expenses": totalExpenses,
		"net_balance":    totalIncome - totalExpenses,
	})
}

// GetEventStats returns statistics related to events in JSON format.
// It fetches the total number of events for the authenticated user.
func (h *StatisticsHandlers) GetEventStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	// Fetch total events count.
	totalEvents, err := h.eventService.GetTotalEventsCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total d'événements"})
		return
	}

	// Return event statistics as JSON.
	c.JSON(http.StatusOK, gin.H{
		"total_events": totalEvents,
	})
}

// GetDocumentStats returns statistics related to documents in JSON format.
// It fetches the total number of documents for the authenticated user.
func (h *StatisticsHandlers) GetDocumentStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	// Fetch total documents count.
	totalDocuments, err := h.documentService.GetTotalDocumentsCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total de documents"})
		return
	}

	// Return document statistics as JSON.
	c.JSON(http.StatusOK, gin.H{
		"total_documents": totalDocuments,
	})
}
