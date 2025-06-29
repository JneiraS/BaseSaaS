package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Page profil (protégée)
func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	navbar := components.NavBar()

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":  "Profil",
		"user":   user,
		"navbar": navbar,
	})
}
