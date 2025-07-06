package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// FinanceHandlers encapsule les dépendances pour les handlers financiers.
type FinanceHandlers struct {
	financeService *services.FinanceService
}

// NewFinanceHandlers crée une nouvelle instance de FinanceHandlers.
func NewFinanceHandlers(financeService *services.FinanceService) *FinanceHandlers {
	return &FinanceHandlers{financeService: financeService}
}

// ListTransactions affiche la liste des transactions.
func (h *FinanceHandlers) ListTransactions(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	transactions, err := h.financeService.GetTransactionsByUserID(user.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des transactions"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "transactions.tmpl", gin.H{
		"title":        "Mes Transactions",
		"navbar":       navbar,
		"user":         user,
		"transactions": transactions,
		"csrf_token":   csrfToken,
	})
}

// ShowCreateTransactionForm affiche le formulaire de création d'une nouvelle transaction.
func (h *FinanceHandlers) ShowCreateTransactionForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "transaction_form.tmpl", gin.H{
		"title":       "Ajouter une nouvelle transaction",
		"navbar":      navbar,
		"user":        user,
		"csrf_token":  csrfToken,
		"transaction": models.Transaction{Date: time.Now()}, // Valeurs par défaut
	})
}

// CreateTransaction gère la soumission du formulaire de création de transaction.
func (h *FinanceHandlers) CreateTransaction(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newTransaction models.Transaction
	if err := c.ShouldBind(&newTransaction); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de transaction invalides: " + err.Error()})
		return
	}

	newTransaction.UserID = user.ID

	if err := h.financeService.CreateTransaction(&newTransaction); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création de la transaction: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/finance/transactions")
}

// ShowEditTransactionForm affiche le formulaire de modification d'une transaction existante.
func (h *FinanceHandlers) ShowEditTransactionForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	transaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	// Vérifier que la transaction appartient bien à l'utilisateur connecté
	if transaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "transaction_form.tmpl", gin.H{
		"title":       "Modifier la transaction",
		"navbar":      navbar,
		"user":        user,
		"csrf_token":  csrfToken,
		"transaction": transaction,
	})
}

// UpdateTransaction gère la soumission du formulaire de modification de transaction.
func (h *FinanceHandlers) UpdateTransaction(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	var updatedTransaction models.Transaction
	if err := c.ShouldBind(&updatedTransaction); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de transaction invalides: " + err.Error()})
		return
	}

	// Récupérer la transaction existante pour s'assurer qu'elle appartient à l'utilisateur
	existingTransaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	if existingTransaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Mettre à jour les champs de la transaction existante avec les données du formulaire
	existingTransaction.Amount = updatedTransaction.Amount
	existingTransaction.Type = updatedTransaction.Type
	existingTransaction.Description = updatedTransaction.Description
	existingTransaction.Date = updatedTransaction.Date

	if err := h.financeService.UpdateTransaction(existingTransaction); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour de la transaction: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/finance/transactions")
}

// DeleteTransaction gère la suppression d'une transaction.
func (h *FinanceHandlers) DeleteTransaction(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	// Vérifier que la transaction appartient bien à l'utilisateur connecté avant de supprimer
	existingTransaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	if existingTransaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	if err := h.financeService.DeleteTransaction(uint(transactionID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression de la transaction: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/finance/transactions")
}
