package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/JneiraS/BaseSasS/internal/adapters/middleware"
	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/JneiraS/BaseSasS/internal/database"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
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
	authService    *services.AuthService
	authHandlers   *AuthHandlers
	profileService *services.ProfileService
	db             *gorm.DB
	router         *gin.Engine
	cfg            *config.Config
	userRepo       repositories.UserRepository
}

// NewApp crée et initialise une nouvelle instance de l'application.
func NewApp() (*App, error) {
	app := &App{}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	app.cfg = cfg

	db, err := database.InitDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.db = db
	log.Println("Database connection established.")

	app.userRepo = repositories.NewGormUserRepository(app.db)

	app.profileService = services.NewProfileService(app.userRepo)

	// L'authentification est optionnelle, le serveur peut démarrer sans.
	if err := app.initOIDCProvider(); err != nil {
		log.Printf("AVERTISSEMENT: Authentification indisponible: %v", err)
	}

	if err := app.db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed.")

	app.authHandlers = NewAuthHandlers(app.authService, app.cfg)

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

	provider, err := oidc.NewProvider(ctx, app.cfg.OIDCProviderURL)
	if err != nil {
		return fmt.Errorf("impossible de se connecter au provider OIDC: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     app.cfg.ClientID,
		ClientSecret: app.cfg.ClientSecret,
		RedirectURL:  app.cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	app.authService = services.NewAuthService(provider, oauth2Config, app.userRepo)

	return nil
}

// setupServer configure et retourne une instance du serveur Gin.
func (app *App) setupServer() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.SecurityHeaders(app.cfg))

	secretKey := []byte(app.cfg.SessionSecret)
	store := cookie.NewStore(secretKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   app.cfg.SessionMaxAge,
		HttpOnly: app.cfg.SessionHttpOnly,
		Secure:   app.cfg.SessionSecure,
		SameSite: app.cfg.SessionSameSiteMode(),
	})

	r.Use(sessions.Sessions(app.cfg.CookieName, store))
		r.Use(middleware.CSRFProtection(app.cfg))
	r.Use(middleware.ContextInjector())

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
	r.GET("/profile", app.authRequired(), app.ProfileHandler)
	r.POST("/profile/update", app.authRequired(), app.UpdateProfileHandler)

	// Les routes d'authentification ne sont actives que si le service OIDC est configuré.
	if app.authService != nil {
		r.GET("/login", middleware.AuthConfigured(app.authService), app.authHandlers.LoginHandler)
		r.POST("/logout", middleware.AuthConfigured(app.authService), app.authHandlers.LogoutHandler)
		r.GET("//callback", middleware.AuthConfigured(app.authService), app.authHandlers.CallbackHandler)
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
