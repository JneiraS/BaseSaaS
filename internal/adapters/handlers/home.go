package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Page d'accueil
func HomeHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Accueil",
		"user":  user,
	})
}
