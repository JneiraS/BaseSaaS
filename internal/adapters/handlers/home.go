package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/components/elements"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Page d'accueil
func HomeHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user := session.Get("user")
	csrfToken := c.MustGet("csrf_token").(string)
	conn_button := elements.Button("Connexion", "btn btn-primary", "/login")
	logout_button := elements.Button("Déconnexion", "btn btn-primary", "/logout")
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":            "Accueil",
		"user":             user,
		"connexion_button": conn_button,
		"logout_button":    logout_button,
		"navbar":           navbar,
	})
}
