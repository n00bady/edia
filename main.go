package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize Fyne app
	myApp := app.NewWithID("xyz.n00bady.edia")
	myWindow := myApp.NewWindow("EDIA")

	// Get app storage path (Needs to be combatible with android 8 at least)
	dataDir := myApp.Storage().RootURI().Path()
	dbPath := filepath.Join(dataDir, "entries.db")
	log.Printf("Database path: %s", dbPath)

	// Initialize database
	db, err := initDB(dbPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return
	}
	defer db.Close()

	// Place holder same as Entry struct
	landlord_name := widget.NewEntry()
	landlord_name.SetPlaceHolder("Εκμισθωτής")

	renter_name := widget.NewEntry()
	renter_name.SetPlaceHolder("Μισθωτής")

	// I hate this
	lats := make([]*widget.Entry, 4)
	longs := make([]*widget.Entry, 4)
	for i := range 4 {
		lats[i] = widget.NewEntry()
		lats[i].SetPlaceHolder(fmt.Sprintf("Latitude %d", i+1))
		longs[i] = widget.NewEntry()
		longs[i].SetPlaceHolder(fmt.Sprintf("Longitude %d", i+1))
	}

	acres := widget.NewEntry()
	acres.SetPlaceHolder("Στρέμματα")

	t := widget.NewEntry()
	t.SetPlaceHolder("ΕΙΔΟΣ ΚΑΛ/ΓΕΙΑΣ")

	start_in := widget.NewLabel("Starting Date")
	start_in.Alignment = fyne.TextAlignCenter
	start_l := widget.NewLabel("")
	start_l.Alignment = fyne.TextAlignCenter
	start_d := &Date{dateChosen: start_l}
	// start_l := time.Now().Format("02/01/2006")
	// start_d := &date{dateChosen: start_l}
	startDate := xwidget.NewCalendar(time.Now(), start_d.onSelected)

	end_in := widget.NewLabel("End Date")
	end_in.Alignment = fyne.TextAlignCenter
	end_l := widget.NewLabel("")
	end_l.Alignment = fyne.TextAlignCenter
	end_d := &Date{dateChosen: start_l}
	// end_l := time.Now().Format("02/01/2006")
	// end_d := &date{dateChosen: end_l}
	endDate := xwidget.NewCalendar(time.Now(), end_d.onSelected)

	r := widget.NewEntry()
	r.SetPlaceHolder("Μίσθωμα")

	// entry := Entry{}

	// Save button
	saveBtn := widget.NewButton("Save", func() {
		// convert to float64 and gather the coordinates
		coords := make([]Coordinate, 0, 4)
		for i := range 4 {
			latValue, errLat := strconv.ParseFloat(lats[i].Text, 64)
			longValue, errLon := strconv.ParseFloat(longs[i].Text, 64)

			if errLat != nil || errLon != nil {
				log.Printf("Error parsing coordinates")
			}

			coords = append(coords, Coordinate{Latitude: latValue, Longitude: longValue})
		}
		// and the size
		size, err := strconv.ParseFloat(acres.Text, 64)
		if err != nil {
			log.Printf("Error parsing the size of the land")
		}

		money, err := strconv.ParseFloat(r.Text, 32)

		if err != nil {
			log.Printf("Error parsing the rent price")
		}
		newEntry := Entry{
			LandlordName: landlord_name.Text,
			RenterName:   renter_name.Text,
			Coords:       coords,
			Timestamp:    time.Now(),
			Size:         size,
			Type:         t.Text,
			Start:        start_d.dateChosen.Text,
			End:          end_d.dateChosen.Text,
			Rent:         money,
		}
		// entry = newEntry
		err = saveEntry(db, newEntry)
		if err != nil {
			log.Printf("Error saving entry: %v", err)
			return
		}
		// entry.SetText("")
		log.Printf("Saved entry: %s\n%s\n%f\netc...\n", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent)
	})

	title := widget.NewLabel("EDIA")
	title.Alignment = fyne.TextAlignCenter
	coords_l := widget.NewLabel("GEO Coordinates")
	separator := widget.NewSeparator()
	duration := widget.NewLabel("ΔΙΑΡΚΕΙΑ")
	// Layout
	left_container := container.NewVBox(
		landlord_name,
		renter_name,
		coords_l,
		lats[0],
		longs[0],
		separator,
		lats[1],
		longs[1],
		separator,
		lats[2],
		longs[2],
		separator,
		lats[3],
		longs[3],
		separator,
	)
	left_container.Resize(fyne.NewSize(300, 300))
	right_container := container.NewVBox(
		acres,
		t,
		r,
		duration,
		start_in,
		startDate,
		end_in,
		endDate,
		separator,
		saveBtn,
	)
	right_container.Resize(fyne.NewSize(300,300))

	content := container.NewHBox(
		left_container,
		right_container,
	)

	body := container.NewVBox(
		title,
		content,
	)

	// Set window content and size
	myWindow.SetContent(body)
	myWindow.Resize(fyne.NewSize(300, 600))

	// Runing the app
	myWindow.ShowAndRun()
}

func (d *Date) onSelected(t time.Time) {
	d.dateChosen.SetText(t.Format("02/01/2006"))
}
