package middleware

import (
	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders is a middleware that adds important security headers to every HTTP response.
// These headers help protect against common web vulnerabilities like XSS, clickjacking, and MIME-type sniffing.
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy (CSP) - Restricts content sources.
		// 'self' allows content from the same origin.
		// Additional sources for CDNs (e.g., fonts.googleapis.com) might need to be added based on application needs.
		c.Header("Content-Security-Policy", cfg.ContentSecurityPolicy)

		// X-Content-Type-Options - Prevents browsers from MIME-sniffing a response away from the declared content-type.
		// This helps prevent XSS attacks where a browser might interpret a non-script file as a script.
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options - Prevents clickjacking by disallowing embedding the page in iframes.
		// "DENY" prevents any domain from framing the content.
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection - Enables the browser's built-in XSS filter.
		// "1; mode=block" enables the filter and blocks the page if an XSS attack is detected.
		c.Header("X-XSS-Protection", "1; mode=block")

		// Proceed to the next handler in the chain.
		c.Next()
	}
}
