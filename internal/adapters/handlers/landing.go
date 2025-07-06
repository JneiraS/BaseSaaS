package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func (app *App) LandingPage(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user := session.Get("user")
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "landing.tmpl", gin.H{
		"title":      "Welcome to BaseSasS",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		// Gérer l'erreur de sauvegarde de session si nécessaire
		// log.Printf("Erreur lors de la sauvegarde de session dans LandingPage: %v", err)
	}
}
