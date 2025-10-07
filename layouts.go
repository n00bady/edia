package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

func newEntryWithLabel(ph string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(ph)

	return entry
}

func AddForm(appState *AppState) (fyne.CanvasObject, error) {
	entriesMap := make(map[string]*widget.Entry)

	coordsLabel := widget.NewLabel("ΓεωΣυντεταγμένες")
	durationLabel := widget.NewLabel("Διαρκεια")

	labelsEntries := []string{
		"Όνομα Εγγραφής",
		"Εκμισθωτής",
		"Μισθωτής",
		"Στρέμματα",
		"Είδος Καλ/γειας",
		"Μίσθωμα",
	}

	for _, l := range labelsEntries {
		tmpEnt := newEntryWithLabel(l)
		entriesMap[l] = tmpEnt
	}

	for i := range 4 {
		lat := newEntryWithLabel(fmt.Sprintf("Πλάτος %d", i+1))
		entriesMap[fmt.Sprintf("Πλάτος %d", i+1)] = lat
		long := newEntryWithLabel(fmt.Sprintf("Μήκος %d", i+1))
		entriesMap[fmt.Sprintf("Μήκος %d", i+1)] = long
	}

	// Starting date input and it's button that opens a calendar for easier date choosing
	start_input := widget.NewEntry()
	start_input.SetPlaceHolder("ΑΠΟ")
	startDateButton := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		showCalendar(start_input, appState.window)
	})
	startDateInput := container.NewBorder(nil, nil, nil, startDateButton, start_input)

	// Same as starting date but for the ending date
	end_input := widget.NewEntry()
	end_input.SetPlaceHolder("ΕΩΣ")
	endDateButton := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		showCalendar(end_input, appState.window)
	})
	endDateInput := container.NewBorder(nil, nil, nil, endDateButton, end_input)

	// Save button
	saveBtn := widget.NewButton("Αποθήκευση", func() {
		// Convert to float64 and gather the coordinates
		coords := make([]Coordinate, 0, 4)
		for i := range 4 {
			latValue, errLat := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Πλάτος %d", i+1)].Text, 5)
			longValue, errLon := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Μήκος %d", i+1)].Text, 5)

			if errLat != nil || errLon != nil {
				log.Printf("Error parsing coordinates")
				dialog.ShowError(errLat, appState.window)
			}

			coords = append(coords, Coordinate{Latitude: latValue, Longitude: longValue})
		}
		// and the size
		size, err := strconv.ParseFloat(entriesMap["Στρέμματα"].Text, 64)
		if err != nil {
			log.Printf("Error parsing the size of the land")
			dialog.ShowError(err, appState.window)
		}

		money, err := strconv.ParseFloat(entriesMap["Μίσθωμα"].Text, 32)
		if err != nil {
			log.Printf("Error parsing the rent price")
			dialog.ShowError(err, appState.window)
		}
		money = TruncateFloatTo2Decimals(money)

		// We build the new entry here
		newEntry := Entry{
			NickName:     entriesMap["Όνομα Εγγραφής"].Text,
			LandlordName: entriesMap["Εκμισθωτής"].Text,
			RenterName:   entriesMap["Μισθωτής"].Text,
			Coords:       coords,
			Timestamp:    time.Now(),
			Size:         size,
			Type:         entriesMap["Είδος Καλ/γειας"].Text,
			Start:        start_input.Text,
			End:          end_input.Text,
			Rent:         money,
		}

		err = saveEntry(appState.db, newEntry)
		if err != nil {
			log.Printf("Error saving entry: %v", err)
			dialog.ShowError(err, appState.window)
			return
		}

		log.Printf("Saved entry: %s\n%s\n%f\netc...\n", newEntry.LandlordName, newEntry.RenterName, newEntry.Rent)
		dialog.ShowInformation("Database:", "Saved successfully!", appState.window)

		// return to mainView
		mainview, err := listView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
			dialog.ShowError(err, appState.window)
		}
		appState.window.SetContent(mainview)
	})

	// Cancel button to go back
	backButton := widget.NewButton("Cancel", func() {
		tmp, err := listView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	if fyne.CurrentDevice().IsMobile() {
		// --Mobile layout--

		leftContainer := container.NewVBox(
			entriesMap["Όνομα Εγγραφής"],
			entriesMap["Εκμισθωτής"],
			entriesMap["Μισθωτής"],
			layout.NewSpacer(),
			coordsLabel,
		)
		for i := range 4 {
			coordContainer := container.NewGridWithColumns(2, entriesMap[fmt.Sprintf("Πλάτος %d", i+1)], entriesMap[fmt.Sprintf("Μήκος %d", i+1)])
			leftContainer.Add(coordContainer)
		}

		rightContainer := container.NewVBox(
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			layout.NewSpacer(),
			durationLabel,
			startDateInput,
			endDateInput,
		)

		content := container.NewGridWithColumns(2, leftContainer, rightContainer)
		buttons := container.NewGridWithColumns(2, backButton, saveBtn)

		body := container.NewVBox(
			content,
			layout.NewSpacer(),
			buttons,
		)

		allInputs := []fyne.CanvasObject{
			entriesMap["Όνομα Εγγραφής"],
			entriesMap["Εκμισθωτής"],
			entriesMap["Μισθωτής"],
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			entriesMap["Πλάτος 1"], entriesMap["Μήκος 1"],
			entriesMap["Πλάτος 2"], entriesMap["Μήκος 2"],
			entriesMap["Πλάτος 3"], entriesMap["Μήκος 3"],
			entriesMap["Πλάτος 4"], entriesMap["Μήκος 4"],
		}

		// Unfocuses to prevent tapping every single entry field when draging
		// body.OnScrolled = func(p fyne.Position) {
		// 	appState.window.Canvas().Unfocus()
		// }
		// TODO: Figure out an easy way to be able to scroll when you tap and drag
		// on an entry field.

		focusChain(allInputs, appState, body)

		log.Printf("mobileForm created successfully.")

		return body, nil

	} else {
		// --Desktop layout--

		// LEFT
		left_container := container.NewVBox(
			entriesMap["Όνομα Εγγραφής"],
			entriesMap["Εκμισθωτής"],
			entriesMap["Μισθωτής"],
			layout.NewSpacer(),
			coordsLabel,
		)
		for i := range 4 {
			coordContainer := container.NewGridWithColumns(2, entriesMap[fmt.Sprintf("Πλάτος %d", i+1)], entriesMap[fmt.Sprintf("Μήκος %d", i+1)])
			left_container.Add(coordContainer)
		}

		// RIGHT
		right_container := container.NewVBox(
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			layout.NewSpacer(),
			durationLabel,
			startDateInput,
			endDateInput,
		)

		// Putting both left and right containters on a grid
		content := container.NewGridWithColumns(2, left_container, right_container)
		buttons := container.NewGridWithColumns(2, backButton, saveBtn)

		// Finally add everything into a VBox and call it a day
		body := container.NewVBox(
			content,
			layout.NewSpacer(),
			buttons,
		)

		log.Printf("desktopForm created successfully.")
		
		return body, nil
	}
}

func editForm(appState *AppState, id int) (fyne.CanvasObject, error) {
	log.Printf("Creating desktop edit form...")

	entriesMap := make(map[string]*widget.Entry)

	selectedEntry, err := getEntry(appState.db, id)
	if err != nil {
		return nil, err
	}

	coordsLabel := widget.NewLabel("ΓεωΣυντεταγμένες")
	durationLabel := widget.NewLabel("Διαρκεια")

	labelsEntries := []string{
		"Εκμισθωτής",
		"Μισθωτής",
		"Στρέμματα",
		"Είδος Καλ/γειας",
		"Μίσθωμα",
	}

	for _, l := range labelsEntries {
		tmpEnt := newEntryWithLabel(l)
		entriesMap[l] = tmpEnt
	}

	// Assign values to entries from the selected Entry
	entriesMap["Εκμισθωτής"].SetText(selectedEntry.LandlordName)
	entriesMap["Μισθωτής"].SetText(selectedEntry.RenterName)
	entriesMap["Στρέμματα"].SetText(strconv.FormatFloat(selectedEntry.Size, 'f', -1, 64))
	entriesMap["Είδος Καλ/γειας"].SetText(selectedEntry.Type)
	entriesMap["Μίσθωμα"].SetText(strconv.FormatFloat(selectedEntry.Rent, 'f', -1, 64))

	for i := range 4 {
		lat := newEntryWithLabel(fmt.Sprintf("Πλάτος %d", i+1))
		lat.SetText(strconv.FormatFloat(selectedEntry.Coords[i].Latitude, 'f', -1, 64))
		entriesMap[fmt.Sprintf("Πλάτος %d", i+1)] = lat

		long := newEntryWithLabel(fmt.Sprintf("Μήκος %d", i+1))
		long.SetText(strconv.FormatFloat(selectedEntry.Coords[i].Longitude, 'f', -1, 64))
		entriesMap[fmt.Sprintf("Μήκος %d", i+1)] = long
	}

	// Starting date input and it's button that opens a calendar for easier date choosing
	start_input := widget.NewEntry()
	start_input.SetPlaceHolder("ΑΠΟ")
	start_input.SetText(selectedEntry.Start)
	startDateButton := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		showCalendar(start_input, appState.window)
	})
	startDateInput := container.NewBorder(nil, nil, nil, startDateButton, start_input)

	// Same as starting date but for the ending date
	end_input := widget.NewEntry()
	end_input.SetPlaceHolder("ΕΩΣ")
	end_input.SetText(selectedEntry.End)
	endDateButton := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		showCalendar(end_input, appState.window)
	})
	endDateInput := container.NewBorder(nil, nil, nil, endDateButton, end_input)

	saveBtn := widget.NewButton("Αποθήκευση", func() {
		// Convert to float64 and gather the coordinates
		coords := make([]Coordinate, 0, 4)
		for i := range 4 {
			latValue, err := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Πλάτος %d", i+1)].Text, 5)
			if err != nil {
				latValue = 0
				log.Printf("Error parsing coordinates: %v", err)
				dialog.ShowError(err, appState.window)
			}
			longValue, err := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Μήκος %d", i+1)].Text, 5)
			if err != nil {
				longValue = 0
				log.Printf("Error parsing coordinates: %v", err)
				dialog.ShowError(err, appState.window)
			}

			coords = append(coords, Coordinate{Latitude: latValue, Longitude: longValue})
		}
		// and the size
		size, err := strconv.ParseFloat(entriesMap["Στρέμματα"].Text, 64)
		if err != nil {
			log.Printf("Error parsing the size of the land")
			dialog.ShowError(err, appState.window)
		}

		money, err := strconv.ParseFloat(entriesMap["Μίσθωμα"].Text, 32)
		if err != nil {
			log.Printf("Error parsing the rent price")
			dialog.ShowError(err, appState.window)
		}
		money = TruncateFloatTo2Decimals(money)

		// We build the new entry here
		editedEntry := Entry{
			ID:           id,
			LandlordName: entriesMap["Εκμισθωτής"].Text,
			RenterName:   entriesMap["Μισθωτής"].Text,
			Coords:       coords,
			Timestamp:    time.Now(),
			Size:         size,
			Type:         entriesMap["Είδος Καλ/γειας"].Text,
			Start:        start_input.Text,
			End:          end_input.Text,
			Rent:         money,
		}

		err = updateEntry(appState.db, editedEntry)
		if err != nil {
			log.Printf("Error saving entry: %v", err)
			dialog.ShowError(err, appState.window)
			return
		}

		log.Printf("Saved entry: %d", editedEntry.ID)
		dialog.ShowInformation("Database:", "Saved successfully!", appState.window)
	})

	backButton := widget.NewButton("Cancel", func() {
		tmp, err := listView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	// LEFT
	left_container := container.NewVBox(
		entriesMap["Εκμισθωτής"],
		entriesMap["Μισθωτής"],
		layout.NewSpacer(),
		coordsLabel,
	)
	for i := range 4 {
		coordContainer := container.NewGridWithColumns(2, entriesMap[fmt.Sprintf("Πλάτος %d", i+1)], entriesMap[fmt.Sprintf("Μήκος %d", i+1)])
		left_container.Add(coordContainer)
	}

	// RIGHT
	right_container := container.NewVBox(
		entriesMap["Στρέμματα"],
		entriesMap["Είδος Καλ/γειας"],
		entriesMap["Μίσθωμα"],
		layout.NewSpacer(),
		durationLabel,
		startDateInput,
		endDateInput,
	)

	// Putting both left and right containters on a grid
	content := container.NewGridWithColumns(2, left_container, right_container)
	buttons := container.NewGridWithColumns(2, backButton, saveBtn)

	// Finally add everything into a VBox and call it a day
	body := container.NewVBox(
		content,
		layout.NewSpacer(),
		buttons,
	)

	return body, nil
}

func mainView(appState *AppState) (fyne.CanvasObject, error) {
	listViewButton := widget.NewButton("Χωράφια", func() {
		lView, err := listView(appState)
		if err != nil {
			log.Printf("error constructing listView: %v", err)
			dialog.ShowError(err, appState.window)
		}

		appState.window.SetContent(lView)
	})

	landLordButton := widget.NewButton("Ιδιοκτήτες", func(){})

	renterButton := widget.NewButton("Μισθωτές", func() {})

	body := container.New(layout.NewCenterLayout(), container.NewVBox(listViewButton, landLordButton, renterButton))

	return body, nil
}

func listView(appState *AppState) (fyne.CanvasObject, error) {
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
			if err != nil {
				dialog.ShowError(err, appState.window)
				os.Exit(1)
			}
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
			label.SetText(fmt.Sprintf("%d: %s", entry.ID, entry.NickName))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d", id)
		if id >= 0 && id < len(entries) {
			log.Printf("Showing popup for item: %d", entries[id].ID)
			showDetailsPopup(entries[id], appState, list, &entries)
			list.UnselectAll()
		}
	}

	var addButton fyne.CanvasObject

	if fyne.CurrentDevice().IsMobile() {
		addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			tmp, err := AddForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})
	} else {
		addButton = widget.NewButtonWithIcon("Add New Entry", theme.ContentAddIcon(), func() {
			tmp, err := AddForm(appState)
			if err != nil {
				log.Printf("error constructing desktop layout: %v", err)
			}
			appState.window.SetContent(tmp)
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

// if and when the xwidget.NumericalEntry works this will actually be usefull
func focusChain(inputs []fyne.CanvasObject, appState *AppState, scrollContainer *fyne.Container) {
	lastInput := inputs[len(inputs)-1]
	for i, input := range inputs {
		switch e := input.(type) {
		case *widget.Entry:
			e.OnSubmitted = func(_ string) {
				// if i < len(inputs)-1 {
				// 	appState.window.Canvas().Focus(inputs[i+1].(fyne.Focusable))
				//
				// 	// TODO: Fix the scrolling when focusing on the next entry
				// 	offSet := inputs[i+1].Position()
				// 	scrollContainer.ScrollToOffset(offSet)
				// } else {
				// 	fyne.CurrentDevice().(mobile.Device).HideVirtualKeyboard()
				// 	// e.Disable()
				// 	// e.Enable()
				// }
				if input == lastInput {
					fyne.CurrentDevice().(mobile.Device).HideVirtualKeyboard()
					appState.window.Canvas().Unfocus()
				} else {
					// appState.window.Canvas().FocusNext()
					appState.window.Canvas().Focus(inputs[i+1].(fyne.Focusable))
				}
			}
		case *xwidget.NumericalEntry:
			e.OnSubmitted = func(s string) {
				if i < len(inputs)-1 {
					appState.window.Canvas().Focus(inputs[i+1].(fyne.Focusable))
				} else {
					fyne.CurrentDevice().(mobile.Device).HideVirtualKeyboard()
					// e.Disable()
					// e.Enable()
				}
			}
		}
	}
}

// Details popup for the list
func showDetailsPopup(entry Entry, appState *AppState, list *widget.List, entries *[]Entry) {
	log.Printf("Showing popup for: %d", entry.ID)

	editButton := widget.NewButton("Edit", nil)
	closeButton := widget.NewButton("Close", nil)
	deleteButton := widget.NewButton("Delete", nil)

	buttonsContainer := container.NewVBox(
		closeButton,
		container.NewGridWithColumns(2, editButton, deleteButton),
	)

	// TODO: Did I forget about coordinates ???
	fmt.Printf("Coords: %d", len(entry.Coords))
	coordsContainer := container.NewVBox()
	// for i := range 4 {
	// 	lat := fmt.Sprintf("Πλάτος %d: %s", i, strconv.FormatFloat(entry.Coords[i].Latitude, 'f', -1, 64))
	// 	long := fmt.Sprintf("Μήκος %d: %s", i, strconv.FormatFloat(entry.Coords[i].Longitude, 'f', -1, 64))
	// 	coordContainer := container.NewGridWithColumns(2, widget.NewLabel(lat), widget.NewLabel(long))
	// 	coordContainer.Add(coordContainer)
	// }

	// Add all the details!
	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("ID: %d", entry.ID)),
		widget.NewLabel(fmt.Sprintf("Εκμισθωτής: %s", entry.LandlordName)),
		widget.NewLabel(fmt.Sprintf("Μισθωτής: %s", entry.RenterName)),
		widget.NewLabel(fmt.Sprintf("Μίσθωμα: %.2f€", entry.Rent)),
		widget.NewLabel(fmt.Sprintf("ΑΠΟ: %s", entry.Start)),
		widget.NewLabel(fmt.Sprintf("ΕΩΣ: %s", entry.End)),
		widget.NewLabel(fmt.Sprintf("Είδος Καλ/γειας: %s", entry.Type)),
		widget.NewLabel(fmt.Sprintf("Στρέμματα: %.3f", entry.Size)),
		layout.NewSpacer(),
		coordsContainer,
		buttonsContainer,
	)

	popup := widget.NewModalPopUp(content, appState.window.Canvas())

	// Iterate over all canvas objects of content
	// and if its a label enable text wrapping
	for _, obj := range content.Objects {
		if label, ok := obj.(*widget.Label); ok {
			label.Wrapping = fyne.TextWrapWord
		}
	}

	editButton.OnTapped = func() {
		editForm, err := editForm(appState, entry.ID)
		if err != nil {
			log.Printf("error creating desktopEditForm for %d: %v", entry.ID, err)
		}
		appState.window.SetContent(editForm)
		popup.Hide()
	}
	closeButton.OnTapped = func() {
		popup.Hide()
	}
	deleteButton.OnTapped = func() {
		err := delEntry(appState.db, entry.ID)
		if err != nil {
			dialog.ShowError(err, appState.window)
		}
		dialog.ShowInformation("Deleted", fmt.Sprintf("Deteled entry %d!", entry.ID), appState.window)
		if err != nil {
			dialog.ShowError(err, appState.window)
		}
		// Need to do all that to update the list in the mainView after a deletion
		// TODO: Maybe make it better
		*entries, err = getAll(appState.db)
		if err != nil {
			log.Printf("Error updating the list: %v", err)
		}
		list.Refresh()
		popup.Hide()
	}

	contentMinHeight := content.MinSize().Height
	popup.Resize(fyne.NewSize(320, contentMinHeight))

	popup.Show()
	fmt.Printf("Popup displayed for: %d", entry.ID)
}

// Show Calendar for easy date picking
// There is another calendar widget floating around maybe I should check it out
// or customize this to add a button for the year?
func showCalendar(entry *widget.Entry, window fyne.Window) {
	log.Printf("Showing popup date picker.")
	calendar := xwidget.NewCalendar(time.Now(), func(t time.Time) {
		dateString := t.Format("02-01-2006")
		entry.SetText(dateString)

		for _, overlay := range window.Canvas().Overlays().List() {
			overlay.Hide()
		}
	})

	popup := dialog.NewCustom("Select Date", "Cancel", calendar, window)
	popup.Show()
}
