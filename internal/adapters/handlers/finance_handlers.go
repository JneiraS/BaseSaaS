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

// FinanceHandlers encapsulates the dependencies for financial HTTP handlers.
// It holds a reference to the FinanceService, which contains the business logic for financial operations.
type FinanceHandlers struct {
	financeService *services.FinanceService
}

// NewFinanceHandlers creates a new instance of FinanceHandlers.
// It takes a FinanceService as a dependency, adhering to the dependency inversion principle.
func NewFinanceHandlers(financeService *services.FinanceService) *FinanceHandlers {
	return &FinanceHandlers{financeService: financeService}
}

// ListTransactions displays a list of financial transactions for the authenticated user.
// It retrieves transactions from the FinanceService and renders them using the "transactions.tmpl" template.
func (h *FinanceHandlers) ListTransactions(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve transactions associated with the current user.
	transactions, err := h.financeService.GetTransactionsByUserID(user.ID)
	if err != nil {
		// Handle error, e.g., display an error message to the user.
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des transactions"})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the transactions list page.
	c.HTML(http.StatusOK, "transactions.tmpl", gin.H{
		"title":        "Mes Transactions",
		"navbar":       navbar,
		"user":         user,
		"transactions": transactions,
		"csrf_token":   csrfToken,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ListTransactions: %v", err)
	}
}

// ShowCreateTransactionForm displays the form for creating a new financial transaction.
// It provides a default date for convenience.
func (h *FinanceHandlers) ShowCreateTransactionForm(c *gin.Context) {
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

	// Render the transaction creation form.
	c.HTML(http.StatusOK, "transaction_form.tmpl", gin.H{
		"title":       "Ajouter une nouvelle transaction",
		"navbar":      navbar,
		"user":        user,
		"csrf_token":  csrfToken,
		"transaction": models.Transaction{Date: time.Now()}, // Default values
	})
	// Save session changes if any.
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowCreateTransactionForm: %v", err)
	}
}

// CreateTransaction handles the submission of the new transaction creation form.
// It binds the form data to a Transaction model, sets the UserID, and calls the service to create the transaction.
func (h *FinanceHandlers) CreateTransaction(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newTransaction models.Transaction
	// Bind form data to the newTransaction struct. If binding fails, return a bad request error.
	if err := c.ShouldBind(&newTransaction); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de transaction invalides: " + err.Error()})
		return
	}

	newTransaction.UserID = user.ID // Assign the current user's ID to the new transaction.

	// Call the service to create the transaction. Handle any errors during creation.
	if err := h.financeService.CreateTransaction(&newTransaction); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création de la transaction: " + err.Error()})
		return
	}

	// Redirect to the transactions list page upon successful creation.
	c.Redirect(http.StatusFound, "/finance/transactions")
}

// ShowEditTransactionForm displays the form for editing an existing financial transaction.
// It retrieves the transaction by ID, ensures it belongs to the authenticated user, and populates the form.
func (h *FinanceHandlers) ShowEditTransactionForm(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the transaction ID from the URL parameter.
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	// Retrieve the transaction from the service.
	transaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	// Verify that the transaction belongs to the authenticated user for security.
	if transaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the transaction edit form.
	c.HTML(http.StatusOK, "transaction_form.tmpl", gin.H{
		"title":       "Modifier la transaction",
		"navbar":      navbar,
		"user":        user,
		"csrf_token":  csrfToken,
		"transaction": transaction,
	})
	// Save session changes if any.
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowEditTransactionForm: %v", err)
	}
}

// UpdateTransaction handles the submission of the transaction modification form.
// It retrieves the existing transaction, binds updated data, ensures ownership, and calls the service to update.
func (h *FinanceHandlers) UpdateTransaction(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the transaction ID from the URL parameter.
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	var updatedTransaction models.Transaction
	// Bind form data to a temporary updatedTransaction struct.
	if err := c.ShouldBind(&updatedTransaction); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de transaction invalides: " + err.Error()})
		return
	}

	// Retrieve the existing transaction to ensure it belongs to the user before updating.
	existingTransaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	if existingTransaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Update the fields of the existing transaction with the new data from the form.
	existingTransaction.Amount = updatedTransaction.Amount
	existingTransaction.Type = updatedTransaction.Type
	existingTransaction.Description = updatedTransaction.Description
	existingTransaction.Date = updatedTransaction.Date

	// Call the service to update the transaction. Handle any errors during update.
	if err := h.financeService.UpdateTransaction(existingTransaction); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour de la transaction: " + err.Error()})
		return
	}

	// Redirect to the transactions list page upon successful update.
	c.Redirect(http.StatusFound, "/finance/transactions")
}

// DeleteTransaction handles the deletion of a financial transaction.
// It retrieves the transaction by ID, ensures it belongs to the authenticated user, and calls the service to delete it.
func (h *FinanceHandlers) DeleteTransaction(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the transaction ID from the URL parameter.
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de transaction invalide"})
		return
	}

	// Verify that the transaction belongs to the authenticated user before deletion.
	existingTransaction, err := h.financeService.GetTransactionByID(uint(transactionID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Transaction non trouvée"})
		return
	}

	if existingTransaction.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Call the service to delete the transaction. Handle any errors during deletion.
	if err := h.financeService.DeleteTransaction(uint(transactionID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression de la transaction: " + err.Error()})
		return
	}

	// Redirect to the transactions list page upon successful deletion.
	c.Redirect(http.StatusFound, "/finance/transactions")
}
