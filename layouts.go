package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
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

	durationLabel := widget.NewLabel("Διαρκεια")

	var landLords []OwnerDetails
	var renters []RenterDetails
	var selectedFileBytes []byte

	labelsEntries := []string{
		"Όνομα Εγγραφής",
		"Μισθωτής",
		"ATAK",
		"KAEK",
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

	// Button to add multiple landlords
	landLordsLabelsContainer := container.NewVBox()
	addLandLord := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		showOwnerEntriesPopup(appState, &landLords, *landLordsLabelsContainer, func(s string) {
			landLordsLabelsContainer.Add(widget.NewLabel(s))
			landLordsLabelsContainer.Refresh()
		})
	})
	renterLabelsContainer := container.NewVBox()
	addRenter := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		showRenterEntriesPopup(appState, &renters, *renterLabelsContainer, func(s string) {
			renterLabelsContainer.Add(widget.NewLabel(s))
			renterLabelsContainer.Refresh()
		})
	})

	// Button to add Geo Coordinates
	addGeoLocButton := widget.NewButtonWithIcon("Add GeoCoordinates", theme.ContentAddIcon(), func() {
		showGeoLocForm(appState, entriesMap)
	})

	labelEmisth := widget.NewLabel("Μησθωτήριο")
	buttonEmisth := widget.NewButtonWithIcon("Add Μησθωτήριο", theme.FileIcon(), func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			selectedFileBytes, err = io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, appState.window)
				return
			}
		}, appState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png"}))
		dlg.Show()
	})
	containerEmisth := container.NewHBox(labelEmisth, buttonEmisth)

	// Save button
	saveBtn := widget.NewButton("Αποθήκευση", func() {
		// Convert to float64 and gather the coordinates
		coords := make([]Coordinates, 0, 4)
		for i := range 4 {
			latValue, errLat := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Πλάτος %d", i+1)].Text, 5)
			longValue, errLon := ParseFloatToXDecimals(entriesMap[fmt.Sprintf("Μήκος %d", i+1)].Text, 5)

			if errLat != nil || errLon != nil {
				log.Printf("Error parsing coordinates")
				dialog.ShowError(errLat, appState.window)
			}

			coords = append(coords, Coordinates{Latitude: latValue, Longitude: longValue})
		}

		atak, err := strconv.ParseInt(entriesMap["ATAK"].Text, 10, 32)
		if err != nil {
			log.Printf("Error parsing ATAK")
			dialog.ShowError(err, appState.window)
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

		log.Printf("--- landlords: %v", landLords)
		for _, l := range landLords {
			log.Printf("--- landlord: %s, %s\n", l.FirstName, l.LastName)
		}

		log.Printf("--- renters: %v\n", renters)
		for _, r := range renters {
			log.Printf("--- renter: %s, %s\n", r.FirstName, r.LastName)
		}

		// We build the new entry here
		newEntry := Entry{
			Name:      entriesMap["Όνομα Εγγραφής"].Text,
			Owners:    landLords,
			Renters:   renters,
			Coords:    coords,
			Timestamp: time.Now(),
			ATAK:      uint(atak),
			KAEK:      entriesMap["KAEK"].Text,
			Size:      size,
			Type:      entriesMap["Είδος Καλ/γειας"].Text,
			Start:     start_input.Text,
			End:       end_input.Text,
			Rent:      money,
			emisth:    selectedFileBytes,
		}

		err = saveEntry(appState.db, newEntry)
		if err != nil {
			log.Printf("Error saving entry: %v", err)
			dialog.ShowError(err, appState.window)
			return
		}

		log.Println("Entry saved!")
		dialog.ShowInformation("Database:", "Saved successfully!", appState.window)

		// return to mainView
		mainview, err := nameView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
			dialog.ShowError(err, appState.window)
		}
		appState.window.SetContent(mainview)
	})

	// Cancel button to go back
	backButton := widget.NewButton("Cancel", func() {
		tmp, err := nameView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	// --Different Layouts for Mobile and Desktop--
	if fyne.CurrentDevice().IsMobile() {
		// --Mobile layout--
		landlordsContainer := container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές: "), addLandLord, landLordsLabelsContainer)
		rentersContainer := container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές: "), addRenter, renterLabelsContainer)

		leftContainer := container.NewVBox(
			entriesMap["Όνομα Εγγραφής"],
			landlordsContainer,
			rentersContainer,
			layout.NewSpacer(),
			addGeoLocButton,
		)

		rightContainer := container.NewVBox(
			entriesMap["ATAK"],
			entriesMap["KAEK"],
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			containerEmisth,
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

		// TODO: Modify to bring it inline with the rest of the changes.
		allInputs := []fyne.CanvasObject{
			entriesMap["Όνομα Εγγραφής"],
			entriesMap["ATAK"],
			entriesMap["KAEK"],
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			// entriesMap["Πλάτος 1"], entriesMap["Μήκος 1"],
			// entriesMap["Πλάτος 2"], entriesMap["Μήκος 2"],
			// entriesMap["Πλάτος 3"], entriesMap["Μήκος 3"],
			// entriesMap["Πλάτος 4"], entriesMap["Μήκος 4"],
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
		landlordsContainer := container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές"), addLandLord, landLordsLabelsContainer)
		rentersContainer := container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές"), addRenter, renterLabelsContainer)

		// LEFT
		left_container := container.NewVBox(
			entriesMap["Όνομα Εγγραφής"],
			landlordsContainer,
			rentersContainer,
			layout.NewSpacer(),
			addGeoLocButton,
		)

		// RIGHT
		right_container := container.NewVBox(
			entriesMap["ATAK"],
			entriesMap["KAEK"],
			entriesMap["Στρέμματα"],
			entriesMap["Είδος Καλ/γειας"],
			entriesMap["Μίσθωμα"],
			containerEmisth,
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

func editForm(appState *AppState, id uint) (fyne.CanvasObject, error) {
	log.Printf("Creating desktop edit form...")
	var selectedFileBytes []byte

	entriesMap := make(map[string]*widget.Entry)

	selectedEntry, err := getEntry(appState.db, id)
	if err != nil {
		return nil, err
	}

	landLords := selectedEntry.Owners
	renters := selectedEntry.Renters

	durationLabel := widget.NewLabel("Διαρκεια")

	labelsEntries := []string{
		"Όνομα",
		"ATAK",
		"KAEK",
		"Στρέμματα",
		"Είδος Καλ/γειας",
		"Μίσθωμα",
	}

	for _, l := range labelsEntries {
		tmpEnt := newEntryWithLabel(l)
		entriesMap[l] = tmpEnt
	}

	// Assign values to entries from the selected Entry
	entriesMap["Όνομα"].SetText(selectedEntry.Name)
	entriesMap["ATAK"].SetText(strconv.FormatUint(uint64(selectedEntry.ATAK), 10))
	entriesMap["KAEK"].SetText(selectedEntry.KAEK)
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

	landLordsLabelsContainer := container.NewVBox()
	for _, l := range landLords {
		landLordsLabelsContainer.Add(widget.NewLabel(l.FirstName + " " + l.LastName))
	}
	addLandLord := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		showOwnerEntriesPopup(appState, &landLords, *landLordsLabelsContainer, func(s string) {
			landLordsLabelsContainer.Add(widget.NewLabel(s))
			landLordsLabelsContainer.Refresh()
		})
	})
	rentersLabelContainer := container.NewVBox()
	for _, r := range renters {
		rentersLabelContainer.Add(widget.NewLabel(r.FirstName + " " + r.LastName))
	}
	addRenters := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		showRenterEntriesPopup(appState, &renters, *rentersLabelContainer, func(s string) {
			rentersLabelContainer.Add(widget.NewLabel(s))
			rentersLabelContainer.Refresh()
		})
	})

	addGeoLocButton := widget.NewButtonWithIcon("Add GeoCoordinates", theme.ContentAddIcon(), func() {
		showGeoLocForm(appState, entriesMap)
	})

	labelEmisth := widget.NewLabel("Μησθωτήριο")
	buttonEmisth := widget.NewButtonWithIcon("Add Μησθωτήριο", theme.FileIcon(), func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			selectedFileBytes, err = io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, appState.window)
				return
			}
		}, appState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png"}))
		dlg.Show()
	})
	containerEmisth := container.NewHBox(labelEmisth, buttonEmisth)

	saveBtn := widget.NewButton("Αποθήκευση", func() {
		// Convert to float64 and gather the coordinates
		coords := make([]Coordinates, 0, 4)
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

			coords = append(coords, Coordinates{Latitude: latValue, Longitude: longValue})
		}

		atak, err := strconv.ParseInt(entriesMap["ATAK"].Text, 10, 32)
		if err != nil {
			log.Printf("Error parsing ATAK")
			dialog.ShowError(err, appState.window)
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
			ID:        id,
			Name:      entriesMap["Όνομα"].Text,
			Owners:    landLords,
			Renters:   renters,
			Coords:    coords,
			Timestamp: time.Now(),
			ATAK:      uint(atak),
			KAEK:      entriesMap["KAEK"].Text,
			Size:      size,
			Type:      entriesMap["Είδος Καλ/γειας"].Text,
			Start:     start_input.Text,
			End:       end_input.Text,
			Rent:      money,
			emisth:    selectedFileBytes,
		}
		// editedEntry.LandlordName = append(editedEntry.LandlordName, entriesMap["Εκμισθωτής"].Text)

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
		tmp, err := nameView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(body)
	})

	landLordsContainer := container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές: "), addLandLord, landLordsLabelsContainer)
	rentersContainer := container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές: "), addRenters, rentersLabelContainer)

	// LEFT
	left_container := container.NewVBox(
		entriesMap["Όνομα"],
		landLordsContainer,
		rentersContainer,
		layout.NewSpacer(),
		addGeoLocButton,
	)

	// RIGHT
	right_container := container.NewVBox(
		entriesMap["ATAK"],
		entriesMap["KAEK"],
		entriesMap["Στρέμματα"],
		entriesMap["Είδος Καλ/γειας"],
		entriesMap["Μίσθωμα"],
		containerEmisth,
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
		lView, err := nameView(appState)
		if err != nil {
			log.Printf("error constructing nameView: %v\n", err)
			dialog.ShowError(err, appState.window)
		}

		appState.window.SetContent(lView)
	})

	landLordButton := widget.NewButton("Ιδιοκτήτες", func() {
		view, err := ownersView(appState)
		if err != nil {
			log.Printf("error constructing ownersView: %v\n", err)
			dialog.ShowError(err, appState.window)
		}

		appState.window.SetContent(view)
	})

	renterButton := widget.NewButton("Μισθωτές", func() {
		view, err := rentersView(appState)
		if err != nil {
			log.Printf("error constructing renterView: %v\n", err)
		}

		appState.window.SetContent(view)
	})

	body := container.New(layout.NewCenterLayout(), container.NewVBox(listViewButton, landLordButton, renterButton))

	return body, nil
}

