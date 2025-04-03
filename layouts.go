package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
)

func mobileLayout(appState *AppState) (fyne.CanvasObject, error) {
	soon := widget.NewLabel("Soon(tm)")

	body := container.NewVBox(soon)

	return body, nil
}

func desktopLayout(appState *AppState) (fyne.CanvasObject, error) {
	landlord_name := widget.NewEntry()
	landlord_name.SetPlaceHolder("Εκμισθωτής")

	renter_name := widget.NewEntry()
	renter_name.SetPlaceHolder("Μισθωτής")

	// I hate this
	lats := make([]*widget.Entry, 4)
	longs := make([]*widget.Entry, 4)
	for i := range 4 {
		lats[i] = widget.NewEntry()
		lats[i].SetPlaceHolder(fmt.Sprintf("Πλάτος %d", i+1))
		longs[i] = widget.NewEntry()
		longs[i].SetPlaceHolder(fmt.Sprintf("Μήκος %d", i+1))
	}

	acres := widget.NewEntry()
	acres.SetPlaceHolder("Στρέμματα")

	t := widget.NewEntry()
	t.SetPlaceHolder("Ειδος Καλ/γειας")

	// Starting date input and it's button that opens a calendar for easier date choosing
	start_input := widget.NewEntry()
	start_input.SetPlaceHolder("ΑΠΟ")
	startDateButton := widget.NewButton("Pick a date", func() {
		showCalendar(start_input, appState.window)
	})

	// Same as starting date but for the ending date
	end_input := widget.NewEntry()
	end_input.SetPlaceHolder("ΕΩΣ")
	endDateButton := widget.NewButton("Pick a date", func() {
		showCalendar(end_input, appState.window)
	})

	r := widget.NewEntry()
	r.SetPlaceHolder("Μίσθωμα")

	// Save button
	saveBtn := widget.NewButton("Αποθήκευση", func() {
		// Convert to float64 and gather the coordinates
		coords := make([]Coordinate, 0, 4)
		for i := range 4 {
			latValue, errLat := strconv.ParseFloat(lats[i].Text, 64)
			longValue, errLon := strconv.ParseFloat(longs[i].Text, 64)

			if errLat != nil || errLon != nil {
				log.Printf("Error parsing coordinates")
				dialog.ShowError(errLat, appState.window)
			}

			coords = append(coords, Coordinate{Latitude: latValue, Longitude: longValue})
		}
		// and the size
		size, err := strconv.ParseFloat(acres.Text, 64)
		if err != nil {
			log.Printf("Error parsing the size of the land")
			dialog.ShowError(err, appState.window)
		}

		money, err := strconv.ParseFloat(r.Text, 32)
		if err != nil {
			log.Printf("Error parsing the rent price")
			dialog.ShowError(err, appState.window)
		}

		// We build the new entry here
		newEntry := Entry{
			LandlordName: landlord_name.Text,
			RenterName:   renter_name.Text,
			Coords:       coords,
			Timestamp:    time.Now(),
			Size:         size,
			Type:         t.Text,
			Start:        start_input.Text,
			End:          start_input.Text,
			Rent:         money,
		}

		err = saveEntry(appState.db, newEntry)
		if err != nil {
			log.Printf("Error saving entry: %v", err)
			dialog.ShowError(err, appState.window)
			return
		}

		log.Printf("Saved entry: %s\n%s\n%f\netc...\n", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent)
		dialog.ShowInformation("Database: ", fmt.Sprintf("Saved entry: %s\n%s\n%f\netc...\nin %s", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent, appState.db), appState.window)
	})

	title := widget.NewLabel("EDIA")
	title.Alignment = fyne.TextAlignCenter
	coords_l := widget.NewLabel("Γεωγραφικές Συντεταγμένες")
	separator := widget.NewSeparator()
	duration := widget.NewLabel("Διαρκεια")

	// Layout for the left container
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
	)

	// Layout for the right container
	right_container := container.NewVBox(
		acres,
		t,
		r,
		duration,
		start_input,
		startDateButton,
		end_input,
		endDateButton,
		separator,
	)

	// Putting both left and right containters on a grid
	content := container.NewGridWithColumns(2, left_container, right_container)

	// Finally add everything into a VBox and call it a day
	body := container.NewVBox(
		title,
		content,
		saveBtn,
	)

	return body, nil
}
