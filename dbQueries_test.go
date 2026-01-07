package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSaveEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Test Fatal Error: %v", err)
	}
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	entry := Entry{
		Name:      "Test",
		Timestamp: now,
		Size:      150.5,
		Type:      "A",
		Rent:      1200.0,
		Start:     "2025-01-01",
		End:       "2025-12-31",
		Owners: []OwnerDetails{
			{FirstName: "Ιωάννης", LastName: "Ελάφις", FathersName: "Τζιμακος", AFM: 123456789},
			{FirstName: "Ιωάννα", LastName: "Σμιθιδη", FathersName: "Ιάκωβος", AFM: 987654321},
		},
		Renters: []RenterDetails{
			{FirstName: "Αλική", LastName: "Ναυτικού", AFM: 111222333},
		},
		Coords: []Coordinates{
			{Latitude: 37.9838, Longitude: 23.7275},
			{Latitude: 37.9840, Longitude: 23.7280},
		},
	}

	// Begin transaction
	mock.ExpectBegin()

	// Insert to entries
	mock.ExpectExec(`INSERT INTO entries \(name, timestamp, size, type, rent, start, end\) VALUES \(\?, \?, \?, \?, \?, \?, \?\)`).
		WithArgs("Test", now, 150.5, "A", 1200.0, "2025-01-01", "2025-12-31").
		WillReturnResult(sqlmock.NewResult(1, 1)) // lastInsertId = 1, rowsAffected = 1

	mock.ExpectQuery(`SELECT id FROM owners WHERE afm = \?`).
		WithArgs(123456789).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO owners .*`).
		WithArgs("Ιωάννης", "Ελάφις", "Τζιμακος", 123456789, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(10, 1))

	// Link first owner
	mock.ExpectExec(`INSERT OR IGNORE INTO entries_owner \(entry_id, owner_id\) VALUES \(\?, \?\)`).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Second owner - not found → insert → new ID=11
	mock.ExpectQuery(`SELECT id FROM owners WHERE afm = \?`).
		WithArgs(987654321).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO owners .*`).
		WithArgs("Ιωάννα", "Σμιθιδη", "Ιάκωβος", 987654321, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(11, 1))

	// Link second owner
	mock.ExpectExec(`INSERT OR IGNORE INTO entries_owner \(entry_id, owner_id\) VALUES \(\?, \?\)`).
		WithArgs(1, 11).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Renters: getOrCreateRenter
	// Renter - not found → insert → new ID=20
	mock.ExpectQuery(`SELECT id FROM renters WHERE afm = \?`).
		WithArgs(111222333).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO renters .*`).
		WithArgs("Αλική", "Ναυτικού", sqlmock.AnyArg(), 111222333, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(20, 1))

	// Link renter
	mock.ExpectExec(`INSERT OR IGNORE INTO entries_renter \(entry_id, renter_id\) VALUES \(\?, \?\)`).
		WithArgs(1, 20).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Coordinates - two inserts
	mock.ExpectExec(`INSERT INTO coordinates \(entry_id, latitude, longitude\) VALUES \(\?, \?, \?\)`).
		WithArgs(1, 37.9838, 23.7275).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO coordinates \(entry_id, latitude, longitude\) VALUES \(\?, \?, \?\)`).
		WithArgs(1, 37.9840, 23.7280).
		WillReturnResult(sqlmock.NewResult(2, 1))

	// Commit
	mock.ExpectCommit()

	// Call the function under test
	err = saveEntry(db, entry)
	if err != nil {
		t.Fatalf("saveEntry returned unexpected error: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetOrCreateOwner_ExistingOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	owner := OwnerDetails{
		FirstName:   "John",
		LastName:    "Doe",
		FathersName: "Jim",
		AFM:         123456789,
	}

	// Expect SELECT to find existing owner
	mock.ExpectQuery(`SELECT id FROM ownerDetails WHERE firstName = \? AND lastName = \?`).
		WithArgs("John", "Doe").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	tx, _ := db.Begin() // sqlmock handles tx automatically

	ownerID, err := getOrCreateOwner(tx, owner)
	if err != nil {
		t.Fatalf("getOrCreateOwner returned unexpected error: %v", err)
	}
	if ownerID != 42 {
		t.Fatalf("expected ownerID 42, got %d", ownerID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetOrCreateOwner_NewOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	owner := OwnerDetails{
		FirstName:   "Jane",
		LastName:    "Smith",
		FathersName: "Jack",
		AFM:         987654321,
		ADT:         "AB123456",
		ATA:         111,
		E9:          []byte{1, 2, 3},
		Notes:       "New owner",
	}

	// Expect SELECT to return no rows
	mock.ExpectQuery(`SELECT id FROM ownerDetails WHERE firstName = \? AND lastName = \?`).
		WithArgs("Jane", "Smith").
		WillReturnError(sql.ErrNoRows)

	// Expect INSERT and return new ID = 15
	mock.ExpectExec(`INSERT INTO ownerDetails \(firstName, lastName, fathersName, afm, adt, ata, e9, notes\) VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\)`).
		WithArgs("Jane", "Smith", "Jack", 987654321, "AB123456", 111, []byte{1, 2, 3}, "New owner").
		WillReturnResult(sqlmock.NewResult(15, 1))

	tx, _ := db.Begin()

	ownerID, err := getOrCreateOwner(tx, owner)
	if err != nil {
		t.Fatalf("getOrCreateOwner returned unexpected error: %v", err)
	}
	if ownerID != 15 {
		t.Fatalf("expected ownerID 15, got %d", ownerID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetOrCreateRenter_ExistingRenter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	renter := RenterDetails{
		FirstName: "Alice",
		LastName:  "Brown",
		AFM:       111222333,
	}

	mock.ExpectQuery(`SELECT id FROM renterDetails WHERE firstName = \? AND lastName = \?`).
		WithArgs("Alice", "Brown").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	tx, _ := db.Begin()

	renterID, err := getOrCreateRenters(tx, renter)
	if err != nil {
		t.Fatalf("getOrCreateRenter returned unexpected error: %v", err)
	}
	if renterID != 7 {
		t.Fatalf("expected renterID 7, got %d", renterID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetOrCreateRenter_NewRenter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	renter := RenterDetails{
		FirstName:   "Bob",
		LastName:    "Wilson",
		FathersName: "Bill",
		AFM:         555666777,
		ADT:         "XY987654",
		Notes:       "First time renter",
	}

	mock.ExpectQuery(`SELECT id FROM renterDetails WHERE firstName = \? AND lastName = \?`).
		WithArgs("Bob", "Wilson").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO renterDetails \(firstName, lastName, fathersName, afm, adt, notes\) VALUES \(\?, \?, \?, \?, \?, \?\)`).
		WithArgs("Bob", "Wilson", "Bill", 555666777, "XY987654", "First time renter").
		WillReturnResult(sqlmock.NewResult(23, 1))

	tx, _ := db.Begin()

	renterID, err := getOrCreateRenters(tx, renter)
	if err != nil {
		t.Fatalf("getOrCreateRenter returned unexpected error: %v", err)
	}
	if renterID != 23 {
		t.Fatalf("expected renterID 23, got %d", renterID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
