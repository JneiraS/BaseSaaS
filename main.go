package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	h "github.com/JneiraS/BaseSasS/internal/adapters/handlers"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func LoadEnv() error {
	return godotenv.Load()
}

var (
	provider *oidc.Provider
)

func init() {
	// Enregistrer le type User pour gob
	gob.Register(models.User{})

	// Chargement des variables d'environnement
	err := LoadEnv()
	if err != nil {
		log.Printf("Avertissement: Impossible de charger .env: %v", err)
	}

	// Initialiser le provider OIDC de manière sécurisée
	initOIDCProvider()
}

func initOIDCProvider() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenir l'URL du provider
	providerURL := os.Getenv("OIDC_PROVIDER_URL")
	if providerURL == "" {
		providerURL = "http://localhost:8080"
	}

	log.Printf("Tentative de connexion au provider OIDC: %s", providerURL)

	var err error
	provider, err = oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Printf("AVERTISSEMENT: Impossible de se connecter au provider OIDC (%s): %v", providerURL, err)
		log.Printf("L'authentification ne sera pas disponible. Assurez-vous que Zitadel fonctionne.")
		return
	}

	// Vérifier les variables d'environnement
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Printf("AVERTISSEMENT: CLIENT_ID ou CLIENT_SECRET manquant dans les variables d'environnement")
		return
	}

	h.Oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// IMPORTANT: Passer le provider aux handlers
	h.Provider = provider

	log.Println("✅ Provider OIDC initialisé avec succès")
}

func SetupServer() *gin.Engine {
	r := gin.Default()

	// Configuration des sessions
	secretKey := []byte("ma-cle-secrete-de-32-caracteres-minimum-pour-securite")
	store := cookie.NewStore(secretKey)

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	})

	r.Use(sessions.Sessions("mysession", store))

	// Middleware de logging
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		log.Printf("Session avant requête - Path: %s, Session ID: %v", c.Request.URL.Path, session.Get("id"))
		c.Next()
	})

	// Templates
	r.LoadHTMLGlob("templates/*")

	r.SetFuncMap(template.FuncMap{
		"safe": func(s any) template.HTML {
			switch v := s.(type) {
			case string:
				return template.HTML(v)
			case fmt.Stringer:
				return template.HTML(v.String())
			default:
				return template.HTML(fmt.Sprint(v))
			}
		},
	})

	return r
}

func main() {
	r := SetupServer()

	// Routes
	r.GET("/", h.HomeHandler)
	r.GET("/login", h.LoginHandler)
	r.GET("/callback", h.CallbackHandler)
	r.GET("/profile", authRequired(), h.ProfileHandler)
	r.GET("/logout", h.LogoutHandler)

	// Route de diagnostic
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"server":        "running",
			"oidc_provider": provider != nil,
			"oauth2_config": h.Oauth2Config != nil,
		}
		c.JSON(http.StatusOK, status)
	})

	log.Println("🚀 Serveur démarré sur :3000")
	log.Println("📋 Visitez http://localhost:3000/health pour vérifier l'état des services")
	r.Run(":3000")
}

// Middleware d'authentification
func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		log.Printf("Middleware auth - User: %v", user)
		log.Printf("Middleware auth - Path: %s", c.Request.URL.Path)

		if user == nil {
			if c.Request.URL.Path == "/login" {
				c.Next()
				return
			}
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
