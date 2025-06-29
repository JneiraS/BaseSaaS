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
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type App struct {
	authService  *services.AuthService
	authHandlers *h.AuthHandlers
}

func LoadEnv() error {
	return godotenv.Load()
}

func init() {
	gob.Register(models.User{})
}

func (app *App) initOIDCProvider() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	providerURL := os.Getenv("OIDC_PROVIDER_URL")
	if providerURL == "" {
		providerURL = "http://localhost:8080"
	}

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return fmt.Errorf("impossible de se connecter au provider OIDC: %w", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("CLIENT_ID ou CLIENT_SECRET manquant")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	app.authService = services.NewAuthService(provider, oauth2Config)
	app.authHandlers = h.NewAuthHandlers(app.authService)

	return nil
}

func (app *App) setupServer() *gin.Engine {
	r := gin.Default()

	secretKey := []byte(os.Getenv("SESSION_SECRET"))
	if len(secretKey) == 0 {
		secretKey = []byte("ma-cle-secrete-de-32-caracteres-minimum-pour-securite")
	}

	store := cookie.NewStore(secretKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	})

	r.Use(sessions.Sessions("mysession", store))

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

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	return r
}

func (app *App) setupRoutes(r *gin.Engine) {
	r.GET("/", h.HomeHandler)
	r.GET("/profile", app.authRequired(), h.ProfileHandler)

	if app.authService != nil {
		r.GET("/login", app.authHandlers.LoginHandler)
		r.GET("/logout", app.authHandlers.LogoutHandler)
		r.GET("/callback", app.authHandlers.CallbackHandler)
	}

	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"server":       "running",
			"auth_service": app.authService != nil,
		}
		c.JSON(http.StatusOK, status)
	})
}

func (app *App) authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	app := &App{}

	if err := LoadEnv(); err != nil {
		log.Printf("Avertissement: Impossible de charger .env: %v", err)
	}

	if err := app.initOIDCProvider(); err != nil {
		log.Printf("AVERTISSEMENT: Authentification indisponible: %v", err)
	}

	r := app.setupServer()
	app.setupRoutes(r)

	log.Println("🚀 Serveur démarré sur :3000")
	r.Run(":3000")
}
