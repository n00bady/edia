package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	
	_ "github.com/mattn/go-sqlite3"
)

// Initialize the DB if it doesn't exists
func initDB(dbPath string) (*sql.DB, error) {
	// Make sure that the directory exists (old android version need this)
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// This is a placeholder needs to be change according with the Entry struct!
	createTableSQL := `
        CREATE TABLE IF NOT EXISTS entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
			Timestamp DATETIME NOT NULL,
			LandLord TEXT NOT NULL,
			Renter TEXT NOT NULL,
			Size REAL NOT NULL,
			Type TEXT NOT NULL,
			Rent REAL NOT NULL,
			Start DATETIME NOT NULL,
			End DATETIME NOT NULL
        );
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}
	createCoordsTable := `
		CREATE TABLE IF NOT EXISTS coordinates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry_id INTEGER NOT NULL,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			FOREIGN KEY (entry_id) REFERENCES entries(id)
		);
	`
	_, err = db.Exec(createCoordsTable)
	if err != nil {
		return nil, fmt.Errorf("error creating coordinates table: %v", err)
	}

	return db, nil
}

// Save an entry
func saveEntry(db *sql.DB, entry Entry) error {
	insertSQL := `
        INSERT INTO entries (Timestamp, LandLord, Renter, Size, Type, Rent, Start, End) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(insertSQL, entry.Timestamp, entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End)
	if err != nil {
		return fmt.Errorf("error inserting entry: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	entry.ID = int(id)

	insertCoordsSQL := `
		INSERT INTO coordinates (entry_id, latitude, longitude) VALUES (?, ?, ?)`
	for i := range 4 {
		_, err := db.Exec(insertCoordsSQL, entry.ID, entry.Coords[i].Latitude, entry.Coords[i].Longitude)
		if err != nil {
			return fmt.Errorf("error inserting entry coordinate: %v", err)
		}
	}

	return nil
}

// TODO: query to retrieve and entry for display
func getAll(db *sql.DB) ([]Entry, error) {
	selectSQL := `
		SELECT * FROM entries
	`
	result, err := db.Query(selectSQL)
	if err != nil {
		return nil, fmt.Errorf("error quering the database: %v", err)
	}

	var entries []Entry
	for result.Next() {
		var entry Entry
		err := result.Scan(&entry.ID, &entry.Timestamp, &entry.LandlordName, &entry.RenterName, &entry.Size, &entry.Type, &entry.Rent, &entry.Start, &entry.End)
		if err != nil {
			return nil, fmt.Errorf("error retrieving entries from database: %v", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
// TODO: queries for searching a variety of fields
