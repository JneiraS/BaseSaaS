package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// ContextInjector injecte la session et le jeton CSRF dans le contexte Gin.
func ContextInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		csrfToken := csrf.GetToken(c)

		c.Set("session", session)
		c.Set("csrf_token", csrfToken)

		c.Next()
	}
}
