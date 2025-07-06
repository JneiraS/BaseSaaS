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

// MemberHandlers encapsule les dépendances pour les handlers des membres.
type MemberHandlers struct {
	memberService *services.MemberService
}

// NewMemberHandlers crée une nouvelle instance de MemberHandlers.
func NewMemberHandlers(memberService *services.MemberService) *MemberHandlers {
	return &MemberHandlers{memberService: memberService}
}

// ListMembers affiche la liste des membres.
func (h *MemberHandlers) ListMembers(c *gin.Context) {
	// Récupérer l'utilisateur connecté depuis la session
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Récupérer les membres associés à cet utilisateur
	members, err := h.memberService.GetMembersByUserID(user.ID)
	if err != nil {
		// Gérer l'erreur, par exemple, afficher un message d'erreur
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des membres"})
		return
	}

	// Récupérer le jeton CSRF pour la navbar
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "members.tmpl", gin.H{
		"title":   "Mes Membres",
		"navbar":  navbar,
		"user":    user,
		"members": members,
	})
}

// ShowCreateMemberForm affiche le formulaire de création d'un nouveau membre.
func (h *MemberHandlers) ShowCreateMemberForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "member_form.tmpl", gin.H{
		"title":      "Ajouter un nouveau membre",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"member":     models.Member{MembershipStatus: models.StatusActive, JoinDate: time.Now()}, // Valeurs par défaut
	})
}

// CreateMember gère la soumission du formulaire de création de membre.
func (h *MemberHandlers) CreateMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newMember models.Member
	// Bind le formulaire à la structure Member
	if err := c.ShouldBind(&newMember); err != nil {
		// Gérer l'erreur de binding (ex: données invalides)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de membre invalides: " + err.Error()})
		return
	}

	// Assigner l'ID de l'utilisateur connecté au membre
	newMember.UserID = user.ID

	// Appeler le service pour créer le membre
	if err := h.memberService.CreateMember(&newMember); err != nil {
		// Gérer l'erreur de création
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création du membre: " + err.Error()})
		return
	}

	// Rediriger vers la liste des membres après succès
	c.Redirect(http.StatusFound, "/members")
}

// ShowEditMemberForm affiche le formulaire de modification d'un membre existant.
func (h *MemberHandlers) ShowEditMemberForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	member, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return	
	}

	// Vérifier que le membre appartient bien à l'utilisateur connecté
	if member.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "member_form.tmpl", gin.H{
		"title":      "Modifier le membre",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"member":     member,
	})
}

// UpdateMember gère la soumission du formulaire de modification de membre.
func (h *MemberHandlers) UpdateMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	var updatedMember models.Member
	if err := c.ShouldBind(&updatedMember); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données de membre invalides: " + err.Error()})
		return
	}

	// Récupérer le membre existant pour s'assurer qu'il appartient à l'utilisateur
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Mettre à jour les champs du membre existant avec les données du formulaire
	existingMember.FirstName = updatedMember.FirstName
	existingMember.LastName = updatedMember.LastName
	existingMember.Email = updatedMember.Email
	existingMember.MembershipStatus = updatedMember.MembershipStatus
	existingMember.JoinDate = updatedMember.JoinDate
	existingMember.EndDate = updatedMember.EndDate

	if err := h.memberService.UpdateMember(existingMember); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour du membre: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/members")
}

// DeleteMember gère la suppression d'un membre.
func (h *MemberHandlers) DeleteMember(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// Vérifier que le membre appartient bien à l'utilisateur connecté avant de supprimer
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	if err := h.memberService.DeleteMember(uint(memberID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression du membre: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/members")
}

// MarkPayment gère le marquage d'un paiement pour un membre.
func (h *MemberHandlers) MarkPayment(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de membre invalide"})
		return
	}

	// Vérifier que le membre appartient bien à l'utilisateur connecté avant de marquer le paiement
	existingMember, err := h.memberService.GetMemberByID(uint(memberID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Membre non trouvé"})
		return
	}

	if existingMember.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Marquer le paiement avec la date actuelle
	if err := h.memberService.MarkPaymentReceived(uint(memberID), time.Now()); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors du marquage du paiement: " + err.Error()})
		return
	}

	// Rediriger vers la liste des membres après succès
	c.Redirect(http.StatusFound, "/members")
}
