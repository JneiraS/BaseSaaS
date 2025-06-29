// main.go
package main

import (
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/JneiraS/BaseSasS/domain/models"
	h "github.com/JneiraS/BaseSasS/internal/adapters/handlers"
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
		log.Fatal("Erreur lors du chargement des variables d'environnement:", err)
	}

	// Configuration Zitadel
	ctx := context.Background()

	// Remplacez par l'URL de votre instance Zitadel
	provider, err = oidc.NewProvider(ctx, "http://localhost:8080")
	if err != nil {
		log.Fatal("Erreur provider OIDC:", err)
	}

	h.Oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}
func main() {
	r := gin.Default()

	// Configuration des sessions avec une clé plus longue et sécurisée
	secretKey := []byte("ma-cle-secrete-de-32-caracteres-minimum-pour-securite")
	store := cookie.NewStore(secretKey)

	// Options de session plus permissives pour le développement
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,                    // 1 jour
		HttpOnly: false,                    // Permettre l'accès JavaScript pour debug
		Secure:   false,                    // HTTP en développement
		SameSite: http.SameSiteDefaultMode, // Plus permissif
	})

	r.Use(sessions.Sessions("mysession", store))

	// Middleware de logging des sessions
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		log.Printf("Session avant requête - Path: %s, Session ID: %v", c.Request.URL.Path, session.Get("id"))
		c.Next()
	})

	// Templates
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", h.HomeHandler)
	r.GET("/login", h.LoginHandler)
	r.GET("/callback", h.CallbackHandler)
	r.GET("/profile", h.ProfileHandler)
	r.GET("/logout", h.LogoutHandler)

	log.Println("Serveur démarré sur :3000")
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
			// Éviter la redirection infinie
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
