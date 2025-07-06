package middleware

import (
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthConfigured is a middleware that checks if the authentication service is properly configured.
// It ensures that OIDC provider details are available before allowing access to authentication-dependent routes.
func AuthConfigured(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the AuthService is configured. If not, return a service unavailable error.
		if !authService.IsConfigured() {
			log.Printf("ERREUR: AuthService n'est pas configuré")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service d'authentification temporairement indisponible",
				"message": "Le provider OIDC n'est pas accessible. Vérifiez que Zitadel fonctionne.",
			})
			c.Abort() // Abort the request chain as authentication is not possible.
			return
		}
		// If configured, proceed to the next handler in the chain.
		c.Next()
	}
}
