package middleware

import (
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
			c.String(http.StatusBadRequest, "CSRF token mismatch")
			c.Abort()
		},
	})
}
