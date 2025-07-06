package middleware

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// ContextInjector injecte la session et le jeton CSRF dans le contexte Gin.
func ContextInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		csrfToken := csrf.GetToken(c)
		log.Printf("DEBUG: CSRF Token injected: %s", csrfToken)

		c.Set("session", session)
		c.Set("csrf_token", csrfToken)

		c.Next()
	}
}
