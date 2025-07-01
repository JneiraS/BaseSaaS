package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// CSRFProtection applique une protection contre les attaques CSRF.
func CSRFProtection() gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: "a-very-strong-and-secret-key-for-csrf", // Doit être changé et chargé depuis les envs
		ErrorFunc: func(c *gin.Context) {
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort()
		},
	})
}