func rentersView(appState *AppState) (fyne.CanvasObject, error) {
	log.Println("Creating the rentersView...")
	renters, err := getAllRenters(appState.db)
	log.Printf("query Results: %v\n", renters)
	if err != nil {
		return nil, err
	}

	list := widget.NewList(
		func() int {
			return len(renters)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			if lii < 0 || lii >= len(renters) {
				log.Printf("Invalid item ID: %d\n", lii)
				return
			}

			renter := renters[lii]
			label, ok := co.(*widget.Label)
			if !ok {
				log.Printf("Canvas object is not *widget.Label, it's: %s\n", fmt.Sprintf("%T", co))
				return
			}
			label.SetText(fmt.Sprintf("%s", renter.FirstName+" "+renter.LastName))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d\n", id)
		if id >= 0 && id < len(renters) {
			log.Printf("Showing popup for item: %v\n", renters[id].LastName)
			showDetailsRenterPopup(appState, &renters[id])
			list.UnselectAll()
		}
	}

	var addButton fyne.CanvasObject
	var backButton fyne.CanvasObject

	if fyne.CurrentDevice().IsMobile() {
		addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			tmp, err := AddForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})

		backButton = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
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

		backButton = widget.NewButtonWithIcon("Back", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})
	}
	addButton.Resize(fyne.NewSize(200, 200))
	backButton.Resize(fyne.NewSize(200, 200))

	body := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		container.NewVScroll(list),
		container.New(
			layout.NewVBoxLayout(),
			layout.NewSpacer(),
			container.New(
				layout.NewHBoxLayout(),
				layout.NewSpacer(),
				container.NewPadded(backButton),
				container.NewPadded(addButton),
			),
		),
	)
	log.Println("rentersView created successfully!")

	return body, nil
}

