package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/domain/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Variable globale pour le provider (doit être définie dans main.go)
var Provider *oidc.Provider

// CallbackHandler gère le callback OAuth2
func CallbackHandler(c *gin.Context) {
	// Vérifier que le provider et la config OAuth2 sont disponibles
	if Provider == nil {
		log.Printf("ERREUR: Provider OIDC non initialisé")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service d'authentification non disponible",
			"message": "Le provider OIDC n'est pas initialisé",
		})
		return
	}

	if !isOAuth2Configured() {
		log.Printf("ERREUR: OAuth2Config non initialisé")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Configuration OAuth2 non disponible",
		})
		return
	}

	session := sessions.Default(c)

	// Récupérer le code et le state depuis les paramètres de requête
	code := c.Query("code")
	state := c.Query("state")

	log.Printf("Callback reçu - Code: %s, State: %s", code, state)

	// Vérifier le state pour éviter les attaques CSRF
	savedState := session.Get("state")
	if savedState == nil {
		log.Printf("ERREUR: Aucun state trouvé en session")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State manquant en session",
		})
		return
	}

	if state != savedState.(string) {
		log.Printf("ERREUR: State invalide. Reçu: %s, Attendu: %s", state, savedState)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State invalide - possible attaque CSRF",
		})
		return
	}

	// Nettoyer le state de la session
	session.Delete("state")

	// Échanger le code contre un token
	ctx := context.Background()
	token, err := Oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("ERREUR échange de code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur lors de l'échange du code",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Token obtenu avec succès")

	// Extraire l'ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Printf("ERREUR: ID token manquant")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID token manquant dans la réponse",
		})
		return
	}

	// Vérifier l'ID token
	verifier := Provider.Verifier(&oidc.Config{ClientID: Oauth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("ERREUR vérification ID token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur de vérification du token",
			"details": err.Error(),
		})
		return
	}

	// Extraire les claims
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Sub           string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		log.Printf("ERREUR extraction des claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur d'extraction des informations utilisateur",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Claims extraits: %+v", claims)

	// Créer l'objet utilisateur
	user := models.User{
		ID:    claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}

	// Sauvegarder l'utilisateur en session
	session.Set("user", user)
	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde utilisateur en session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur de sauvegarde de session",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Utilisateur connecté avec succès: %s", user.Email)

	// Rediriger vers le profil
	c.Redirect(http.StatusFound, "/profile")
}
