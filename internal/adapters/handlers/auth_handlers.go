package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandlers struct {
	authService *services.AuthService
}

func NewAuthHandlers(authService *services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

func (h *AuthHandlers) LoginHandler(c *gin.Context) {
	if !h.authService.IsConfigured() {
		log.Printf("ERREUR: AuthService n'est pas configuré")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service d'authentification temporairement indisponible",
			"message": "Le provider OIDC n'est pas accessible. Vérifiez que Zitadel fonctionne.",
		})
		return
	}

	session := sessions.Default(c)

	if user := session.Get("user"); user != nil {
		log.Printf("Utilisateur déjà connecté, redirection vers profile")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	state := generateRandomState()
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
	if !h.authService.IsConfigured() {
		log.Printf("ERREUR: AuthService non configuré")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Service d'authentification non disponible",
		})
		return
	}

	session := sessions.Default(c)
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

	ctx := context.Background()
	token, err := h.authService.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("ERREUR échange de code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur lors de l'échange du code",
			"details": err.Error(), // Ajout de cette ligne pour les détails de l'erreur
		})
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Printf("ERREUR: ID token manquant")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID token manquant",
		})
		return
	}

	verifier := h.authService.Provider.Verifier(&oidc.Config{ClientID: h.authService.Oauth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("ERREUR vérification ID token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur de vérification du token",
		})
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		log.Printf("ERREUR extraction des claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur d'extraction des informations utilisateur",
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

func (h *AuthHandlers) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

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
	cookieName := "session" // Nom du cookie utilisé par votre store
	c.SetCookie(cookieName, "", -1, "/", "localhost", false, true)

	// Rediriger l'utilisateur
	c.Redirect(http.StatusFound, "http://localhost:3000/")
}

func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-state-" + base64.StdEncoding.EncodeToString([]byte("simple-fallback"))
	}
	return base64.StdEncoding.EncodeToString(b)
}
