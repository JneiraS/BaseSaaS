package middleware

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// ContextInjector is a middleware that injects common application-wide data
// such as the session and CSRF token into the Gin context.
// This makes these objects easily accessible by subsequent handlers in the request chain.
func ContextInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the current session from the Gin context.
		session := sessions.Default(c)
		// Get the CSRF token for the current request.
		csrfToken := csrf.GetToken(c)
		log.Printf("DEBUG: CSRF Token injected: %s", csrfToken)

		// Set the session and CSRF token in the Gin context.
		// These can then be retrieved by handlers using c.MustGet("session") or c.MustGet("csrf_token").
		c.Set("session", session)
		c.Set("csrf_token", csrfToken)

		// Proceed to the next handler in the chain.
		c.Next()
	}
}
