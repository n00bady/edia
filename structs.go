package main

import (
	"time"
)

// For the calendar date selection
// type Date struct {
// 	dateChosen *widget.Label
// }

// Coordinates for the land
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

// An entry is a placeholder for the actual structure that
// will hold all the data about each Rent and Contracts for
// farm land
type Entry struct {
	ID           int
	Timestamp    time.Time
	RenterName   string
	LandlordName string
	Coords       []Coordinate
	Size         float64
	Type         string
	Start        string
	End          string
	Rent         float64
}

