package handlers

import (
	"context"
	"encoding/json"
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

// App encapsulates all application dependencies and configurations.
// It holds references to various services, handlers, the database connection,
// the Gin router, and application-wide configuration.
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

// NewApp creates and initializes a new instance of the application.
// It loads configuration, initializes the database, sets up all services and handlers,
// and configures the Gin router with middleware and routes.
func NewApp() (*App, error) {
	app := &App{}

	// Load application configuration from environment variables or .env file.
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	app.cfg = cfg

	// Initialize the database connection.
	db, err := database.InitDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.db = db
	log.Println("Database connection established.")

	// Initialize repositories (data access layer).
	app.userRepo = repositories.NewGormUserRepository(app.db)
	memberRepo := repositories.NewGormMemberRepository(app.db)
	eventRepo := repositories.NewGormEventRepository(app.db)
	transactionRepo := repositories.NewGormTransactionRepository(app.db)
	documentRepo := repositories.NewGormDocumentRepository(app.db)
	pollRepo := repositories.NewGormPollRepository(app.db)
	voteRepo := repositories.NewGormVoteRepository(app.db)

	// Initialize services (business logic layer).
	app.profileService = services.NewProfileService(app.userRepo)
	app.memberService = services.NewMemberService(memberRepo)
	app.eventService = services.NewEventService(eventRepo)
	app.emailService = services.NewEmailService(app.cfg)
	app.financeService = services.NewFinanceService(transactionRepo)
	app.documentService = services.NewDocumentService(documentRepo, app.cfg)
	app.pollService = services.NewPollService(pollRepo, voteRepo)

	// Initialize OIDC provider for authentication. This is optional;
	// the server can start without it if OIDC configuration is missing.
	if err := app.initOIDCProvider(); err != nil {
		log.Printf("WARNING: Authentication unavailable: %v", err)
	}

	// Auto-migrate database schemas for all models.
	if err := app.db.AutoMigrate(&repositories.UserDB{}, &repositories.MemberDB{}, &repositories.EventDB{}, &repositories.TransactionDB{}, &repositories.DocumentDB{}, &repositories.PollDB{}, &repositories.OptionDB{}, &repositories.VoteDB{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed.")

	// Initialize handlers (API/UI layer), injecting their respective services.
	app.authHandlers = NewAuthHandlers(app.authService, app.cfg)
	app.memberHandlers = NewMemberHandlers(app.memberService)
	app.eventHandlers = NewEventHandlers(app.eventService)
	app.communicationHandlers = NewCommunicationHandlers(app.emailService, app.memberService)
	app.financeHandlers = NewFinanceHandlers(app.financeService)
	app.documentHandlers = NewDocumentHandlers(app.documentService)
	app.statisticsHandlers = NewStatisticsHandlers(app.memberService, app.financeService, app.eventService, app.documentService)
	app.pollHandlers = NewPollHandlers(app.pollService)

	// Set up the Gin server and define all application routes.
	router := app.setupServer()
	app.setupRoutes(router)
	app.router = router

	return app, nil
}

// Run starts the application's HTTP server.
func (app *App) Run() {
	log.Println("🚀 Server started on :3000")
	if err := app.router.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// initOIDCProvider configures the OpenID Connect (OIDC) provider.
// It sets up the OIDC client and the OAuth2 configuration for authentication flows.
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

// setupServer configures and returns a Gin engine instance with global middleware.
// This includes security headers, session management, CSRF protection, and context injection.
func (app *App) setupServer() *gin.Engine {
	r := gin.Default()

	// Apply security headers to all responses.
	r.Use(middleware.SecurityHeaders(app.cfg))

	// Configure and apply cookie-based session management.
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

	// Apply CSRF protection middleware.
	r.Use(middleware.CSRFProtection(app.cfg))

	// Inject common context variables into Gin's context.
	r.Use(middleware.ContextInjector())

	// Set up custom template functions for HTML rendering.
	// These functions provide utilities like safe HTML rendering, arithmetic operations,
	// percentage calculations, and JSON marshaling for templates.
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

	// Load HTML templates from the "templates" directory.
	r.LoadHTMLGlob("templates/*")
	// Serve static files from the "static" directory.
	r.Static("/static", "./static")
	return r
}

// setupRoutes defines all application routes and assigns them to their respective handlers.
// Routes are grouped by functionality (e.g., members, events, finance) and protected
// by the authRequired middleware where necessary.
func (app *App) setupRoutes(r *gin.Engine) {
	// Public routes
	r.GET("/", app.LandingPage)
	r.GET("/home", HomeHandler) // Connected home page

	// Profile management routes (authentication required)
	r.GET("/profile", app.authRequired(), app.ProfileHandler)
	r.POST("/profile/update", app.authRequired(), app.UpdateProfileHandler)

	// Member management routes (authentication required)
	r.GET("/members", app.authRequired(), app.memberHandlers.ListMembers)
	r.GET("/members/new", app.authRequired(), app.memberHandlers.ShowCreateMemberForm)
	r.POST("/members/new", app.authRequired(), app.memberHandlers.CreateMember)
	r.GET("/members/edit/:id", app.authRequired(), app.memberHandlers.ShowEditMemberForm)
	r.POST("/members/edit/:id", app.authRequired(), app.memberHandlers.UpdateMember)
	r.POST("/members/delete/:id", app.authRequired(), app.memberHandlers.DeleteMember)
	r.POST("/members/mark-payment/:id", app.authRequired(), app.memberHandlers.MarkPayment)

	// Event management routes (authentication required)
	r.GET("/events", app.authRequired(), app.eventHandlers.ListEvents)
	r.GET("/events/new", app.authRequired(), app.eventHandlers.ShowCreateEventForm)
	r.POST("/events/new", app.authRequired(), app.eventHandlers.CreateEvent)
	r.GET("/events/edit/:id", app.authRequired(), app.eventHandlers.ShowEditEventForm)
	r.POST("/events/edit/:id", app.authRequired(), app.eventHandlers.UpdateEvent)
	r.POST("/events/delete/:id", app.authRequired(), app.eventHandlers.DeleteEvent)

	// Communication routes (authentication required)
	r.GET("/communication/email", app.authRequired(), app.communicationHandlers.ShowEmailForm)
	r.POST("/communication/email", app.authRequired(), app.communicationHandlers.SendEmailToMembers)

	// Financial management routes (authentication required)
	r.GET("/finance/transactions", app.authRequired(), app.financeHandlers.ListTransactions)
	r.GET("/finance/transactions/new", app.authRequired(), app.financeHandlers.ShowCreateTransactionForm)
	r.POST("/finance/transactions/new", app.authRequired(), app.financeHandlers.CreateTransaction)
	r.GET("/finance/transactions/edit/:id", app.authRequired(), app.financeHandlers.ShowEditTransactionForm)
	r.POST("/finance/transactions/edit/:id", app.authRequired(), app.financeHandlers.UpdateTransaction)
	r.POST("/finance/transactions/delete/:id", app.authRequired(), app.financeHandlers.DeleteTransaction)

	// Document management routes (authentication required)
	r.GET("/documents", app.authRequired(), app.documentHandlers.ListDocuments)
	r.GET("/documents/upload", app.authRequired(), app.documentHandlers.ShowUploadForm)
	r.POST("/documents/upload", app.authRequired(), app.documentHandlers.UploadDocument)
	r.GET("/documents/download/:id", app.authRequired(), app.documentHandlers.DownloadDocument)
	r.POST("/documents/delete/:id", app.authRequired(), app.documentHandlers.DeleteDocument)

	// Poll management routes (authentication required)
	r.GET("/polls", app.authRequired(), app.pollHandlers.ListPolls)
	r.GET("/polls/new", app.authRequired(), app.pollHandlers.ShowCreatePollForm)
	r.POST("/polls/new", app.authRequired(), app.pollHandlers.CreatePoll)
	r.GET("/polls/:id", app.authRequired(), app.pollHandlers.ShowPollDetails)
	r.POST("/polls/:id/vote", app.authRequired(), app.pollHandlers.VoteOnPoll)
	r.POST("/polls/delete/:id", app.authRequired(), app.pollHandlers.DeletePoll)

	// Statistics API routes (authentication required)
	r.GET("/api/stats/members", app.authRequired(), app.statisticsHandlers.GetMemberStats)
	r.GET("/api/stats/finance", app.authRequired(), app.statisticsHandlers.GetFinanceStats)
	r.GET("/api/stats/documents", app.authRequired(), app.statisticsHandlers.GetDocumentStats)
	r.GET("/api/stats/events", app.authRequired(), app.statisticsHandlers.GetEventStats)

	// Dashboard route (authentication required)
	r.GET("/dashboard", app.authRequired(), app.statisticsHandlers.ShowDashboard)

	// Authentication routes (active only if OIDC service is configured)
	if app.authService != nil {
		r.GET("/login", middleware.AuthConfigured(app.authService), app.authHandlers.LoginHandler)
		r.POST("/logout", middleware.AuthConfigured(app.authService), app.authHandlers.LogoutHandler)
		r.GET("//callback", middleware.AuthConfigured(app.authService), app.authHandlers.CallbackHandler)
	}

	// Health check endpoint for monitoring.
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"server":       "running",
			"auth_service": app.authService != nil,
		}
		c.JSON(http.StatusOK, status)
	})

	// Middleware to inject CSRF token into the template context for all requests.
	r.Use(func(c *gin.Context) {
		c.Set("csrf_token", csrf.GetToken(c))
		c.Next()
	})
}

// authRequired is a middleware that checks if a user is authenticated.
// If no user is found in the session, it redirects to the login page.
func (app *App) authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log all incoming cookies for debugging
		for _, cookie := range c.Request.Cookies() {
			log.Printf("AuthRequired: Received cookie - Name: %s, Value: %s", cookie.Name, cookie.Value)
		}

		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			log.Printf("AuthRequired: User not found in session. Redirecting to /login. Request path: %s", c.Request.URL.Path)
			c.Redirect(http.StatusFound, "/login")
			c.Abort() // Abort the request chain
			return
		}
		log.Printf("AuthRequired: User %v found in session. Request path: %s", user, c.Request.URL.Path)
		c.Next() // Proceed to the next handler in the chain
	}
}
