package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

	createLandLordDetails := `
		CREATE TABLE IF NOT EXISTS landlordDetails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry_id INTEGER NOT NULL,
			firstName TEXT NOT NULL,
			lastName TEXT NOT NULL,
			fathersName TEXT,
			afm INTEGER,
			adt TEXT,
			ata INTEGER,
			e9 BLOB,
			notes TEXT
		);
	`

	log.Printf("Creating table LandLordDetails...")
	_, err = db.Exec(createLandLordDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating landlordDetails table: %v", err)
	}

	log.Printf("Tables created successfully!")

	return db, nil
}

// Save an entry
func saveEntry(db *sql.DB, entry Entry) error {
	insertSQL := `
        INSERT INTO entries (NickName, Timestamp, Renter, Size, Type, Rent, Start, End) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	log.Printf("Saving new entry in the database...")
	result, err := db.Exec(insertSQL, entry.NickName, entry.Timestamp, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End)
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

	log.Printf("Saving the landlordDetails...")
	insertLandLordDetails := `
		INSERT INTO landlordDetails (entry_id, firstName, lastName, fathersName, afm, adt, ata, e9, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, e := range entry.LandlordName {
		_, err := db.Exec(insertLandLordDetails, entry.ID, e.FirstName, e.LastName, e.FathersName, e.AFM, e.ADT, e.ATA, e.E9, e.Notes)
		if err != nil {
			return fmt.Errorf("error inserting land lord details: %v", err)
		}
		log.Printf("Inserting into landlordDetails: %s, %s", e.FirstName, e.LastName)
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
			return fmt.Errorf("could not get coordinates IDs for entry: %d", entry.ID)
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
		UPDATE entries SET  Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?`
	updateSQLCoords := `
		UPDATE coordinates SET Latitude = ?, Longitude = ? WHERE entry_id = ? AND id = ?`
	// updateLandLords := `
	// 	UPDATE landlordDetails SET firstName = ?, lastName = ?, fathersName = ?, afm = ?, adt = ?, ata = ?, e9 = ?, notes = ? WHERE entry_id = ? AND id = ?`

	// landlordNamesJSON, _ := json.Marshal(entry.LandlordName)
	res, err := tx.Exec(updateSQL, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID)
	if err != nil {
		return fmt.Errorf("error updating entry %d: %v", entry.ID, err)
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

	// for i, e := range entry.LandlordName {
	// 	res, err = tx.Exec(updateLandLords, e.FirstName, e.LastName, e.FathersName, e.AFM, e.ADT, e.ATA, e.E9, e.Notes, entry.ID, landlordsIds[i])
	// 	if err != nil {
	// 		return fmt.Errorf("error updating entry %d: %v", entry.ID, err)
	// 	}
	// 	rowsAff, rowsErr := res.RowsAffected()
	// 	log.Printf("Rows Affected: %d", rowsAff)
	// 	log.Printf("Result error: %s", rowsErr)
	// }

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error commiting transaction to the database: %v", err)
	}

	err = replaceLandlordsForEntry(db, entry)
	if err != nil {
		return err
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
		err := result.Scan(&entry.ID, &entry.NickName, &entry.Timestamp, &entry.RenterName, &entry.Size, &entry.Type, &entry.Rent, &entry.Start, &entry.End)
		if err != nil {
			return nil, fmt.Errorf("error retrieving entries from database: %v", err)
		}

		selectLandLords := `SELECT firstName, lastName, fathersName, afm, adt, ata, e9, notes FROM landlordDetails WHERE entry_id = ?`

		var landlords []LandlordDetails
		res, err := db.Query(selectLandLords, entry.ID)
		if err != nil {
			return nil, fmt.Errorf("error quering the land lords database: %v", err)
		}
		for res.Next() {
			var landlord LandlordDetails
			err := res.Scan(&landlord.FirstName, &landlord.LastName, &landlord.FathersName, &landlord.AFM, &landlord.ADT, &landlord.ATA, &landlord.E9, &landlord.Notes)
			if err != nil {
				return nil, fmt.Errorf("error retrieving land lords details for entry %d: %v", entry.ID, err)
			}

			landlords = append(landlords, landlord)
		}
		entry.LandlordName = landlords

		entries = append(entries, entry)
	}
	log.Printf("Retrieved the results successfully!")
	log.Printf("RESULTS:\n%v", entries)

	return entries, nil
}

func getEntry(db *sql.DB, id int) (*Entry, error) {
	selectSQL := `
		SELECT * FROM entries WHERE ID = ?
	`
	selectSQLForCoords := `
		SELECT latitude, longitude FROM coordinates WHERE entry_id = ?
	`
	selectSQLForLandLords := `
		SELECT firstName, lastName, fathersName, afm, adt, ata, e9, notes FROM landlordDetails WHERE entry_id = ?
	`

	var entry Entry

	log.Printf("Quering the database for entry with id: %d", id)
	result := db.QueryRow(selectSQL, id)
	err := result.Scan(&entry.ID, &entry.NickName, &entry.Timestamp, &entry.RenterName, &entry.Size, &entry.Type, &entry.Rent, &entry.Start, &entry.End)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving entry with id: %d: %v", id, err)
	}

	resultLandLords, err := db.Query(selectSQLForLandLords, id)
	if err != nil {
		return nil, err
	}
	for resultLandLords.Next() {
		var landlord LandlordDetails
		err := resultLandLords.Scan(&landlord.FirstName, &landlord.LastName, &landlord.FathersName, &landlord.AFM, &landlord.ADT, &landlord.ATA, &landlord.E9, &landlord.Notes)
		if err != nil {
			return nil, err
		}
		entry.LandlordName = append(entry.LandlordName, landlord)
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

	delLandLords := `
		DELETE FROM landlordDetails WHERE id = ?
	`

	log.Printf("Deleting entry %d and it's coordinates", id)

	log.Printf("Deleting coordinates for entry: %d", id)
	_, err := db.Exec(delCoordsSQL, id)
	if err != nil {
		return fmt.Errorf("could not delete coordinates for entry %d", id)
	}

	log.Printf("Deleting landlord details for entry: %d", id)
	_, err = db.Exec(delLandLords, id)
	if err != nil {
		return fmt.Errorf("could not delete landlord details for entry %d", id)
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

func getAllLords(db *sql.DB) ([]LandlordDetails, error) {
	getLandLords := `
	SELECT 
		firstName, 
		lastName, 
		GROUP_CONCAT(entry_id) 	AS owned_entry_ids 
	FROM landlordDetails 
	GROUP BY firstName, lastName 
	ORDER BY lastName, firstName;
	`

	log.Printf("Quering landlordDetails database...")
	result, err := db.Query(getLandLords)
	if err != nil {
		return nil, fmt.Errorf("error quering the landlords database: %v", err)
	}

	var landlords []LandlordDetails
	var entryIDs string
	for result.Next() {
		var landlord LandlordDetails
		err := result.Scan(&landlord.FirstName, &landlord.LastName, &entryIDs)
		if err != nil {
			return nil, fmt.Errorf("error retrieving landlordDetails from database: %v", err)
		}
		idsStr := strings.Split(entryIDs, ",")
		for _, s := range idsStr {
			id, _ := strconv.ParseInt(s, 10, 64)
			landlord.Entry_IDs = append(landlord.Entry_IDs, int(id))
		}

		landlords = append(landlords, landlord)
	}

	return landlords, nil
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

	_, err = db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM landlordDetails) WHERE name = 'landlordDetails'")
	if err != nil {
		return err
	}

	return nil
}

// TODO: queries for searching a variety of fields
func replaceLandlordsForEntry(db *sql.DB, entry Entry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM landlordDetails WHERE entry_id = ?`, entry.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	stm, err := tx.Prepare(`
		INSERT INTO landlordDetails (entry_id, firstName, lastName, fathersName, afm,adt, ata, e9,notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stm.Close()

	for _, l := range entry.LandlordName {
		_, err := stm.Exec(entry.ID, l.FirstName, l.LastName, l.FathersName, l.AFM, l.ADT, l.ATA, l.E9, l.Notes)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("inserting new landlords failed for entry %d with error: %v", entry.ID, err)
		}
	}

	return tx.Commit()
}
