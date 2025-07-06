package database

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabase initializes and returns a new GORM database connection.
// Currently, it is configured to use SQLite as the database.
func InitDatabase() (*gorm.DB, error) {
	// Open a connection to the SQLite database file named "basesass.db".
	// gorm.Config{} can be used to provide additional GORM configurations.
	db, err := gorm.Open(sqlite.Open("basesass.db"), &gorm.Config{})
	if err != nil {
		// If there's an error connecting to the database, wrap it and return.
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
