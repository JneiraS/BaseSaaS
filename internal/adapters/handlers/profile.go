package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// Page profil (protégée)
func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	csrfToken := csrf.GetToken(c)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":  "Profil",
		"user":   user,
		"navbar": navbar,
	})
}
