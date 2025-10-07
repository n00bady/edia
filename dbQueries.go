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
	// Make sure that the directory exists (old android version needs this)
	log.Printf("Initializing database...")
	err := os.MkdirAll(filepath.Dir(dbPath), 0o755)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	log.Printf("Opening the database in: %s", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	createTableSQL := `
        CREATE TABLE IF NOT EXISTS entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
			NickName TEXT NOT NUll,
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
        INSERT INTO entries (NickName, Timestamp, LandLord, Renter, Size, Type, Rent, Start, End) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	log.Printf("Saving new entry in the database...")
	result, err := db.Exec(insertSQL, entry.NickName, entry.Timestamp, entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End)
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
	// Get the IDs of the coordinates first
	getCoordsSQL := `SELECT id FROM coordinates WHERE entry_id = ? ORDER BY id`
	getCoords, err := db.Query(getCoordsSQL, entry.ID)
	if err != nil {
		return fmt.Errorf("could not query the database for the coordinates IDs: %s", err)
	}
	var coord_IDs []int
	for getCoords.Next() {
		var id int
		err := getCoords.Scan(&id)
		if err != nil {
			return fmt.Errorf("could not get coordinates IDs for entery: %d", entry.ID)
		}
		coord_IDs = append(coord_IDs, id)
	}
	log.Printf("coord_IDs: %v", coord_IDs)
	getCoords.Close()

	// Now we can update the the entries and the coordinates tables properly
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting a database transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	updateSQL := `
		UPDATE entries SET LandLord = ?, Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?`
	updateSQLCoords := `
		UPDATE coordinates SET Latitude = ?, Longitude = ? WHERE entry_id = ? AND id = ?`

	res, err := tx.Exec(updateSQL, entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID)
	if err != nil {
		return fmt.Errorf("error updating entry with id: %d: %v", entry.ID, err)
	}
	rowsAff, RowErr := res.RowsAffected()
	log.Printf("Rows Affected: %d", rowsAff)
	log.Printf("Result error: %s", RowErr)

	if len(entry.Coords) > 4 || len(entry.Coords) < 1 {
		tx.Rollback()
		return fmt.Errorf("expected up to 4 pair of coordinates got %d, for entry ID: %d", len(entry.Coords), entry.ID)
	}

	for i := range len(entry.Coords) {
		res, err = tx.Exec(updateSQLCoords, entry.Coords[i].Latitude, entry.Coords[i].Longitude, entry.ID, coord_IDs[i])
		tmpRows, rowErr := res.RowsAffected()
		log.Printf("Update coordinates %d rows affected: %d", i+1, tmpRows)
		log.Printf("Result error: %s", rowErr)
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
		err := result.Scan(&entry.ID, &entry.NickName, &entry.Timestamp, &entry.LandlordName, &entry.RenterName, &entry.Size, &entry.Type, &entry.Rent, &entry.Start, &entry.End)
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
	err = resetAutoIncrement(db)
	if err != nil {
		return fmt.Errorf("reset autoincrement error")
	}

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
