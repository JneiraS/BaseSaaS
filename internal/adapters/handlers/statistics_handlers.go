package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// StatisticsHandlers encapsule les dépendances pour les handlers de statistiques.
type StatisticsHandlers struct {
	memberService   *services.MemberService
	financeService  *services.FinanceService
	eventService    *services.EventService
	documentService *services.DocumentService
}

// NewStatisticsHandlers crée une nouvelle instance de StatisticsHandlers.
func NewStatisticsHandlers(memberService *services.MemberService, financeService *services.FinanceService, eventService *services.EventService, documentService *services.DocumentService) *StatisticsHandlers {
	return &StatisticsHandlers{
		memberService:   memberService,
		financeService:  financeService,
		eventService:    eventService,
		documentService: documentService,
	}
}

// ShowDashboard affiche la page du tableau de bord.
func (h *StatisticsHandlers) ShowDashboard(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"title":      "Tableau de Bord",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		// Gérer l'erreur de sauvegarde de session si nécessaire
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowDashboard: %v", err)
	}
}

// GetMemberStats retourne les statistiques sur les membres.
func (h *StatisticsHandlers) GetMemberStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	totalMembers, err := h.memberService.GetTotalMembersCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total de membres"})
		return
	}

	membersByStatus, err := h.memberService.GetMembersCountByStatus(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des membres par statut"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_members":     totalMembers,
		"members_by_status": membersByStatus,
	})
}

// GetFinanceStats retourne les statistiques financières.
func (h *StatisticsHandlers) GetFinanceStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	totalIncome, err := h.financeService.GetTotalIncome(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du total des revenus"})
		return
	}

	totalExpenses, err := h.financeService.GetTotalExpenses(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du total des dépenses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_income":   totalIncome,
		"total_expenses": totalExpenses,
		"net_balance":    totalIncome - totalExpenses,
	})
}

// GetEventStats retourne les statistiques sur les événements.
func (h *StatisticsHandlers) GetEventStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	totalEvents, err := h.eventService.GetTotalEventsCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total d'événements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_events": totalEvents,
	})
}

// GetDocumentStats retourne les statistiques sur les documents.
func (h *StatisticsHandlers) GetDocumentStats(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	totalDocuments, err := h.documentService.GetTotalDocumentsCount(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du nombre total de documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_documents": totalDocuments,
	})
}
