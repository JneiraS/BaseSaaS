package middleware

import (
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthConfigured est un middleware qui vérifie si le service d'authentification est configuré.
func AuthConfigured(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authService.IsConfigured() {
			log.Printf("ERREUR: AuthService n'est pas configuré")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service d'authentification temporairement indisponible",
				"message": "Le provider OIDC n'est pas accessible. Vérifiez que Zitadel fonctionne.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
