package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	OIDCProviderURL       string
	ClientID              string
	ClientSecret          string
	RedirectURL           string
	SessionSecret         string
	SessionMaxAge         int
	SessionHttpOnly       bool
	SessionSecure         bool
	SessionSameSite       string
	AppURL                string
	CookieName            string
	CSRFSecret            string
	ContentSecurityPolicy string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		OIDCProviderURL:       os.Getenv("OIDC_PROVIDER_URL"),
		ClientID:              os.Getenv("CLIENT_ID"),
		ClientSecret:          os.Getenv("CLIENT_SECRET"),
		RedirectURL:           getEnv("OIDC_REDIRECT_URL", "http://localhost:3000/callback"),
		SessionSecret:         os.Getenv("SESSION_SECRET"),
		SessionMaxAge:         getEnvAsInt("SESSION_MAX_AGE", 86400),    // 24 heures
		SessionHttpOnly:       getEnvAsBool("SESSION_HTTP_ONLY", false), // true en production
		SessionSecure:         getEnvAsBool("SESSION_SECURE", false),    // true en production
		SessionSameSite:       getEnv("SESSION_SAMESITE", "Lax"),
		AppURL:                getEnv("APP_URL", "http://localhost:3000"),
		CookieName:            getEnv("COOKIE_NAME", "mysession"),
		CSRFSecret:            os.Getenv("CSRF_SECRET"),
		ContentSecurityPolicy: getEnv("CONTENT_SECURITY_POLICY", "default-src 'self'; script-src 'self' 'sha256-nhU1dNZtRMH0wGMdWus+C2+OLS90BrB/ybY9vr8XxvA='; style-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com; font-src 'self' https://fonts.gstatic.com; object-src 'none';"),
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("CLIENT_ID ou CLIENT_SECRET manquant")
	}

	return cfg, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func (c *Config) SessionSameSiteMode() http.SameSite {
	switch c.SessionSameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
