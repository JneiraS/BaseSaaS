package handlers

import (
	"context"
	"log"
	"net/http"

	m "github.com/JneiraS/BaseSasS/domain/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var (
	provider *oidc.Provider
)

// Callback après login
func CallbackHandler(c *gin.Context) {
	session := sessions.Default(c)

	// Vérifier le state
	savedState := session.Get("state")
	queryState := c.Query("state")

	log.Printf("Callback - Saved state: %v, Query state: %s", savedState, queryState)

	if savedState == nil {
		log.Printf("Aucun state sauvegardé en session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun state en session"})
		return
	}

	if savedState.(string) != queryState {
		log.Printf("State invalide")
		c.JSON(http.StatusBadRequest, gin.H{"error": "State invalide"})
		return
	}

	// Vérifier s'il y a une erreur dans la callback
	if errMsg := c.Query("error"); errMsg != "" {
		log.Printf("Erreur OAuth: %s - %s", errMsg, c.Query("error_description"))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Erreur OAuth",
			"details": errMsg,
		})
		return
	}

	// Échanger le code contre un token
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code manquant"})
		return
	}

	log.Printf("Échange du code: %s", code)

	token, err := Oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Erreur échange token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur échange token",
			"details": err.Error(),
		})
		return
	}

	// Extraire l'ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Printf("Pas d'ID token dans la réponse")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Pas d'ID token"})
		return
	}

	// Vérifier l'ID token
	verifier := provider.Verifier(&oidc.Config{ClientID: Oauth2Config.ClientID})
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Printf("Token invalide: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token invalide",
			"details": err.Error(),
		})
		return
	}

	// Extraire les infos utilisateur
	var user m.User
	if err := idToken.Claims(&user); err != nil {
		log.Printf("Erreur claims: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur claims",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Utilisateur connecté: %+v", user)

	// Nettoyer et sauvegarder la session
	session.Clear()
	session.Set("user", user)

	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde session utilisateur: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erreur sauvegarde session",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Session utilisateur sauvegardée avec succès")
	c.Redirect(http.StatusFound, "/profile")
}