func ownersView(appState *AppState) (fyne.CanvasObject, error) {
	log.Println("Creating the landlordView...")
	landlords, err := getAllOwners(appState.db)
	log.Printf("query Results: %v\n", landlords)
	if err != nil {
		return nil, err
	}

	list := widget.NewList(
		func() int {
			return len(landlords)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			if lii < 0 || lii >= len(landlords) {
				log.Printf("Invalid item ID: %d\n", lii)
				return
			}

			landlord := landlords[lii]
			label, ok := co.(*widget.Label)
			if !ok {
				log.Printf("Canvas object is not *widget.Label, its: %s\n", fmt.Sprintf("%T", co))
				return
			}
			label.SetText(fmt.Sprintf("%s", landlord.FirstName+" "+landlord.LastName))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d\n", id)
		if id >= 0 && id < len(landlords) {
			log.Printf("Showing popup for item: %v\n", landlords[id].LastName)
			showDetailsOwnerPopup(appState, &landlords[id])
			list.UnselectAll()
		}
	}

	var addButton fyne.CanvasObject
	var backButton fyne.CanvasObject

	if fyne.CurrentDevice().IsMobile() {
		addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			tmp, err := AddForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})

		backButton = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
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

		backButton = widget.NewButtonWithIcon("Back", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})
	}
	addButton.Resize(fyne.NewSize(200, 200))
	backButton.Resize(fyne.NewSize(200, 200))

	body := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		container.NewVScroll(list),
		container.New(
			layout.NewVBoxLayout(),
			layout.NewSpacer(),
			container.New(
				layout.NewHBoxLayout(),
				layout.NewSpacer(),
				container.NewPadded(backButton),
				container.NewPadded(addButton),
			),
		),
	)
	log.Println("landlordsView created successfully!")

	return body, nil
}

