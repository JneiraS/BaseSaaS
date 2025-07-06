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

// MemberHandlers encapsulates the dependencies for member-related HTTP handlers.
// It holds a reference to the MemberService, which contains the business logic for members.
type MemberHandlers struct {
	memberService *services.MemberService
}

// NewMemberHandlers creates a new instance of MemberHandlers.
// It takes a MemberService as a dependency, adhering to the dependency inversion principle.
func NewMemberHandlers(memberService *services.MemberService) *MemberHandlers {
	return &MemberHandlers{memberService: memberService}
}

// ListMembers displays a list of members for the authenticated user.
// It retrieves members from the MemberService and renders them using the "members.tmpl" template.
func (h *MemberHandlers) ListMembers(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve members associated with the current user.
	members, err := h.memberService.GetMembersByUserID(user.ID)
	if err != nil {
		// Handle error, e.g., display an error message to the user.
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des membres"})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the members list page.
	c.HTML(http.StatusOK, "members.tmpl", gin.H{
		"title":      "Mes Membres",
		"navbar":     navbar,
		"user":       user,
		"members":    members,
		"csrf_token": csrfToken, // Add CSRF token to the template context
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ListMembers: %v", err)
	}
}

// ShowCreateMemberForm displays the form for creating a new member.
// It provides default values for membership status and join date for convenience.
func (h *MemberHandlers) ShowCreateMemberForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the member creation form.
	c.HTML(http.StatusOK, "member_form.tmpl", gin.H{
		"title":      "Ajouter un nouveau membre",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"member":     models.Member{MembershipStatus: models.StatusActive, JoinDate: time.Now()}, // Default values
	})
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowCreateMemberForm: %v", err)
	}
}

// CreateMember handles the submission of the new member creation form.
// It binds the form data to a Member model, sets the UserID, and calls the service to create the member.
func (h *MemberHandlers) CreateMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newMember models.Member
	// Bind form data to the newMember struct. If binding fails, return a bad request error.
	if err := c.ShouldBind(&newMember); err != nil {
		// Handle binding error (e.g., invalid data)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de membre invalides: " + err.Error()})
		return
	}

	// Assign the current user's ID to the new member.
	newMember.UserID = user.ID

	// Call the service to create the member.
	if err := h.memberService.CreateMember(&newMember); err != nil {
		// Handle creation error
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création du membre: " + err.Error()})
		return
	}

	// Redirect to the members list page upon success.
	c.Redirect(http.StatusFound, "/members")
}

// ShowEditMemberForm displays the form for editing an existing member.
// It retrieves the member by ID, ensures it belongs to the authenticated user, and populates the form.
func (h *MemberHandlers) ShowEditMemberForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the member ID from the URL parameter.
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// Retrieve the member from the service.
	member, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	// Verify that the member belongs to the authenticated user for security.
	if member.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the member edit form.
	c.HTML(http.StatusOK, "member_form.tmpl", gin.H{
		"title":      "Modifier le membre",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"member":     member,
	})
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowEditMemberForm: %v", err)
	}
}

// UpdateMember handles the submission of the member modification form.
// It retrieves the existing member, binds updated data, ensures ownership, and calls the service to update.
func (h *MemberHandlers) UpdateMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the member ID from the URL parameter.
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// 1. Retrieve the existing member from the database.
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	// 2. Verify that the member belongs to the authenticated user.
	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// 3. Bind the form data to a new struct for validation.
	var formMember models.Member
	if err := c.ShouldBind(&formMember); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de membre invalides: " + err.Error()})
		return
	}

	// 4. Update the fields of the existing member with the new data from the form.
	existingMember.FirstName = formMember.FirstName
	existingMember.LastName = formMember.LastName
	existingMember.Email = formMember.Email
	existingMember.MembershipStatus = formMember.MembershipStatus
	existingMember.JoinDate = formMember.JoinDate
	existingMember.EndDate = formMember.EndDate
	existingMember.LastPaymentDate = formMember.LastPaymentDate

	// 5. Call the service to save the updated member.
	if err := h.memberService.UpdateMember(existingMember); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour du membre: " + err.Error()})
		return
	}

	// Redirect to the members list page upon successful update.
	c.Redirect(http.StatusFound, "/members")
}

// DeleteMember handles the deletion of a member.
// It retrieves the member by ID, ensures it belongs to the authenticated user, and calls the service to delete it.
func (h *MemberHandlers) DeleteMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the member ID from the URL parameter.
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// Verify that the member belongs to the authenticated user before deletion.
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Call the service to delete the member. Handle any errors during deletion.
	if err := h.memberService.DeleteMember(uint(memberID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression du membre: " + err.Error()})
		return
	}

	// Redirect to the members list page upon successful deletion.
	c.Redirect(http.StatusFound, "/members")
}

// MarkPayment handles marking a payment for a member.
// It retrieves the member by ID, ensures it belongs to the authenticated user, and calls the service to update the payment status.
func (h *MemberHandlers) MarkPayment(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the member ID from the URL parameter.
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// Verify that the member belongs to the authenticated user before marking payment.
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Mark the payment with the current date.
	if err := h.memberService.MarkPaymentReceived(uint(memberID), time.Now()); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors du marquage du paiement: " + err.Error()})
		return
	}

	// Redirect to the members list page upon successful payment marking.
	c.Redirect(http.StatusFound, "/members")
}
