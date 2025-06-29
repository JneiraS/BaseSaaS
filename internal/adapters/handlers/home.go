package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components/elements"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Page d'accueil
func HomeHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	profile_button := elements.Button("Mon profil", "btn btn-primary", "/profile")
	conn_button := elements.Button("Connexion", "btn btn-primary", "/login")
	logout_button := elements.Button("Déconnexion", "btn btn-primary", "/logout")

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":            "Accueil",
		"user":             user,
		"connexion_button": conn_button,
		"profile_button":   profile_button,
		"logout_button":    logout_button,
	})
}
