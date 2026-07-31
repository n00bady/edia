package main

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetOwners_ReturnsExpectedOwners(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	cols := []string{"id", "firstName", "lastName", "fathersName", "afm", "adt", "e9", "homeAddress", "phoneNumber", "email", "accountantInfo", "notes"}
	mockRows := sqlmock.NewRows(cols).AddRow(1, "John", "Doe", "Jr", 12345, "ADT", []byte("e9"), "Home", "555", "john@doe", "acc", "notes")

	query := `
		SELECT o.id, o.firstName, o.lastName, o.fathersName, o.afm, o.adt, o.e9, o.homeAddress, o.phoneNumber, o.email, o.accountantInfo, o.notes
		FROM ownerDetails o
		JOIN entries_owner eo ON o.id = eo.owner_id
		WHERE eo.entry_id = ?`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).WillReturnRows(mockRows)

	got, err := getOwners(db, Entry{ID: 1})
	if err != nil {
		t.Fatalf("getOwners returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 owner, got %d", len(got))
	}
	if got[0].FirstName != "John" || got[0].LastName != "Doe" {
		t.Fatalf("unexpected owner returned: %+v", got[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetCoords_ReturnsExpectedCoordinates(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	cols := []string{"id", "entry_id", "latitude", "longitude"}
	mockRows := sqlmock.NewRows(cols).AddRow(10, 1, 37.1234, 23.4567)

	query := `
		SELECT id, entry_id, latitude, longitude
		FROM coordinates
		WHERE entry_id = ?`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).WillReturnRows(mockRows)

	got, err := getCoords(db, Entry{ID: 1})
	if err != nil {
		t.Fatalf("getCoords returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 coordinate, got %d", len(got))
	}
	if got[0].Latitude != 37.1234 || got[0].Longitude != 23.4567 {
		t.Fatalf("unexpected coordinate returned: %+v", got[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetYearRange_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	// sqlmock expects a regex, escape the query
	rows := sqlmock.NewRows([]string{"oldest", "newest"}).AddRow(int64(1990), int64(2026))
	mock.ExpectQuery(`MIN\(year\).*MAX\(year\)`).WillReturnRows(rows)

	oldest, newest, err := getYearRange(db)
	if err != nil {
		t.Fatalf("getYearRange returned error: %v", err)
	}
	if oldest != 1990 || newest != 2026 {
		t.Fatalf("unexpected year range: %d-%d", oldest, newest)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestDelEntry_InvalidIDReturnsError(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	if err := delEntry(db, 0); err == nil {
		t.Fatalf("expected error for invalid id, got nil")
	}
}

func TestDelEntry_DeleteFlow(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	// SELECT EXISTS(...) -> return true
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM entries WHERE id = ?)")).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// DELETE FROM entries WHERE id = ?
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM entries WHERE id = ?")).WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := delEntry(db, 1); err != nil {
		t.Fatalf("delEntry returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestGetAllOwnersAndRenters_ScanMapping(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating sqlmock: %v", err)
	}
	defer db.Close()

	ownerCols := []string{"id", "firstName", "lastName", "fathersName", "afm", "adt", "e9", "homeAddress", "phoneNumber", "email", "accountantInfo", "notes"}
	mockOwners := sqlmock.NewRows(ownerCols).AddRow(2, "Alice", "Smith", "", 0, "", []byte{}, "", "", "alice@example.com", "", "")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM ownerDetails")).WillReturnRows(mockOwners)

	owners, err := getAllOwners(db)
	if err != nil {
		t.Fatalf("getAllOwners returned error: %v", err)
	}
	if len(owners) != 1 || owners[0].FirstName != "Alice" {
		t.Fatalf("unexpected owners result: %+v", owners)
	}

	renterCols := []string{"id", "firstName", "lastName", "fathersName", "afm", "adt", "e9", "notes"}
	mockRenters := sqlmock.NewRows(renterCols).AddRow(3, "Bob", "Jones", "", 0, "", []byte{}, "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM renterDetails")).WillReturnRows(mockRenters)

	renters, err := getAllRenters(db)
	if err != nil {
		t.Fatalf("getAllRenters returned error: %v", err)
	}
	if len(renters) != 1 || renters[0].FirstName != "Bob" {
		t.Fatalf("unexpected renters result: %+v", renters)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
