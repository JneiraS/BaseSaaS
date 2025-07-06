package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/components/elements"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// HomeHandler displays the application's home page.
// It retrieves user session information and renders the "index.tmpl" template,
// passing dynamic content like navigation bar and authentication buttons.
func HomeHandler(c *gin.Context) {
	// Retrieve the session from the Gin context.
	session := c.MustGet("session").(sessions.Session)
	// Get the user object from the session. This might be nil if the user is not logged in.
	user := session.Get("user")
	// Retrieve the CSRF token from the Gin context, used for form submissions.
	csrfToken := c.MustGet("csrf_token").(string)

	// Create dynamic buttons based on authentication status.
	conn_button := elements.Button("Connexion", "btn btn-primary", "/login")
	logout_button := elements.Button("Déconnexion", "btn btn-primary", "/logout")

	// Generate the navigation bar component, which might vary based on user login status.
	navbar := components.NavBar(user, csrfToken, session)

	// Render the home page template with dynamic data.
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":            "Accueil",
		"user":             user,
		"connexion_button": conn_button,
		"logout_button":    logout_button,
		"navbar":           navbar,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans HomeHandler: %v", err)
	}
}
