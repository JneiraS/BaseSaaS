package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Config holds all application-wide configuration settings.
// These settings are typically loaded from environment variables or a .env file.
type Config struct {
	// OIDC (OpenID Connect) Configuration
	OIDCProviderURL string // URL of the OIDC identity provider (e.g., Zitadel, Auth0)
	ClientID        string // Client ID for the OIDC application
	ClientSecret    string // Client Secret for the OIDC application
	RedirectURL     string // Callback URL after successful authentication

	// Session Management Configuration
	SessionSecret   string // Secret key used to encrypt and sign session cookies
	SessionMaxAge   int    // Maximum age of the session cookie in seconds
	SessionHttpOnly bool   // If true, the session cookie is only accessible via HTTP(S) requests (not JavaScript)
	SessionSecure   bool   // If true, the session cookie is only sent over HTTPS
	SessionSameSite string // SameSite policy for the session cookie (e.g., "Lax", "Strict", "None")
	AppURL          string // Base URL of the application
	CookieName      string // Name of the session cookie

	// CSRF (Cross-Site Request Forgery) Protection Configuration
	CSRFSecret string // Secret key for CSRF token generation and validation

	// Security Headers Configuration
	ContentSecurityPolicy string // Value for the Content-Security-Policy HTTP header

	// SMTP (Simple Mail Transfer Protocol) Configuration for sending emails
	SMTPHost     string // SMTP server host
	SMTPPort     int    // SMTP server port
	SMTPUsername string // Username for SMTP authentication
	SMTPPassword string // Password for SMTP authentication
	EmailSender  string // Email address used as the sender

	// Document Storage Configuration
	DocumentStoragePath string // File system path where uploaded documents are stored
}

// LoadConfig loads application configuration from environment variables.
// It sets default values for optional parameters and performs basic validation.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		OIDCProviderURL:       os.Getenv("OIDC_PROVIDER_URL"),
		ClientID:              os.Getenv("CLIENT_ID"),
		ClientSecret:          os.Getenv("CLIENT_SECRET"),
		RedirectURL:           getEnv("OIDC_REDIRECT_URL", "http://localhost:3000/callback"),
		SessionSecret:         os.Getenv("SESSION_SECRET"),
		SessionMaxAge:         getEnvAsInt("SESSION_MAX_AGE", 86400),    // 24 hours
		SessionHttpOnly:       getEnvAsBool("SESSION_HTTP_ONLY", false), // true in production
		SessionSecure:         false,    // false car nous sommes en HTTP sur localhost
		SessionSameSite:       "Lax",    // Lax est la valeur par défaut et fonctionne bien pour les requêtes same-site
		AppURL:                getEnv("APP_URL", "http://localhost:3000"),
		CookieName:            getEnv("COOKIE_NAME", "mysession"),
		CSRFSecret:            os.Getenv("CSRF_SECRET"),
		ContentSecurityPolicy: getEnv("CONTENT_SECURITY_POLICY", "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net 'sha256-nhU1dNZtRMH0wGMdWus+C2+OLS90BrB/ybY9vr8XxvA='; style-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; object-src 'none';"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587), // Default SMTP port
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		EmailSender:  getEnv("EMAIL_SENDER", "no-reply@assoss.com"),

		DocumentStoragePath: getEnv("DOCUMENT_STORAGE_PATH", "./data/documents"),
	}

	// Basic validation for essential OIDC configuration.
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("CLIENT_ID ou CLIENT_SECRET manquant")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value if not set.
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value if not set or invalid.
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsBool retrieves an environment variable as a boolean or returns a default value if not set or invalid.
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

// SessionSameSiteMode converts the string representation of SameSite policy to http.SameSite enum.
func (c *Config) SessionSameSiteMode() http.SameSite {
	switch c.SessionSameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode // Default to Lax if an unknown value is provided.
	}
}
