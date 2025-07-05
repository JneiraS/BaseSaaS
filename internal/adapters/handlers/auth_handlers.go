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

type AuthHandlers struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthHandlers(authService *services.AuthService, cfg *config.Config) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandlers) LoginHandler(c *gin.Context) {
	

	session := c.MustGet("session").(sessions.Session)

	if user := session.Get("user"); user != nil {
		log.Printf("Utilisateur déjà connecté, redirection vers profile")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	state, err := generateRandomState()
	if err != nil {
		log.Printf("ERREUR génération de l'état: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur interne du serveur",
		})
		return
	}
	session.Set("state", state)

	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur session",
		})
		return
	}

	authURL := h.authService.Oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandlers) CallbackHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	code := c.Query("code")
	state := c.Query("state")

	savedState := session.Get("state")
	if savedState == nil || state != savedState.(string) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State invalide",
		})
		return
	}

	session.Delete("state")

	token, err := h.exchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		log.Printf("ERREUR échange de code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de l'échange du code",
		})
		return
	}

	claims, err := h.verifyIDTokenAndExtractClaims(c.Request.Context(), token)
	if err != nil {
		log.Printf("ERREUR vérification ID token ou extraction claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur d'authentification",
		})
		return
	}

	user, err := h.authService.FindOrCreateUserFromClaims(claims)
	if err != nil {
		log.Printf("ERREUR FindOrCreateUserFromClaims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la gestion de l'utilisateur",
		})
		return
	}

	session.Set("user", user)
	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde utilisateur: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur de sauvegarde de session",
		})
		return
	}

	c.Redirect(http.StatusFound, "/profile")
}

func (h *AuthHandlers) exchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := h.authService.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (h *AuthHandlers) verifyIDTokenAndExtractClaims(ctx context.Context, token *oauth2.Token) (struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Sub   string `json:"sub"`
		}{}, fmt.Errorf("ID token manquant")
	}

	verifier := h.authService.Provider.Verifier(&oidc.Config{ClientID: h.authService.Oauth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Sub   string `json:"sub"`
		}{}, fmt.Errorf("erreur de vérification du token: %w", err)
	}

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

func (h *AuthHandlers) LogoutHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)

	// Log utilisateur avant suppression
	if user := session.Get("user"); user != nil {
		log.Printf("Déconnexion de l'utilisateur: %v", user)
	}

	// Supprimer les données de session
	session.Clear()

	// Sauvegarder la session vide
	if err := session.Save(); err != nil {
		log.Printf("Erreur lors de la sauvegarde de session vide: %v", err)
	}

	// Supprimer le cookie manuellement (optionnel si Save le fait déjà)
	cookieName := h.cfg.CookieName
	c.SetCookie(cookieName, "", -1, "/", "", false, true)

	// Rediriger l'utilisateur
	c.Redirect(http.StatusFound, h.cfg.AppURL)
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
