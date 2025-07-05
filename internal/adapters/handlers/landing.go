package handlers

import (
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func (app *App) LandingPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	csrfToken := csrf.GetToken(c)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "landing.tmpl", gin.H{
		"title":   "Welcome to BaseSasS",
		"navbar":  navbar,
		"user":    user,
		"csrf_token": csrfToken,
	})
}
