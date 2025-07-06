package handlers

import (
	"context"
	"encoding/json" // Ajoutez cette ligne
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/JneiraS/BaseSasS/internal/adapters/middleware"
	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/JneiraS/BaseSasS/internal/database"
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
	authService           *services.AuthService
	authHandlers          *AuthHandlers
	profileService        *services.ProfileService
	memberService         *services.MemberService
	eventService          *services.EventService
	emailService          *services.EmailService
	financeService        *services.FinanceService
	documentService       *services.DocumentService
	pollService           *services.PollService
	memberHandlers        *MemberHandlers
	eventHandlers         *EventHandlers
	communicationHandlers *CommunicationHandlers
	financeHandlers       *FinanceHandlers
	documentHandlers      *DocumentHandlers
	statisticsHandlers    *StatisticsHandlers
	pollHandlers          *PollHandlers // Ajout des handlers de sondages
	db                    *gorm.DB
	router                *gin.Engine
	cfg                   *config.Config
	userRepo              repositories.UserRepository
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
	memberRepo := repositories.NewGormMemberRepository(app.db) // Créez le repo de membres

	app.profileService = services.NewProfileService(app.userRepo)
	app.memberService = services.NewMemberService(memberRepo)                // Initialisez le service de membres
	eventRepo := repositories.NewGormEventRepository(app.db)                 // Créez le repo d'événements
	app.eventService = services.NewEventService(eventRepo)                   // Initialisez le service d'événements
	app.emailService = services.NewEmailService(app.cfg)                     // Initialisez le service d'e-mail
	transactionRepo := repositories.NewGormTransactionRepository(app.db)     // Créez le repo de transactions
	app.financeService = services.NewFinanceService(transactionRepo)         // Initialisez le service financier
	documentRepo := repositories.NewGormDocumentRepository(app.db)           // Créez le repo de documents
	app.documentService = services.NewDocumentService(documentRepo, app.cfg) // Initialisez le service de documents
	pollRepo := repositories.NewGormPollRepository(app.db)                   // Créez le repo de sondages
	voteRepo := repositories.NewGormVoteRepository(app.db)                   // Créez le repo de votes
	app.pollService = services.NewPollService(pollRepo, voteRepo)            // Initialisez le service de sondages

	// L'authentification est optionnelle, le serveur peut démarrer sans.
	if err := app.initOIDCProvider(); err != nil {
		log.Printf("AVERTISSEMENT: Authentification indisponible: %v", err)
	}

	if err := app.db.AutoMigrate(&repositories.UserDB{}, &repositories.MemberDB{}, &repositories.EventDB{}, &repositories.TransactionDB{}, &repositories.DocumentDB{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed.")

	app.authHandlers = NewAuthHandlers(app.authService, app.cfg)
	app.memberHandlers = NewMemberHandlers(app.memberService)                                                                    // Initialisez les handlers de membres
	app.eventHandlers = NewEventHandlers(app.eventService)                                                                       // Initialisez les handlers d'événements
	app.communicationHandlers = NewCommunicationHandlers(app.emailService, app.memberService)                                    // Initialisez les handlers de communication
	app.financeHandlers = NewFinanceHandlers(app.financeService)                                                                 // Initialisez les handlers financiers
	app.documentHandlers = NewDocumentHandlers(app.documentService)                                                              // Initialisez les handlers de documents
	app.statisticsHandlers = NewStatisticsHandlers(app.memberService, app.financeService, app.eventService, app.documentService) // Initialisez les handlers de statistiques
	app.pollHandlers = NewPollHandlers(app.pollService)                                                                          // Initialisez les handlers de sondages

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
		"add": func(a, b int64) int64 { return a + b },
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b float64) float64 { return a / b },
		"float": func(a interface{}) float64 {
			switch v := a.(type) {
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case float64:
				return v
			default:
				return 0.0
			}
		},
		"percentage": func(part, total int64) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
		"formatPercent": func(part, total int64) string {
			if total == 0 {
				return "0%"
			}
			percent := float64(part) / float64(total) * 100
			return fmt.Sprintf("%.1f%%", percent)
		},
		"json": func(v interface{}) (template.JS, error) {
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(jsonBytes), nil
		},
		"string": func(i interface{}) string {
			return fmt.Sprintf("%v", i)
		},
		"int": func(f float64) int64 {
			return int64(f)
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

	// Routes pour la gestion des membres
	r.GET("/members", app.authRequired(), app.memberHandlers.ListMembers)
	r.GET("/members/new", app.authRequired(), app.memberHandlers.ShowCreateMemberForm)
	r.POST("/members/new", app.authRequired(), app.memberHandlers.CreateMember)
	r.GET("/members/edit/:id", app.authRequired(), app.memberHandlers.ShowEditMemberForm)
	r.POST("/members/edit/:id", app.authRequired(), app.memberHandlers.UpdateMember)
	r.POST("/members/delete/:id", app.authRequired(), app.memberHandlers.DeleteMember)
	r.POST("/members/mark-payment/:id", app.authRequired(), app.memberHandlers.MarkPayment)

	// Routes pour la gestion des événements
	r.GET("/events", app.authRequired(), app.eventHandlers.ListEvents)
	r.GET("/events/new", app.authRequired(), app.eventHandlers.ShowCreateEventForm)
	r.POST("/events/new", app.authRequired(), app.eventHandlers.CreateEvent)
	r.GET("/events/edit/:id", app.authRequired(), app.eventHandlers.ShowEditEventForm)
	r.POST("/events/edit/:id", app.authRequired(), app.eventHandlers.UpdateEvent)
	r.POST("/events/delete/:id", app.authRequired(), app.eventHandlers.DeleteEvent)

	// Routes pour la communication
	r.GET("/communication/email", app.authRequired(), app.communicationHandlers.ShowEmailForm)
	r.POST("/communication/email", app.authRequired(), app.communicationHandlers.SendEmailToMembers)

	// Routes pour la gestion financière
	r.GET("/finance/transactions", app.authRequired(), app.financeHandlers.ListTransactions)
	r.GET("/finance/transactions/new", app.authRequired(), app.financeHandlers.ShowCreateTransactionForm)
	r.POST("/finance/transactions/new", app.authRequired(), app.financeHandlers.CreateTransaction)
	r.GET("/finance/transactions/edit/:id", app.authRequired(), app.financeHandlers.ShowEditTransactionForm)
	r.POST("/finance/transactions/edit/:id", app.authRequired(), app.financeHandlers.UpdateTransaction)
	r.POST("/finance/transactions/delete/:id", app.authRequired(), app.financeHandlers.DeleteTransaction)

	// Routes pour la gestion des documents
	r.GET("/documents", app.authRequired(), app.documentHandlers.ListDocuments)
	r.GET("/documents/upload", app.authRequired(), app.documentHandlers.ShowUploadForm)
	r.POST("/documents/upload", app.authRequired(), app.documentHandlers.UploadDocument)
	r.GET("/documents/download/:id", app.authRequired(), app.documentHandlers.DownloadDocument)
	r.POST("/documents/delete/:id", app.authRequired(), app.documentHandlers.DeleteDocument)

	// Routes pour les sondages
	r.GET("/polls", app.authRequired(), app.pollHandlers.ListPolls)
	r.GET("/polls/new", app.authRequired(), app.pollHandlers.ShowCreatePollForm)
	r.POST("/polls/new", app.authRequired(), app.pollHandlers.CreatePoll)
	r.GET("/polls/:id", app.authRequired(), app.pollHandlers.ShowPollDetails)
	r.POST("/polls/:id/vote", app.authRequired(), app.pollHandlers.VoteOnPoll)
	r.POST("/polls/delete/:id", app.authRequired(), app.pollHandlers.DeletePoll)

	// Routes pour les statistiques (API)
	r.GET("/api/stats/members", app.authRequired(), app.statisticsHandlers.GetMemberStats)
	r.GET("/api/stats/finance", app.authRequired(), app.statisticsHandlers.GetFinanceStats)
	r.GET("/api/stats/events", app.authRequired(), app.statisticsHandlers.GetEventStats)
	r.GET("/api/stats/documents", app.authRequired(), app.statisticsHandlers.GetDocumentStats)

	// Route pour le tableau de bord
	r.GET("/dashboard", app.authRequired(), app.statisticsHandlers.ShowDashboard)

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
