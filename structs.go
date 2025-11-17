package main

import (
	"database/sql"
	"time"

	"fyne.io/fyne/v2"
	_ "github.com/mattn/go-sqlite3"
)

// For the calendar date selection
// type Date struct {
// 	dateChosen *widget.Label
// }

// It's easier that way
type AppState struct {
	db     *sql.DB
	window fyne.Window
}

// Coordinates for the land
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

// Εκμισθωτές
type LandlordDetails struct {
	Entry_IDs   []int
	Entry_total int
	FirstName   string
	LastName    string
	FathersName string
	AFM         int
	ADT         string
	ATA         int
	E9          []byte
	Notes       string
}

// An entry is a placeholder for the actual structure that
// will hold all the data about each Rent and Contracts for
// farm land
type Entry struct {
	ID           int
	NickName     string
	Timestamp    time.Time
	RenterName   string
	LandlordName []LandlordDetails
	Coords       []Coordinate
	Size         float64
	Type         string
	Start        string
	End          string
	Rent         float64
}
