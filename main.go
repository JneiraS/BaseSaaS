package main

import (
	"encoding/gob"
	"log"

	"github.com/JneiraS/BaseSasS/internal/adapters/handlers"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/joho/godotenv"
)

func LoadEnv() error {
	return godotenv.Load()
}

// init est exécuté avant main pour enregistrer les types nécessaires à l'encodage de session.
func init() {
	gob.Register(models.User{})
}

func main() {
	if err := LoadEnv(); err != nil {
		log.Printf("Avertissement: Impossible de charger .env: %v", err)
	}

	app, err := handlers.NewApp()
	if err != nil {
		log.Fatalf("Erreur lors de la création de l'application: %v", err)
	}

	app.Run()
}
