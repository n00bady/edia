package main

import (
    "errors"
    "fmt"
    "regexp"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestUpdateEntry(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("error creating mock database: %v", err)
    }
    defer db.Close()

    // Sample entry for testing
    entry := Entry{
        ID:           1,
        LandlordName: "Γιαννάκης Ελάφις",
        RenterName:   "Γιάννα Σμιθερίδη",
        Size:         100.5,
        Type:         "Ζαρζαβατικά",
        Rent:         1500.0,
        Start:        "2028-01-01",
        End:          "2025-12-31",
        Coords: []Coordinate{
            {Latitude: 40.7128, Longitude: -74.0060},
            {Latitude: 40.7129, Longitude: -74.0061},
        },
    }

    t.Run("Successful Update", func(t *testing.T) {
        mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM coordinates WHERE entry_id = ?")).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

        mock.ExpectBegin()

        mock.ExpectExec(regexp.QuoteMeta("UPDATE entries SET LandLord = ?, Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?")).
            WithArgs(entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID).
            WillReturnResult(sqlmock.NewResult(0, 1))

        for i, coord := range entry.Coords {
            mock.ExpectExec(regexp.QuoteMeta("UPDATE coordinates SET Latitude = ?, Longitude = ? WHERE entry_id = ? AND id = ?")).
                WithArgs(coord.Latitude, coord.Longitude, entry.ID, i+1).
                WillReturnResult(sqlmock.NewResult(0, 1))
        }

        mock.ExpectCommit()

        err := updateEntry(db, entry)
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Invalid Coordinate Count", func(t *testing.T) {
        invalidEntry := entry
        invalidEntry.Coords = make([]Coordinate, 5)

        mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM coordinates WHERE entry_id = ?")).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

        mock.ExpectBegin()

        mock.ExpectExec(regexp.QuoteMeta("UPDATE entries SET LandLord = ?, Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?")).
            WithArgs(invalidEntry.LandlordName, invalidEntry.RenterName, invalidEntry.Size, invalidEntry.Type, invalidEntry.Rent, invalidEntry.Start, invalidEntry.End, invalidEntry.ID).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectRollback()

        err := updateEntry(db, invalidEntry)
        if err == nil {
            t.Error("expected error for invalid coordinate count, got none")
        }
        expectedErr := fmt.Sprintf("expected up to 4 pair of coordinates got %d, for entry ID: %d", len(invalidEntry.Coords), invalidEntry.ID)
        if err.Error() != expectedErr {
            t.Errorf("expected error %q, got %q", expectedErr, err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Query Coordinate IDs Error", func(t *testing.T) {
        mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM coordinates WHERE entry_id = ?")).
            WithArgs(1).
            WillReturnError(errors.New("database error"))

        err := updateEntry(db, entry)
        if err == nil {
            t.Error("expected error for query failure, got none")
        }
        expectedErr := "could not query the database for the coordinates IDs: database error"
        if err.Error() != expectedErr {
            t.Errorf("expected error %q, got %q", expectedErr, err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Update Entry Error", func(t *testing.T) {
        mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM coordinates WHERE entry_id = ?")).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

        mock.ExpectBegin()

        mock.ExpectExec(regexp.QuoteMeta("UPDATE entries SET LandLord = ?, Renter = ?, Size = ?, Type = ?, Rent = ?, Start = ?, End = ? WHERE id = ?")).
            WithArgs(entry.LandlordName, entry.RenterName, entry.Size, entry.Type, entry.Rent, entry.Start, entry.End, entry.ID).
            WillReturnError(errors.New("update error"))

        mock.ExpectRollback()

        err := updateEntry(db, entry)
        if err == nil {
            t.Error("expected error for update failure, got none")
        }
        expectedErr := fmt.Sprintf("error updating entry with id: %d: update error", entry.ID)
        if err.Error() != expectedErr {
            t.Errorf("expected error %q, got %q", expectedErr, err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })
}

func TestDelEntry(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("error creating mock database: %v", err)
    }
    defer db.Close()

    t.Run("Successful Deletion", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta("Delete From coordinates Where entry_id = ?")).
            WithArgs(1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("Delete From entries Where id = ?")).
            WithArgs(1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM entries) WHERE name = 'entries'")).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM coordinates) WHERE name = 'coordinates'")).
            WillReturnResult(sqlmock.NewResult(0, 1))

        err := delEntry(db, 1)
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Delete Coordinates Error", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta("Delete From coordinates Where entry_id = ?")).
            WithArgs(1).
            WillReturnError(errors.New("delete coordinates error"))

        err := delEntry(db, 1)
        if err == nil {
            t.Error("expected error for delete coordinates failure, got none")
        }
        expectedErr := "could not delete coordinates for entry 1"
        if err.Error() != expectedErr {
            t.Errorf("expected error %q, got %q", expectedErr, err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Delete Entry Error", func(t *testing.T) {
        // Mock the deletion of coordinates
        mock.ExpectExec(regexp.QuoteMeta("Delete From coordinates Where entry_id = ?")).
            WithArgs(1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("Delete From entries Where id = ?")).
            WithArgs(1).
            WillReturnError(errors.New("delete entry error"))

        err := delEntry(db, 1)
        if err == nil {
            t.Error("expected error for delete entry failure, got none")
        }
        expectedErr := "could not delete entry: 1"
        if err.Error() != expectedErr {
            t.Errorf("expected error %q, got %q", expectedErr, err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })

    t.Run("Reset Autoincrement Error", func(t *testing.T) {
        mock.ExpectExec(regexp.QuoteMeta("Delete From coordinates Where entry_id = ?")).
            WithArgs(1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("Delete From entries Where id = ?")).
            WithArgs(1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        mock.ExpectExec(regexp.QuoteMeta("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM entries) WHERE name = 'entries'")).
            WillReturnError(errors.New("reset autoincrement error"))

        err := delEntry(db, 1)
        if err == nil {
            t.Error("expected error for reset autoincrement failure, got none")
        }
        if err.Error() != "reset autoincrement error" {
            t.Errorf("expected error %q, got %q", "reset autoincrement error", err.Error())
        }

        if err := mock.ExpectationsWereMet(); err != nil {
            t.Errorf("unfulfilled expectations: %s", err)
        }
    })
}