func nameView(appState *AppState) (fyne.CanvasObject, error) {
	log.Printf("Creating the listView...")
	entries, err := getAllEntries(appState.db)
	if err != nil {
		return nil, err
	}

	list := widget.NewList(
		func() int {
			return len(entries)
		},
		func() fyne.CanvasObject {
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
			label.SetText(fmt.Sprintf("%s", entry.Name))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d", id)
		if id >= 0 && id < len(entries) {
			log.Printf("Showing popup for item: %d\n", entries[id].ID)
			showDetailsPopup(entries[id], appState, list, &entries, &entries[id].Owners, &entries[id].Renters)
			list.UnselectAll()
		}
	}

	var addButton fyne.CanvasObject
	var backButton fyne.CanvasObject

	if fyne.CurrentDevice().IsMobile() {
		addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			tmp, err := AddForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})

		backButton = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
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

		backButton = widget.NewButtonWithIcon("Back", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
			}
			appState.window.SetContent(tmp)
		})
	}
	addButton.Resize(fyne.NewSize(200, 200))
	backButton.Resize(fyne.NewSize(200, 200))

	body := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		container.NewVScroll(list),
		container.New(
			layout.NewVBoxLayout(),
			layout.NewSpacer(),
			container.New(
				layout.NewHBoxLayout(),
				layout.NewSpacer(),
				container.NewPadded(backButton),
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

func showDetailsRenterPopup(appState *AppState, renter *RenterDetails) {
	log.Printf("Showing popup for: %v\n", renter.LastName)

	closeButton := widget.NewButton("Close", nil)

	buttonsContainer := container.NewVBox(
		closeButton,
	)

	entries, err := getRenterEntries(appState.db, *renter)
	if err != nil {
		dialog.ShowError(err, appState.window)
		return
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
			label.SetText(fmt.Sprintf("%d: %s", entry.ID, entry.Name))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d\n", id)
		if id >= 0 && id < len(entries) {
			log.Printf("Showing popup for item: %d\n", entries[id].ID)
			showDetailsPopup(entries[id], appState, list, &entries, &entries[id].Owners, &entries[id].Renters)
			list.UnselectAll()
		}
	}

	scrollableList := container.NewVScroll(list)
	scrollableList.SetMinSize(fyne.NewSize(appState.window.Canvas().Size().Width*0.7, appState.window.Canvas().Size().Height*0.7))

	content := container.NewVBox(
		scrollableList,
		buttonsContainer,
	)

	popup := widget.NewModalPopUp(content, appState.window.Canvas())

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
	fmt.Printf("Popup displayed for: %v", renter.LastName)
}

func showDetailsOwnerPopup(appState *AppState, owner *OwnerDetails) {
	log.Printf("Showing popup for: %v\n", owner.LastName)

	closeButton := widget.NewButton("Close", nil)

	buttonsContainer := container.NewVBox(
		closeButton,
	)

	entries, err := getOwnerEntries(appState.db, *owner)
	if err != nil {
		dialog.ShowError(err, appState.window)
		return
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
			label.SetText(fmt.Sprintf("%s", entry.Name))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		log.Printf("Selected item: %d", id)
		if id >= 0 && id < len(entries) {
			log.Printf("Showing popup for item: %d", entries[id].ID)
			showDetailsPopup(entries[id], appState, list, &entries, &entries[id].Owners, &entries[id].Renters)
			list.UnselectAll()
		}
	}

	scrollableList := container.NewVScroll(list)
	scrollableList.SetMinSize(fyne.NewSize(appState.window.Canvas().Size().Width*0.7, appState.window.Canvas().Size().Height*0.7))

	content := container.NewVBox(
		scrollableList,
		buttonsContainer,
	)

	popup := widget.NewModalPopUp(content, appState.window.Canvas())

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
	fmt.Printf("Popup displayed for: %v", owner.LastName)
}

// Details popup for the list
func showDetailsPopup(entry Entry, appState *AppState, list *widget.List, entries *[]Entry, owners *[]OwnerDetails, renters *[]RenterDetails) {
	log.Printf("Showing popup for: %d", entry.ID)

	editButton := widget.NewButton("Edit", nil)
	closeButton := widget.NewButton("Close", nil)
	deleteButton := widget.NewButton("Delete", nil)

	buttonsContainer := container.NewVBox(
		closeButton,
		container.NewGridWithColumns(2, editButton, deleteButton),
	)

	coordsContainer := container.NewVBox(widget.NewLabel("Συντετγμένες: "))
	for i, c := range entry.Coords {
		str := fmt.Sprintf("\t%d Lat.: %f Long.: %f", i+1, c.Latitude, c.Longitude)
		coordsContainer.Add(widget.NewLabel(str))
	}

	ownersContainer := container.NewVBox(widget.NewLabel("Εκμισθωτής/ές: "))
	for _, o := range *owners {
		ownersContainer.Add(widget.NewLabel("\t" + o.FirstName + " " + o.LastName))
	}

	rentersContainer := container.NewVBox(widget.NewLabel("Μισθωτής/ες: "))
	for _, r := range *renters {
		rentersContainer.Add(widget.NewLabel("\t" + r.FirstName + " " + r.LastName))
	}

	// Add all the details!
	scrollableContainer := container.NewVScroll(
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("ID: %d", entry.ID)),
			widget.NewLabel(fmt.Sprintf("%s", entry.Name)),
			ownersContainer,
			rentersContainer,
			widget.NewLabel(fmt.Sprintf("Μίσθωμα: %.2f€", entry.Rent)),
			widget.NewLabel(fmt.Sprintf("ΑΠΟ: %s", entry.Start)),
			widget.NewLabel(fmt.Sprintf("ΕΩΣ: %s", entry.End)),
			widget.NewLabel(fmt.Sprintf("Είδος Καλ/γειας: %s", entry.Type)),
			widget.NewLabel(fmt.Sprintf("Στρέμματα: %.3f", entry.Size)),
			layout.NewSpacer(),
			coordsContainer,
		),
	)
	content := container.NewBorder(nil, buttonsContainer, nil, nil, scrollableContainer)

	popup := widget.NewModalPopUp(content, appState.window.Canvas())

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
		*entries, err = getAllEntries(appState.db)
		if err != nil {
			log.Printf("Error updating the list: %v", err)
		}
		list.Refresh()
		popup.Hide()
	}

	popup.Resize(fyne.NewSize(appState.window.Canvas().Size().Width*0.8, appState.window.Canvas().Size().Height*0.8))

	popup.Show()
	fmt.Printf("Popup displayed for: %d", entry.ID)
}

