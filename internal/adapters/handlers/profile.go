package handlers

import (
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/database"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// Page profil (protégée)
func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	user, ok := session.Get("user").(models.User)
	if !ok {
		// Si l'utilisateur n'est pas en session ou n'est pas du bon type, rediriger vers la connexion
		c.Redirect(http.StatusFound, "/login")
		return
	}
	csrfToken := csrf.GetToken(c)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":      "Profil",
		"user":       user,
		"navbar":     navbar,
		"csrf_token": csrfToken,
	})
}

// UpdateProfileHandler gère la mise à jour du profil utilisateur.
func UpdateProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	loggedInUser, ok := session.Get("user").(models.User)
	if !ok {
		log.Printf("DEBUG: Utilisateur non trouvé en session ou typage incorrect.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var updatedUser models.User
	if err := c.ShouldBind(&updatedUser); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire: %v", err)
		session.AddFlash("Erreur lors de la lecture des données du formulaire.", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	log.Printf("DEBUG: Données reçues du formulaire - Nom: %s, Email: %s", updatedUser.Name, updatedUser.Email)
	log.Printf("DEBUG: Utilisateur avant mise à jour - Nom: %s, Email: %s, OIDCID: %s", loggedInUser.Name, loggedInUser.Email, loggedInUser.OIDCID)

	// Mettre à jour les champs modifiables
	loggedInUser.Name = updatedUser.Name
	loggedInUser.Email = updatedUser.Email

	log.Printf("DEBUG: Utilisateur après modification locale - Nom: %s, Email: %s, OIDCID: %s", loggedInUser.Name, loggedInUser.Email, loggedInUser.OIDCID)

	// Sauvegarder l'utilisateur dans la base de données
	result := database.DB.Save(&loggedInUser)
	if result.Error != nil {
		log.Printf("ERREUR: Erreur lors de la mise à jour de l'utilisateur dans la DB: %v", result.Error)
		session.AddFlash("Erreur lors de la mise à jour de votre profil.", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	log.Printf("DEBUG: Utilisateur sauvegardé en DB. Lignes affectées: %d", result.RowsAffected)

	// Mettre à jour l'utilisateur en session
	session.Set("user", loggedInUser)
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session après mise à jour: %v", err)
	}

	log.Printf("DEBUG: Session mise à jour avec l'utilisateur: %+v", loggedInUser)

	session.AddFlash("Votre profil a été mis à jour avec succès !", "success")
	session.Save()
	c.Redirect(http.StatusFound, "/profile")
}
