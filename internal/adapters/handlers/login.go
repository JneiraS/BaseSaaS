package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var (
	Oauth2Config *oauth2.Config
)

// Vérifier si l'OAuth2 est configuré
func isOAuth2Configured() bool {
	return Oauth2Config != nil
}

// Redirection vers login Zitadel
func LoginHandler(c *gin.Context) {
	// Vérifier si OAuth2 est configuré
	if !isOAuth2Configured() {
		log.Printf("ERREUR: OAuth2Config n'est pas initialisé")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service d'authentification temporairement indisponible",
			"message": "Le provider OIDC n'est pas accessible. Vérifiez que Zitadel fonctionne.",
		})
		return
	}

	session := sessions.Default(c)

	// Vérifier si l'utilisateur est déjà connecté
	if user := session.Get("user"); user != nil {
		log.Printf("Utilisateur déjà connecté, redirection vers profile")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Générer un state pour sécuriser la requête
	state := generateRandomState()
	log.Printf("State généré: %s", state)

	// Sauvegarder le state
	session.Set("state", state)

	// Essayer de sauvegarder avec gestion d'erreur détaillée
	if err := session.Save(); err != nil {
		log.Printf("ERREUR sauvegarde session dans loginHandler: %v", err)
		log.Printf("Type d'erreur: %T", err)

		// Essayer de créer une nouvelle session
		session.Clear()
		session.Set("state", state)
		if err2 := session.Save(); err2 != nil {
			log.Printf("ERREUR même après clear: %v", err2)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Erreur session",
				"details": err.Error(),
			})
			return
		}
	}

	log.Printf("State sauvegardé avec succès")

	// Rediriger vers Zitadel
	authURL := Oauth2Config.AuthCodeURL(state)
	log.Printf("Redirection vers: %s", authURL)
	c.Redirect(http.StatusFound, authURL)
}

// Générer un state aléatoire
func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Erreur génération state: %v", err)
		// Fallback simple si crypto/rand échoue
		return "fallback-state-" + base64.StdEncoding.EncodeToString([]byte("simple-fallback"))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Logout
func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	// Log pour debug
	if user := session.Get("user"); user != nil {
		log.Printf("Déconnexion de l'utilisateur: %v", user)
	}

	session.Clear()
	if err := session.Save(); err != nil {
		log.Printf("Erreur lors de la sauvegarde de session vide: %v", err)
	}

	// Redirection vers logout Zitadel (optionnel)
	// Si OAuth2 est configuré, on peut rediriger vers le logout Zitadel
	if isOAuth2Configured() {
		// logoutURL := "http://localhost:8080/oidc/v1/end_session?post_logout_redirect_uri=http://localhost:3000/"
		logoutURL := "http://localhost:3000/"
		c.Redirect(http.StatusFound, logoutURL)
	} else {
		c.Redirect(http.StatusFound, "http://localhost:3000/")
	}
}