func showGeoLocForm(appState *AppState, entriesMap map[string]*widget.Entry) {
	content := container.NewVBox()
	closeButton := widget.NewButton("close", nil)
	for i := range 4 {
		coordContainer := container.NewGridWithColumns(2, entriesMap[fmt.Sprintf("Πλάτος %d", i+1)], entriesMap[fmt.Sprintf("Μήκος %d", i+1)])
		content.Add(coordContainer)
		content.Add(layout.NewSpacer())
	}

	popup := widget.NewModalPopUp(content, appState.window.Canvas())
	closeButton.OnTapped = func() {
		popup.Hide()
	}
	content.Add(closeButton)

	popup.Resize(fyne.NewSize(appState.window.Canvas().Size().Width*0.66, appState.window.Canvas().Size().Height*0.66))

	popup.Show()
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

func showOwnerEntriesPopup(AppState *AppState, owners *[]OwnerDetails, labelContainer fyne.Container, onSave func(string)) {
	log.Printf(">>> landlord: %v\n", owners)
	var landlord OwnerDetails
	var selectedFileBytes []byte

	firstName := widget.NewEntry()
	firstName.PlaceHolder = "Όνομα"

	lastName := widget.NewEntry()
	lastName.PlaceHolder = "Επώνυμο"

	fathersName := widget.NewEntry()
	fathersName.PlaceHolder = "Όνομα Πατρός"

	afm := widget.NewEntry() // ΑΦΜ
	afm.PlaceHolder = "Α.Φ.Μ."

	adt := widget.NewEntry() // ΑΔΤ
	adt.PlaceHolder = "Α.Δ.Τ."

	labelE9 := widget.NewLabel("E9")
	buttonE9 := widget.NewButtonWithIcon("Add E9", theme.FileIcon(), func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			selectedFileBytes, err = io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, AppState.window)
				return
			}
		}, AppState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})

	notes := widget.NewEntry()

	cancelButton := widget.NewButton("Cancel", nil)
	saveButton := widget.NewButton("Save", nil)

	containerE9 := container.NewHBox(labelE9, buttonE9)
	buttonContainer := container.NewHBox(cancelButton, saveButton)
	content := container.NewVBox(firstName, lastName, fathersName, afm, adt, containerE9, notes, buttonContainer)

	popup := widget.NewModalPopUp(content, AppState.window.Canvas())

	cancelButton.OnTapped = func() {
		popup.Hide()
	}

	saveButton.OnTapped = func() {
		log.Println("Saving a landlord named: ", firstName.Text+""+lastName.Text)
		if firstName.Text == "" || lastName.Text == "" {
			dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), AppState.window)
		}
		afm2Uint, err := strconv.ParseUint(afm.Text, 10, 0)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), AppState.window)
		}

		landlord.FirstName = firstName.Text
		landlord.LastName = lastName.Text
		landlord.FathersName = fathersName.Text
		landlord.AFM = uint(afm2Uint)
		landlord.ADT = adt.Text
		landlord.E9 = selectedFileBytes
		landlord.Notes = notes.Text

		log.Printf(">>> landlord struct: %v\n", landlord)
		*owners = append(*owners, landlord)
		labelContainer.Add(widget.NewLabel(landlord.FirstName + " " + landlord.LastName))

		onSave(landlord.FirstName + " " + landlord.LastName)

		popup.Hide()
	}

	popup.Resize(fyne.NewSize(300, 500))
	popup.Show()
}

