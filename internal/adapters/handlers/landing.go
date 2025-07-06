package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LandingPage displays the application's landing page.
// It retrieves user session information and renders the "landing.tmpl" template,
// passing dynamic content like the navigation bar.
func (app *App) LandingPage(c *gin.Context) {
	// Retrieve the session from the Gin context.
	session := c.MustGet("session").(sessions.Session)
	// Get the user object from the session. This might be nil if the user is not logged in.
	user := session.Get("user")
	// Retrieve the CSRF token from the Gin context, used for form submissions.
	csrfToken := c.MustGet("csrf_token").(string)

	// Generate the navigation bar component, which might vary based on user login status.
	navbar := components.NavBar(user, csrfToken, session)

	// Render the landing page template with dynamic data.
	c.HTML(http.StatusOK, "landing.tmpl", gin.H{
		"title":      "Welcome to BaseSasS",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans LandingPage: %v", err)
	}
}
