package main

import (
	"database/sql"
	"time"

	"fyne.io/fyne/v2"
	_ "github.com/mattn/go-sqlite3"
)

// It's easier that way
type AppState struct {
	db     *sql.DB
	app    fyne.App
	window fyne.Window
}

// Main struct/table
type Entry struct {
	ID        uint
	Name      string
	Timestamp time.Time
	Renters   []RenterDetails
	Owners    []OwnerDetails
	Coords    []Coordinates
	ATAK      uint
	KAEK      string
	Size      float64
	Type      string
	Start     string
	End       string
	Rent      float64
}

// Coordinates for the land
type Coordinates struct {
	ID        uint
	EntryID   uint
	Latitude  float64
	Longitude float64
}

// Εκμισθωτές
type OwnerDetails struct {
	ID          uint
	FirstName   string
	LastName    string
	FathersName string
	AFM         uint
	ADT         string
	E9          []byte
	Notes       string
}

// Μισθωτές
type RenterDetails struct {
	ID          uint
	FirstName   string
	LastName    string
	FathersName string
	AFM         uint
	ADT         string
	E9          []byte
	Notes       string
}

// Junction tables
type EntryOwner struct {
	EntryID uint
	OwnerID uint
}

type EntryRenter struct {
	EntryID  uint
	RenterID uint
}