func showRenterEntriesPopup(AppState *AppState, renters *[]RenterDetails, labelContainer fyne.Container, onSave func(string)) {
	log.Printf(">>> renters: %v\n", renters)
	var renter RenterDetails
	var selectedFileBytes []byte

	firstName := widget.NewEntry()
	firstName.PlaceHolder = "Όνομα"

	lastName := widget.NewEntry()
	lastName.PlaceHolder = "Επώνυμο"

	fathersName := widget.NewEntry()
	fathersName.PlaceHolder = "Όνομα Πατρός"

	afm := widget.NewEntry() // ΑΦΜ
	afm.PlaceHolder = "Α.Φ.Μ."

	adt := widget.NewEntry() // ΑΔΤ
	adt.PlaceHolder = "Α.Δ.Τ."

	labelE9 := widget.NewLabel("E9")
	buttonE9 := widget.NewButtonWithIcon("Add E9", theme.FileIcon(), func() {
		dlg := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			selectedFileBytes, err = io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, AppState.window)
				return
			}
		}, AppState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})

	notes := widget.NewEntry()

	cancelButton := widget.NewButton("Cancel", nil)
	saveButton := widget.NewButton("Save", nil)

	containerE9 := container.NewHBox(labelE9, buttonE9)
	buttonContainer := container.NewHBox(cancelButton, saveButton)
	content := container.NewVBox(firstName, lastName, fathersName, afm, adt, containerE9, notes, buttonContainer)

	popup := widget.NewModalPopUp(content, AppState.window.Canvas())

	cancelButton.OnTapped = func() {
		popup.Hide()
	}

	saveButton.OnTapped = func() {
		log.Println("Saving a landlord named: ", firstName.Text+""+lastName.Text)
		if firstName.Text == "" || lastName.Text == "" {
			dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), AppState.window)
		}
		afmINT, err := strconv.ParseUint(afm.Text, 10, 0)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), AppState.window)
		}

		renter.FirstName = firstName.Text
		renter.LastName = lastName.Text
		renter.FathersName = fathersName.Text
		renter.AFM = uint(afmINT)
		renter.ADT = adt.Text
		renter.E9 = selectedFileBytes
		renter.Notes = notes.Text

		log.Printf(">>> renters struct: %v\n", renter)
		*renters = append(*renters, renter)
		labelContainer.Add(widget.NewLabel(renter.FirstName + " " + renter.LastName))

		onSave(renter.FirstName + " " + renter.LastName)

		popup.Hide()
	}

	popup.Resize(fyne.NewSize(300, 500))
	popup.Show()
}
