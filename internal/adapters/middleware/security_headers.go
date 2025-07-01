package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders ajoute des en-têtes de sécurité importants à chaque réponse.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Politique de sécurité du contenu (CSP) - Restreint les sources de contenu.
		// 'self' permet le contenu du même domaine.
		// Vous devrez peut-être ajouter d'autres sources pour les CDN (ex: fonts.googleapis.com).
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; object-src 'none';")

		// Empêche le navigateur de deviner le type MIME.
		c.Header("X-Content-Type-Options", "nosniff")

		// Empêche le clickjacking en interdisant l'intégration dans des iframes.
		c.Header("X-Frame-Options", "DENY")

		// Active la protection XSS intégrée au navigateur.
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}
