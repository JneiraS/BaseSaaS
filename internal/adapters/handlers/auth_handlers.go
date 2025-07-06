package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// AuthHandlers encapsulates the dependencies for authentication-related HTTP handlers.
// It holds references to the AuthService and application configuration.
type AuthHandlers struct {
	authService *services.AuthService
	cfg         *config.Config
}

// NewAuthHandlers creates a new instance of AuthHandlers.
// It takes an AuthService and Config as dependencies.
func NewAuthHandlers(authService *services.AuthService, cfg *config.Config) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		cfg:         cfg,
	}
}

// LoginHandler initiates the OAuth2/OIDC login flow.
// It generates a state parameter, saves it in the session, and redirects the user to the OIDC provider's authorization URL.
func (h *AuthHandlers) LoginHandler(c *gin.Context) {

	session := c.MustGet("session").(sessions.Session)

	// If the user is already logged in, redirect them to the profile page.
	if user := session.Get("user"); user != nil {
		log.Printf("Utilisateur déjà connecté, redirection vers profile")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Generate a cryptographically secure random state to prevent CSRF attacks.
	state, err := generateRandomState()
	if err != nil {
		log.Printf("ERREUR génération de l'état: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur interne du serveur",
		})
		return
	}
	// Save the state in the session for later verification during the callback.
	session.Set("state", state)

	// Save the session to ensure the state is persisted.
	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur session",
		})
		return
	}

	// Construct the authorization URL and redirect the user to the OIDC provider.
	authURL := h.authService.Oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusFound, authURL)
}

// CallbackHandler processes the redirect from the OIDC provider after successful authentication.
// It verifies the state parameter, exchanges the authorization code for tokens, validates the ID token,
// and creates/finds the user in the database before setting the user in the session.
func (h *AuthHandlers) CallbackHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	code := c.Query("code")   // Authorization code from the OIDC provider
	state := c.Query("state") // State parameter from the OIDC provider

	// Retrieve the saved state from the session and compare it with the received state.
	savedState := session.Get("state")
	if savedState == nil || state != savedState.(string) {
		// If states do not match, it indicates a potential CSRF attack.
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State invalide",
		})
		return
	}

	// Remove the state from the session after verification to prevent replay attacks.
	session.Delete("state")

	// Exchange the authorization code for OAuth2 tokens (access token, ID token, etc.).
	token, err := h.exchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		log.Printf("ERREUR échange de code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de l'échange du code",
		})
		return
	}

	// Verify the ID token and extract user claims (e.g., email, name, subject).
	claims, err := h.verifyIDTokenAndExtractClaims(c.Request.Context(), token)
	if err != nil {
		log.Printf("ERREUR vérification ID token ou extraction claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur d'authentification",
		})
		return
	}

	// Find or create the user in the application's database based on the OIDC claims.
	user, err := h.authService.FindOrCreateUserFromClaims(claims)
	if err != nil {
		log.Printf("ERREUR FindOrCreateUserFromClaims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la gestion de l'utilisateur",
		})
		return
	}

	// Set the authenticated user in the session.
	session.Set("user", user)
	// Save the session to persist the user information.
	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde utilisateur: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur de sauvegarde de session",
		})
		return
	}

	// Redirect the user to the profile page after successful login.
	c.Redirect(http.StatusFound, "/profile")
}

// exchangeCodeForToken exchanges the authorization code received from the OIDC provider for an OAuth2 token.
func (h *AuthHandlers) exchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := h.authService.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// verifyIDTokenAndExtractClaims verifies the integrity and authenticity of the ID token
// and extracts the user claims (e.g., email, name, subject).
func (h *AuthHandlers) verifyIDTokenAndExtractClaims(ctx context.Context, token *oauth2.Token) (struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}, error) {
	// Extract the raw ID token string from the OAuth2 token.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Sub   string `json:"sub"`
		}{}, fmt.Errorf("ID token manquant")
	}

	// Verify the ID token using the OIDC provider's verifier.
	verifier := h.authService.Provider.Verifier(&oidc.Config{ClientID: h.authService.Oauth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Sub   string `json:"sub"`
		}{}, fmt.Errorf("erreur de vérification du token: %w", err)
	}

	// Extract claims into a custom struct.
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Sub   string `json:"sub"`
		}{}, fmt.Errorf("erreur d'extraction des informations utilisateur: %w", err)
	}
	return claims, nil
}

// LogoutHandler handles user logout.
// It clears the session and redirects the user to the application's root URL.
func (h *AuthHandlers) LogoutHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)

	// Log user information before clearing the session.
	if user := session.Get("user"); user != nil {
		log.Printf("Déconnexion de l'utilisateur: %v", user)
	}

	// Clear all session data.
	session.Clear()

	// Save the empty session to ensure changes are persisted.
	if err := session.Save(); err != nil {
		log.Printf("Erreur lors de la sauvegarde de session vide: %v", err)
	}

	// Manually delete the session cookie (optional, as session.Save() might handle this).
	cookieName := h.cfg.CookieName
	c.SetCookie(cookieName, "", -1, "/", "", false, true)

	// Redirect the user to the application's base URL after logout.
	c.Redirect(http.StatusFound, h.cfg.AppURL)
}

// generateRandomState generates a cryptographically secure random string to be used as an OAuth2 state parameter.
// This helps prevent Cross-Site Request Forgery (CSRF) attacks.
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
