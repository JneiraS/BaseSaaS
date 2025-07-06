package middleware

import (
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// CSRFProtection applique une protection contre les attaques CSRF.
func CSRFProtection(cfg *config.Config) gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: cfg.CSRFSecret,
		ErrorFunc: func(c *gin.Context) {
			log.Printf("ERREUR CSRF: Token mismatch. Expected: %s, Received: %s", csrf.GetToken(c), c.Request.FormValue("_csrf"))
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort()
		},
	})
}
