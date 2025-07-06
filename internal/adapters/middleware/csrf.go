package middleware

import (
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// CSRFProtection applies Cross-Site Request Forgery (CSRF) protection to the application.
// It uses the gin-csrf middleware to validate CSRF tokens on incoming requests.
func CSRFProtection(cfg *config.Config) gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: cfg.CSRFSecret, // The secret key used to sign and verify CSRF tokens.
		// ErrorFunc is a custom function to handle CSRF validation failures.
		// It logs the error and returns a 400 Bad Request response to the client.
		ErrorFunc: func(c *gin.Context) {
			log.Printf("ERREUR CSRF: Token mismatch. Expected: %s, Received: %s", csrf.GetToken(c), c.Request.FormValue("_csrf"))
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort() // Abort the request chain on CSRF validation failure.
		},
	})
}
