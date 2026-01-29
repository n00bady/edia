package main

import (
	"database/sql"
	"errors"
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
			name TEXT NOT NUll,
			timestamp DATETIME NOT NULL,
			atak INTEGER NOT NULL,
			kaek TEXT NOT NULL,
			size REAL NOT NULL,
			type TEXT NOT NULL,
			rent REAL NOT NULL,
			start TEXT NOT NULL,
			end TEXT NOT NULL
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
			entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL
		);
	`

	log.Printf("Creating coordinates table if it doesn't exist...")
	_, err = db.Exec(createCoordsTable)
	if err != nil {
		return nil, fmt.Errorf("error creating coordinates table: %v", err)
	}

	createOwnerDetails := `
		CREATE TABLE IF NOT EXISTS ownerDetails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			firstName TEXT NOT NULL,
			lastName TEXT NOT NULL,
			fathersName TEXT,
			afm INTEGER,
			adt TEXT,
			e9 BLOB,
			notes TEXT
		);
	`

	log.Printf("Creating table ownerDetails...")
	_, err = db.Exec(createOwnerDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating ownerDetails table: %v", err)
	}

	createRenterDetails := `
		CREATE TABLE IF NOT EXISTS renterDetails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			firstName TEXT NOT NULL,
			lastName TEXT NOT NULL,
			fathersName TEXT,
			afm INTEGER,
			adt TEXT,
			e9 BLOB,
			notes TEXT
		);
	`

	log.Println("Creating table renterDetails...")
	_, err = db.Exec(createRenterDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating renterDetails tables: %v", err)
	}

	createJunctionEntriesOnwer := `
		CREATE TABLE IF NOT EXISTS entries_owner (
			entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
			owner_id INTEGER NOT NULL REFERENCES ownerDetails(id) ON DELETE CASCADE
		);
	`
	log.Println("Creating junction table entries_owner...")
	_, err = db.Exec(createJunctionEntriesOnwer)
	if err != nil {
		return nil, fmt.Errorf("error creating juction table entries_owner: %v", err)
	}

	createJunctionEntriesRenter := `
		CREATE TABLE IF NOT EXISTS entries_renter (
			entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
			renter_id INTEGER NOT NULL REFERENCES renterDetails(id) ON DELETE CASCADE
		);
	`
	log.Println("Creating junction table entries_renter...")
	_, err = db.Exec(createJunctionEntriesRenter)
	if err != nil {
		return nil, fmt.Errorf("error creating juction table entries_renter: %v", err)
	}

	log.Printf("Tables created successfully!")

	return db, nil
}

func saveEntry(db *sql.DB, entry Entry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec(`
		INSERT INTO entries (name, timestamp, atak, kaek, size, type, rent, start, end)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.Name, entry.Timestamp, entry.ATAK, entry.KAEK, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End)
	if err != nil {
		return err
	}

	entryID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// get or create onwer(s)
	for _, o := range entry.Owners {
		ownerID, err := getOrCreateOwner(tx, o)
		if err != nil {
			return err
		}

		// link owner to entry on the junction tablle
		_, err = tx.Exec(`
		INSERT OR IGNORE INTO entries_owner (entry_id, owner_id)
		VALUES (?, ?)`,
			entryID, ownerID)
		if err != nil {
			return err
		}
	}

	// ger or create renter(s)
	for _, r := range entry.Renters {
		renterID, err := getOrCreateRenters(tx, r)
		if err != nil {
			return err
		}

		// link owner to entry on the junction tablle
		_, err = tx.Exec(`
		INSERT OR IGNORE INTO entries_renter (entry_id, renter_id)
		VALUES (?, ?)`,
			entryID, renterID)
		if err != nil {
			return err
		}

	}

	// coordinates
	for _, c := range entry.Coords {
		_, err := tx.Exec(`
			INSERT INTO coordinates (entry_id, latitude, longitude)
			VALUES (?, ?, ?)`,
			entryID, c.Latitude, c.Longitude)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getOrCreateOwner(tx *sql.Tx, o OwnerDetails) (int64, error) {
	var ownerID int64

	query := `SELECT id FROM ownerDetails WHERE firstName = ? AND lastName = ?`
	err := tx.QueryRow(query, o.FirstName, o.LastName).Scan(&ownerID)
	if err == nil {
		return ownerID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO ownerDetails (firstName, lastName, fathersName, afm, adt, e9, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.FirstName, o.LastName, o.FathersName, o.AFM, o.ADT, o.E9, o.Notes)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func getOrCreateRenters(tx *sql.Tx, r RenterDetails) (int64, error) {
	var renterID int64

	query := `SELECT id FROM renterDetails WHERE firstName = ? AND lastName = ?`
	err := tx.QueryRow(query, r.FirstName, r.LastName).Scan(&renterID)
	if err == nil {
		return renterID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO renterDetails (firstName, lastName, fathersName, afm, adt, e9, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		r.FirstName, r.LastName, r.FathersName, r.AFM, r.ADT, r.E9, r.Notes)
	if err != nil {
		return 0, nil
	}

	return res.LastInsertId()
}

func updateEntry(db *sql.DB, entry Entry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE entries
		SET name = ?, timestamp = ?, atak = ?, kaek = ?, size = ?, type = ?, rent = ?, start = ?, end = ?
		WHERE id = ?`,
		entry.Name, entry.Timestamp, entry.ATAK, entry.KAEK, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM entries_owner WHERE entry_id = ?`,
		entry.ID)
	if err != nil {
		return err
	}

	for _, o := range entry.Owners {
		ownerID, err := getOrCreateOwner(tx, o)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO entries_owner (entry_id, owner_id)
			VALUES (?, ?)`,
			entry.ID, ownerID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		DELETE FROM entries_renter WHERE entry_id = ?`,
		entry.ID)
	if err != nil {
		return err
	}

	for _, r := range entry.Renters {
		renterID, err := getOrCreateRenters(tx, r)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO entries_renter (entry_id, renter_id)
			VALUES (?, ?)`,
			entry.ID, renterID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		DELETE FROM coordinates WHERE entry_id = ?`,
		entry.ID)
	if err != nil {
		return err
	}

	for _, c := range entry.Coords {
		_, err := tx.Exec(`
			INSERT INTO coordinates (entry_id, latitude, longitude)
			VALUES (?, ?, ?)`,
			entry.ID, c.Latitude, c.Longitude)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getAllEntries(db *sql.DB) ([]Entry, error) {
	var entries []Entry

	rows, err := db.Query(`SELECT * FROM entries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e Entry

		err := rows.Scan(&e.ID, &e.Name, &e.Timestamp, &e.ATAK, &e.KAEK, &e.Size, &e.Type, &e.Rent, &e.Start, &e.End)
		if err != nil {
			return nil, err
		}

		// get the owners for the entry
		ownerRows, err := db.Query(`
			SELECT o.id, o.firstName, o.lastName, o.fathersName, o.afm, o.adt, o.e9, o.notes 
			FROM ownerDetails o
			JOIN entries_owner eo ON o.id = eo.owner_id
			WHERE eo.entry_id = ?`,
			e.ID)
		if err != nil {
			return nil, err
		}

		for ownerRows.Next() {
			var o OwnerDetails

			err := ownerRows.Scan(&o.ID, &o.FirstName, &o.LastName, &o.FathersName, &o.AFM, &o.ADT, &o.E9, &o.Notes)
			if err != nil {
				ownerRows.Close()
				return nil, err
			}
			e.Owners = append(e.Owners, o)
		}
		ownerRows.Close()

		// get the renters for the entry
		renterRows, err := db.Query(`
			SELECT r.id, r.firstName, r.lastName, r.fathersName, r.afm, r.adt, r.e9, r.notes
			FROM renterDetails r
			JOIN entries_renter er ON r.id = er.renter_id
			WHERE er.entry_id = ?`,
			e.ID)
		if err != nil {
			return nil, err
		}

		for renterRows.Next() {
			var r RenterDetails

			err := renterRows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.FathersName, &r.AFM, &r.ADT, &r.E9, &r.Notes)
			if err != nil {
				renterRows.Close()
				return nil, err
			}
			e.Renters = append(e.Renters, r)
		}
		renterRows.Close()

		// get the coordinates for the entry
		coordRows, err := db.Query(`
			SELECT *
			FROM coordinates
			WHERE entry_id = ?`,
			e.ID)
		if err != nil {
			return nil, err
		}

		for coordRows.Next() {
			var c Coordinates
			err := coordRows.Scan(&c.ID, &c.EntryID, &c.Latitude, &c.Longitude)
			if err != nil {
				coordRows.Close()
				return nil, err
			}
			e.Coords = append(e.Coords, c)
		}
		coordRows.Close()

		entries = append(entries, e)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func getEntry(db *sql.DB, id uint) (Entry, error) {
	var e Entry

	err := db.QueryRow(`
		SELECT * 
		FROM entries 
		WHERE id = ?`, id).Scan(&e.ID, &e.Name, &e.Timestamp, &e.ATAK, &e.KAEK, &e.Size, &e.Type, &e.Rent, &e.Start, &e.End)
	if err != nil {
		return e, err
	}

	// get Owners
	e.Owners, err = getOwners(db, e)
	if err != nil {
		return e, err
	}

	// get Renters
	e.Renters, err = getRenters(db, e)
	if err != nil {
		return e, err
	}

	// get coordinates
	e.Coords, err = getCoords(db, e)
	if err != nil {
		return e, err
	}

	return e, nil
}

func getOwners(db *sql.DB, e Entry) ([]OwnerDetails, error) {
	var owners []OwnerDetails

	rows, err := db.Query(`
		SELECT o.id, o.firstName, o.lastName, o.fathersName, o.afm, o.adt, o.e9, o.notes
		FROM ownerDetails o
		JOIN entries_owner eo ON o.id = eo.owner_id
		WHERE eo.entry_id = ?`,
		e.ID)
	if err != nil {
		return owners, err
	}
	defer rows.Close()

	for rows.Next() {
		var o OwnerDetails

		err := rows.Scan(&o.ID, &o.FirstName, &o.LastName, &o.FathersName, &o.AFM, &o.ADT, &o.E9, &o.Notes)
		if err != nil {
			return owners, err
		}
		owners = append(owners, o)
	}

	return owners, nil
}

func getRenters(db *sql.DB, e Entry) ([]RenterDetails, error) {
	var renters []RenterDetails

	rows, err := db.Query(`
		SELECT r.id, r.firstName, r.lastName, r.fathersName, r.afm, r.adt, r.e9, r.notes
		FROM renterDetails r
		JOIN entries_renter er ON r.id = er.renter_id
		WHERE er.entry_id = ?`,
		e.ID)
	if err != nil {
		return renters, err
	}
	defer rows.Close()

	for rows.Next() {
		var r RenterDetails

		err := rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.FathersName, &r.AFM, &r.ADT, &r.E9, &r.Notes)
		if err != nil {
			return renters, err
		}
		renters = append(renters, r)
	}

	return renters, nil
}

func getCoords(db *sql.DB, e Entry) ([]Coordinates, error) {
	var coordinates []Coordinates

	rows, err := db.Query(`
		SELECT id, entry_id, latitude, longitude
		FROM coordinates
		WHERE entry_id = ?`,
		e.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Coordinates

		err := rows.Scan(&c.ID, &c.EntryID, &c.Latitude, &c.Longitude)
		if err != nil {
			return nil, err
		}
		coordinates = append(coordinates, c)
	}

	return coordinates, nil
}

func delEntry(db *sql.DB, id uint) error {
	if id <= 0 {
		return errors.New("invalid entry id")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM entries WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("entry not found")
	}

	res, err := tx.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("entry not found")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func getAllOwners(db *sql.DB) ([]OwnerDetails, error) {
	var owners []OwnerDetails

	rows, err := db.Query(`SELECT * FROM ownerDetails`)
	if err != nil {
		return owners, err
	}

	for rows.Next() {
		var o OwnerDetails

		err := rows.Scan(&o.ID, &o.FirstName, &o.LastName, &o.FathersName, &o.AFM, &o.ADT, &o.E9, &o.Notes)
		if err != nil {
			return owners, err
		}
		owners = append(owners, o)
	}

	return owners, nil
}

func getAllRenters(db *sql.DB) ([]RenterDetails, error) {
	var renters []RenterDetails

	rows, err := db.Query(`SELECT * FROM renterDetails`)
	if err != nil {
		return renters, err
	}

	for rows.Next() {
		var r RenterDetails

		err := rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.FathersName, &r.AFM, &r.ADT, &r.E9, &r.Notes)
		if err != nil {
			return renters, err
		}
		renters = append(renters, r)
	}

	return renters, nil
}

func getOwnerEntries(db *sql.DB, o OwnerDetails) ([]Entry, error) {
	var entries []Entry

	rows, err := db.Query(`
		SELECT e.id, e.name, e.timestamp, e.atak, e.kaek, e.size, e.type, e.rent, e.start, e.end
		FROM entries e
		JOIN entries_owner eo ON e.id = eo.entry_id
		WHERE eo.owner_id = ?`,
		o.ID)
	if err != nil {
		return entries, err
	}

	for rows.Next() {
		var e Entry

		err := rows.Scan(&e.ID, &e.Name, &e.Timestamp, &e.ATAK, &e.KAEK, &e.Size, &e.Type, &e.Rent, &e.Start, &e.End)
		if err != nil {
			return entries, err
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func getRenterEntries(db *sql.DB, r RenterDetails) ([]Entry, error) {
	var entries []Entry

	rows, err := db.Query(`
		SELECT e.id, e.name, e.timestamp, e.atak, e.kaek, e.size, e.type, e.rent, e.start, e.end
		FROM entries e
		JOIN entries_renter er ON e.id = er.entry_id
		WHERE er.renter_id = ?`,
		r.ID)
	if err != nil {
		return entries, err
	}

	for rows.Next() {
		var e Entry

		err := rows.Scan(&e.ID, &e.Name, &e.Timestamp, &e.ATAK, &e.KAEK, &e.Size, &e.Type, &e.Rent, &e.Start, &e.End)
		if err != nil {
			return entries, err
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// might not need this anymore...
func resetAutoIncrement(db *sql.DB) error {
	_, err := db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM entries) WHERE name = 'entries'")
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM coordinates) WHERE name = 'coordinates'")
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM personDetails) WHERE name = 'personDetails'")
	if err != nil {
		return err
	}

	return nil
}
