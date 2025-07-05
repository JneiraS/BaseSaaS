package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JneiraS/BaseSasS/internal/adapters/middleware"
	"github.com/JneiraS/BaseSasS/internal/database"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// App encapsule les dépendances de l'application.
type App struct {
	authService  *services.AuthService
	authHandlers *AuthHandlers
	db           *gorm.DB
	router       *gin.Engine
}

// NewApp crée et initialise une nouvelle instance de l'application.
func NewApp() (*App, error) {
	app := &App{}

	// L'authentification est optionnelle, le serveur peut démarrer sans.
	if err := app.initOIDCProvider(); err != nil {
		log.Printf("AVERTISSEMENT: Authentification indisponible: %v", err)
	}

	database.InitDatabase()
	app.db = database.DB

	if err := app.db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed.")

	app.authHandlers = NewAuthHandlers(app.authService)

	router := app.setupServer()
	app.setupRoutes(router)
	app.router = router

	return app, nil
}

// Run démarre le serveur de l'application.
func (app *App) Run() {
	log.Println("🚀 Serveur démarré sur :3000")
	if err := app.router.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// initOIDCProvider configure le fournisseur OpenID Connect (OIDC).
func (app *App) initOIDCProvider() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	providerURL := os.Getenv("OIDC_PROVIDER_URL")
	if providerURL == "" {
		providerURL = "http://localhost:8080" // URL par défaut pour le développement
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

	return nil
}

// setupServer configure et retourne une instance du serveur Gin.
func (app *App) setupServer() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.SecurityHeaders())

	secretKey := []byte(os.Getenv("SESSION_SECRET"))
	if len(secretKey) == 0 {
		secretKey = []byte("ma-cle-secrete-de-32-caracteres-minimum-pour-securite")
		log.Println("Avertissement: SESSION_SECRET non définie, utilisation d'une clé par défaut.")
	}

	store := cookie.NewStore(secretKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 24 heures
		HttpOnly: false, // true en production
		Secure:   false, // true en production
		SameSite: http.SameSiteDefaultMode,
	})

	r.Use(sessions.Sessions("mysession", store))
	r.Use(middleware.CSRFProtection())

	// Permet d'utiliser `{{ safe .variable }}` dans les templates pour afficher du HTML.
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

// setupRoutes définit toutes les routes de l'application.
func (app *App) setupRoutes(r *gin.Engine) {
	r.GET("/", app.LandingPage)
	r.GET("/home", HomeHandler)
	r.GET("/profile", app.authRequired(), ProfileHandler)
	r.POST("/profile/update", app.authRequired(), UpdateProfileHandler)

	// Les routes d'authentification ne sont actives que si le service OIDC est configuré.
	if app.authService != nil {
		r.GET("/login", app.authHandlers.LoginHandler)
		r.POST("/logout", app.authHandlers.LogoutHandler)
		r.GET("//callback", app.authHandlers.CallbackHandler)
	}

	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"server":       "running",
			"auth_service": app.authService != nil,
		}
		c.JSON(http.StatusOK, status)
	})

	// Ajoute le jeton CSRF au contexte pour qu'il soit disponible dans les templates.
	r.Use(func(c *gin.Context) {
		c.Set("csrf_token", csrf.GetToken(c))
		c.Next()
	})
}

// authRequired est un middleware qui vérifie si un utilisateur est authentifié.
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
