package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Initialize the DB if it doesn't exists
func initDB(dbPath string) (*sql.DB, error) {
	// Make sure that the directory exists (old android version need this)
	log.Printf("Initializing database...")
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	log.Printf("Opening the database in: %s", dbPath)
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
			Start TEXT NOT NULL,
			End TEXT NOT NULL
        );
	`

	log.Printf("Creating table entries if it doesn't exists...")
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

	log.Printf("Creating coordinates table if it doesn't exist...")
	_, err = db.Exec(createCoordsTable)
	if err != nil {
		return nil, fmt.Errorf("error creating coordinates table: %v", err)
	}

	log.Printf("Tables created successfully!")

	return db, nil
}

// Save an entry
func saveEntry(db *sql.DB, entry Entry) error {
	insertSQL := `
        INSERT INTO entries (Timestamp, LandLord, Renter, Size, Type, Rent, Start, End) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	log.Printf("Saving new entry in the database...")
	result, err := db.Exec(insertSQL, entry.Timestamp, entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End)
	if err != nil {
		return fmt.Errorf("error inserting entry: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	entry.ID = int(id)

	log.Printf("Saving the coordinates in the database...")
	insertCoordsSQL := `
		INSERT INTO coordinates (entry_id, latitude, longitude) VALUES (?, ?, ?)`
	for i := range 4 {
		_, err := db.Exec(insertCoordsSQL, entry.ID, entry.Coords[i].Latitude, entry.Coords[i].Longitude)
		if err != nil {
			return fmt.Errorf("error inserting entry coordinate: %v", err)
		}
	}

	log.Printf("Saved entry successfully!")

	return nil
}

// Update an Entry
func updateEntry(db *sql.DB, entry Entry) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting a database transaction: %v", err)
	}

	updateSQL := `
		UPDATE entries SET LandLord = ?, Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?`
	updateSQLCoords := `
		UPDATE coordinates SET Latitude = ?, Longitude = ? WHERE entry_id = ?`

	res, err := tx.Exec(updateSQL, entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID)
	log.Printf("Update entries result: %s", res)
	if err != nil {
		return fmt.Errorf("error updating entry with id: %d: %v", entry.ID, err)
	}

	for i := range 4 {
		res, err = tx.Exec(updateSQLCoords, entry.Coords[i].Latitude, entry.Coords[i].Longitude, entry.ID)
		log.Printf("Update coordinates %d result: %s", i, res)
		if err != nil {
			return fmt.Errorf("error updating coordinates for entry with id: %d: %v", entry.ID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error commiting transaction: %v", err)
	}

	log.Printf("Updated entry %d successfully!", entry.ID)

	return nil
}

func getAll(db *sql.DB) ([]Entry, error) {
	selectSQL := `
		SELECT * FROM entries
	`

	log.Printf("Quering database...")
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
	log.Printf("Retrieved the results successfully!")
	log.Printf("RESULTS:\n%v", entries)

	return entries, nil
}

func getEntry(db *sql.DB, id int) (*Entry, error) {
	selectSQL := `
		Select * From entries Where ID = ?
	`
	selectSQLForCoords := `
		Select latitude, longitude From coordinates Where entry_id = ?
	`
	var entry Entry

	log.Printf("Quering the database for entry with id: %d", id)
	result := db.QueryRow(selectSQL, id)
	err := result.Scan(&entry.ID, &entry.Timestamp, &entry.LandlordName, &entry.RenterName, &entry.Size, &entry.Type, &entry.Rent, &entry.Start, &entry.End)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving entry with id: %d: %v", id, err)
	}

	resultCoords, err := db.Query(selectSQLForCoords, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving coordinates for entry with id: %d: %v", id, err)
	}
	for resultCoords.Next() {
		var coord Coordinate
		err := resultCoords.Scan(&coord.Latitude, &coord.Longitude)
		if err != nil {
			return nil, fmt.Errorf("error retrieving coordinates for entry with id: %d: %v", id, err)
		}
		entry.Coords = append(entry.Coords, coord)
	}

	return &entry, nil
}

func delEntry(db *sql.DB, id int) error {
	delCoordsSQL := `
		Delete From coordinates Where entry_id = ?
	`
	delEntrySQL2 := `
		Delete From entries Where id = ?
	`
	log.Printf("Deleting entry %d and it's coordinates", id)

	log.Printf("Deleting coordinates for entry: %d", id)
	_, err := db.Exec(delCoordsSQL, id)
	if err != nil {
		return fmt.Errorf("could not delete coordinates for entry %d", id)
	}

	log.Printf("Deleting entry %d from entries.", id)
	_, err = db.Exec(delEntrySQL2, id)
	if err != nil {
		return fmt.Errorf("could not delete entry: %d", id)
	}

	log.Printf("Reseting sqlite autoincrement.")
	resetAutoIncrement(db)

	log.Printf("Deleted entry %d successfully!", id)

	return nil
}

func resetAutoIncrement(db *sql.DB) error {
	_, err := db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM entries) WHERE name = 'entries'")
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM coordinates) WHERE name = 'coordinates'")
	if err != nil {
		return err
	}

	return nil
}

// TODO: queries for searching a variety of fields
