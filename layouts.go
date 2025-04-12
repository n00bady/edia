package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// This is just the desktopForm copy-pasted but changed the layout and added additional
// functions specific for mobile use.
func mobileForm(appState *AppState) (fyne.CanvasObject, error) {
	log.Printf("Creating the mobileForm...")

	title := widget.NewLabel("EDIA")
	title.Alignment = fyne.TextAlignCenter
	coords_l := widget.NewLabel("Γεωγραφικές Συντεταγμένες")
	separator := widget.NewSeparator()
	duration := widget.NewLabel("Διαρκεια")

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
	coords_inputs := container.New(
		layout.NewGridLayoutWithRows(4),
		lats[0], longs[0],
		lats[1], longs[1],
		lats[2], longs[2],
		lats[3], longs[3],
	)

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
		// and the sizekj
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
		dialog.ShowInformation("Database: ", fmt.Sprintf("Saved entry: %s\n%s\n%f\netc...\n", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent), appState.window)
	})

	// Layout for the left container
	left_container := container.NewVBox(
		landlord_name,
		renter_name,
		separator,
		separator,
		coords_l,
		coords_inputs,
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

	backButton := widget.NewButton("Go back", func() {
		tmp, err := mainView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	body := container.NewVScroll(
		container.NewVBox(
			left_container,
			layout.NewSpacer(),
			right_container,
			container.New(
				layout.NewHBoxLayout(),
				layout.NewSpacer(),
				backButton,
				saveBtn,
			),
		),
	)

	allInputs := []*widget.Entry{
		landlord_name,
		renter_name,
		lats[0], longs[0],
		lats[1], longs[1],
		lats[2], longs[2],
		lats[3], longs[3],
		acres,
		t,
		r,
	}

	focusChain(allInputs, appState)

	log.Printf("mobileForm created successfully.")

	return body, nil
}

func desktopForm(appState *AppState) (fyne.CanvasObject, error) {
	log.Printf("Creating desktopForm...")
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
		dialog.ShowInformation("Database: ", fmt.Sprintf("Saved entry: %s\n%s\n%f\netc...\n", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent), appState.window)
	})

	title := widget.NewLabel("EDIA")
	title.Alignment = fyne.TextAlignCenter
	coords_l := widget.NewLabel("Γεωγραφικές Συντεταγμένες")
	// separator := widget.NewSeparator()
	duration := widget.NewLabel("Διαρκεια")

	// Layout for the left container
	left_container := container.NewVBox(
		landlord_name,
		renter_name,
		coords_l,
		lats[0],
		longs[0],
		lats[1],
		longs[1],
		lats[2],
		longs[2],
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
	)

	backButton := widget.NewButton("Cancel", func() {
		tmp, err := mainView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	// Putting both left and right containters on a grid
	content := container.NewGridWithColumns(2, left_container, right_container)
	buttons := container.NewGridWithColumns(2, backButton, saveBtn)

	// Finally add everything into a VBox and call it a day
	body := container.NewVBox(
		title,
		content,
		layout.NewSpacer(),
		buttons,
	)

	log.Printf("desktopForm created successfully.")

	return body, nil
}

func mainView(appState *AppState) (fyne.CanvasObject, error) {
	log.Printf("Creating the mainView...")
	entries, err := getAll(appState.db)
	if err != nil {
		return nil, err
	}

	list := widget.NewList(
		func() int {
			return len(entries)
		},
		func() fyne.CanvasObject {
			log.Printf("Creating list template...")
			return widget.NewLabel("Template")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			log.Printf("Updating item with ID: %d", lii)
			if lii < 0 || lii >= len(entries) {
				log.Printf("Invalid item ID: %d", lii)
				return
			}
			entry := entries[lii]
			label, ok := co.(*widget.Label)
			if !ok {
				log.Printf("Canvas object is not *widget.Label, its: %s", fmt.Sprintf("%T", co))
				return
			}
			label.SetText(fmt.Sprintf("%d: %s -> %s", entry.ID, entry.LandlordName, entry.RenterName))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d", id)
		if id >= 0 && id < len(entries) {
			log.Printf("Showing popup for item: %d", entries[id].ID)
			showDetailsPopup(entries[id], appState)
			list.UnselectAll()
		}
	}

	var addButton fyne.CanvasObject

	if fyne.CurrentDevice().IsMobile() {
		addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			tmp, err := mobileForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})
	} else {
		addButton = widget.NewButtonWithIcon("Add New Entry", theme.ContentAddIcon(), func() {
			tmp, err := desktopForm(appState)
			if err != nil {
				log.Printf("error constructing desktop layout: %v", err)
			}
			body := container.NewBorder(nil, nil, nil, nil, tmp)
			appState.window.SetContent(body)
		})
	}
	addButton.Resize(fyne.NewSize(200, 200))

	body := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		container.NewVScroll(list),
		container.New(
			layout.NewVBoxLayout(),
			layout.NewSpacer(),
			container.New(
				layout.NewHBoxLayout(),
				layout.NewSpacer(),
				container.NewPadded(addButton),
			),
		),
	)

	log.Printf("mainView created successfully!")

	return body, nil
}

func focusChain(inputs []*widget.Entry, appState *AppState) {
	for i, input := range inputs {
		input.OnSubmitted = func(_ string) {
			if i < len(inputs)-1 {
				appState.window.Canvas().Focus(inputs[i+1])
			} else {
				input.Disable()
				input.Enable()
			}
		}
	}
}

// Details popup for the list
func showDetailsPopup(entry Entry, appState *AppState) {
	log.Printf("Showing popup for: %d", entry.ID)
	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("ID: %d", entry.ID)),
		widget.NewLabel(fmt.Sprintf("Timestamp: %s", entry.Timestamp)),
		widget.NewLabel(fmt.Sprintf("Landlord: %s", entry.LandlordName)),
		widget.NewLabel(fmt.Sprintf("Renter: %s", entry.RenterName)),
		widget.NewLabel(fmt.Sprintf("Rent: %f€", entry.Rent)),
		widget.NewLabel(fmt.Sprintf("From: %s", entry.Start)),
		widget.NewLabel(fmt.Sprintf("To: %s", entry.End)),
		widget.NewButton("Close", nil),
	)

	// Iterate over all canvas objects of content 
	// and if its a label enable text wrapping
	for _, obj := range content.Objects {
		if label, ok := obj.(*widget.Label); ok {
			label.Wrapping = fyne.TextWrapWord
		}
	}

	popup := widget.NewModalPopUp(content, appState.window.Canvas())

	// Popup close button
	content.Objects[len(content.Objects)-1].(*widget.Button).OnTapped = func() {
		popup.Hide()
	}

	contentMinHeight := content.MinSize().Height
	popup.Resize(fyne.NewSize(320, contentMinHeight))

	popup.Show()
	fmt.Printf("Popup displayed for: %d", entry.ID)
}
