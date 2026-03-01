package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

func addForm(appState *AppState) (fyne.CanvasObject, error) {
	entriesMap := make(map[string]*widget.Entry)

	durationLabel := widget.NewLabel("Διαρκεια")

	var landLords []OwnerDetails
	var renters []RenterDetails
	var selectedFileBytes []byte

	// TODO: Maybe I should remove the entriesMap so I can use NumericalEntry for some
	labelsEntries := []string{
		"Όνομα Εγγραφής",
		"ATAK",
		"KAEK",
		"Στρέμματα",
		"Είδος Καλ/γειας",
		"Μίσθωμα",
	}

	for i, l := range labelsEntries {
		switch i {
		case 1, 2:
			entriesMap[l] = NewFilteredEntry(`[^0-9]`, l)
		case 4:
			entriesMap[l] = newEntryWithLabel(l)
		case 3, 5:
			entriesMap[l] = NewFilteredEntry(`[^0-9.]`, l)
		default:
			entriesMap[l] = NewFilteredEntry(`[^a-zA-Z]`, l)
		}
	}

	for i := range 4 {
		lat := NewFilteredEntry(`[^0-9.]`, fmt.Sprintf("Πλάτος %d", i+1))
		entriesMap[fmt.Sprintf("Πλάτος %d", i+1)] = lat
		long := NewFilteredEntry(`[^0-9.]`, fmt.Sprintf("Μήκος %d", i+1))
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

	buttonEmisth := widget.NewButtonWithIcon("Μησθωτήριο", theme.FileIcon(), func() {
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

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})
	// containerEmisth := container.NewVBox(labelEmisth, buttonEmisth)

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
		// TODO: Add a check to make sure they have unique names!
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
		mainview, err := contractView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
			dialog.ShowError(err, appState.window)
		}
		appState.window.SetContent(container.NewStack(appState.bg, mainview))
	})

	// Cancel button to go back
	backButton := widget.NewButton("Cancel", func() {
		tmp, err := contractView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(container.NewStack(appState.bg, body))
	})

	// --Different Layouts for Mobile and Desktop--
	if fyne.CurrentDevice().IsMobile() {
		// --Mobile layout--
		landlordsContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές"), addLandLord), landLordsLabelsContainer)
		rentersContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές"), addRenter), renterLabelsContainer)

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
			container.NewGridWithColumns(1, entriesMap["Μίσθωμα"], buttonEmisth),
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
		landlordsContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές"), addLandLord), landLordsLabelsContainer)
		rentersContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές"), addRenter), renterLabelsContainer)

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
			container.NewGridWithColumns(2, entriesMap["Μίσθωμα"], buttonEmisth),
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

	for i, l := range labelsEntries {
		switch i {
		case 1, 2:
			entriesMap[l] = NewFilteredEntry(`[^0-9]`, l)
		case 4:
			entriesMap[l] = newEntryWithLabel(l)
		case 3, 5:
			entriesMap[l] = NewFilteredEntry(`[^0-9.]`, l)
		default:
			entriesMap[l] = NewFilteredEntry(`[^a-zA-Z]`, l)
		}
	}

	// Assign values to entries from the selected Entry
	entriesMap["Όνομα"].SetText(selectedEntry.Name)
	entriesMap["ATAK"].SetText(strconv.FormatUint(uint64(selectedEntry.ATAK), 10))
	entriesMap["KAEK"].SetText(selectedEntry.KAEK)
	entriesMap["Στρέμματα"].SetText(strconv.FormatFloat(selectedEntry.Size, 'f', -1, 64))
	entriesMap["Είδος Καλ/γειας"].SetText(selectedEntry.Type)
	entriesMap["Μίσθωμα"].SetText(strconv.FormatFloat(selectedEntry.Rent, 'f', -1, 64))

	for i := range 4 {
		lat := NewFilteredEntry(`[^0-9.]`, fmt.Sprintf("Πλάτος %d", i+1))
		lat.SetText(strconv.FormatFloat(selectedEntry.Coords[i].Latitude, 'f', -1, 64))
		entriesMap[fmt.Sprintf("Πλάτος %d", i+1)] = lat

		long := NewFilteredEntry(`[^0-9.]`, fmt.Sprintf("Μήκος %d", i+1))
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

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})
	containerEmisth := container.NewVBox(labelEmisth, buttonEmisth)

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
		tmp, err := contractView(appState)
		if err != nil {
			log.Printf("error constructing list layout: %v", err)
		}
		body := container.NewBorder(nil, nil, nil, nil, tmp)
		appState.window.SetContent(container.NewStack(appState.bg, body))
	})

	// landLordsContainer := container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές: "), addLandLord, landLordsLabelsContainer)
	landLordsContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Εκμισθωτές"), addLandLord), landLordsLabelsContainer)
	// rentersContainer := container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές: "), addRenters, rentersLabelContainer)
	rentersContainer := container.NewVBox(container.NewBorder(nil, nil, widget.NewLabel("Μισθωτές"), addRenters), rentersLabelContainer)

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
	listViewButton := widget.NewButton("Συμβόλαια", func() {
		lView, err := contractView(appState)
		if err != nil {
			log.Printf("error constructing nameView: %v\n", err)
			dialog.ShowError(err, appState.window)
		}

		appState.window.SetContent(container.NewStack(appState.bg, lView))
	})

	landLordButton := widget.NewButton("Ιδιοκτήτες", func() {
		view, err := ownersView(appState)
		if err != nil {
			log.Printf("error constructing ownersView: %v\n", err)
			dialog.ShowError(err, appState.window)
		}

		appState.window.SetContent(container.NewStack(appState.bg, view))
	})

	renterButton := widget.NewButton("Μισθωτές", func() {
		view, err := rentersView(appState)
		if err != nil {
			log.Printf("error constructing renterView: %v\n", err)
		}

		appState.window.SetContent(container.NewStack(appState.bg, view))
	})

	customLayout := NewCenteredButtonsLayout(200, 60, 20)
	content := container.New(customLayout, listViewButton, landLordButton, renterButton)
	body := container.NewStack(appState.bg, appState.logo, content)

	return body, nil
}

func rentersView(appState *AppState) (fyne.CanvasObject, error) {
	log.Println("Creating the rentersView...")
	renters, err := getAllRenters(appState.db)
	log.Printf("query Results: %v\n", renters)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(renters))
	for i, s := range renters {
		items[i] = s
	}
	list := buildList(appState, items)

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
			addRenter(appState)
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
			addRenter(appState)
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
	log.Println("Creating the ownerView...")
	landlords, err := getAllOwners(appState.db)
	log.Printf("query Results: %v\n", landlords)
	if err != nil {
		return nil, err
	}

	items := make([]any, len(landlords))
	for i, s := range landlords {
		items[i] = s
	}

	list := buildList(appState, items)

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
			addOwner(appState)
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
			addOwner(appState)
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

func contractView(appState *AppState) (fyne.CanvasObject, error) {
	log.Printf("Creating the contractView...")
	entries, err := getAllEntries(appState.db)
	if err != nil {
		return nil, err
	}

	list := widget.NewList(
		func() int {
			return len(entries)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("Name")
			dateLabel := widget.NewLabel("End Date")

			return container.NewBorder(nil, nil, nil, dateLabel, nameLabel)
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			log.Printf("Updating item with ID: %d", lii)
			if lii < 0 || lii >= len(entries) {
				log.Printf("Invalid item ID: %d", lii)
				return
			}
			entry := entries[lii]
			box := co.(*fyne.Container)
			nameLabel := box.Objects[0].(*widget.Label)
			dateLabel := box.Objects[1].(*widget.Label)
			nameLabel.SetText(entry.Name)
			dateLabel.SetText(fmt.Sprintf("Λήξη: %s", entry.End))
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
			tmp, err := addForm(appState)
			if err != nil {
				log.Printf("error constructing mobile layout: %v", err)
			}
			appState.window.SetContent(container.NewStack(appState.bg, tmp))
		})

		backButton = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
			}
			appState.window.SetContent(container.NewStack(appState.bg, tmp))
		})
	} else {
		addButton = widget.NewButtonWithIcon("Add New Entry", theme.ContentAddIcon(), func() {
			tmp, err := addForm(appState)
			if err != nil {
				log.Printf("error constructing desktop layout: %v", err)
			}
			appState.window.SetContent(container.NewStack(appState.bg, tmp))
		})

		backButton = widget.NewButtonWithIcon("Back", theme.ContentUndoIcon(), func() {
			tmp, err := mainView(appState)
			if err != nil {
				log.Printf("error constructing main layout: %v", err)
			}
			appState.window.SetContent(container.NewStack(appState.bg, tmp))
		})
	}
	addButton.Resize(fyne.NewSize(200, 200))
	backButton.Resize(fyne.NewSize(200, 200))

	emptyMsg := canvas.NewText("Add something to begin.", theme.ForegroundColor())
	emptyMsg.TextSize = 18
	emptyMsg.Alignment = fyne.TextAlignCenter
	emptyMsg.TextStyle = fyne.TextStyle{Italic: true}
	emptyContainer := container.NewCenter(emptyMsg)
	if list.Length() == 0 {
		emptyContainer.Show()
	} else {
		emptyContainer.Hide()
	}

	body := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		emptyContainer,
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

	log.Printf("contractView created successfully!")

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

	misthButton := widget.NewButton("ΜΙΣΘΩΤΗΡΙΟ", func() {
		openFile(entry, appState)
	})

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
			misthButton,
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
		appState.window.SetContent(container.NewStack(appState.bg, editForm))
		popup.Hide()
	}
	closeButton.OnTapped = func() {
		popup.Hide()
	}
	deleteButton.OnTapped = func() {
		dlg := dialog.NewConfirm("Επιβεβαίωση Διαγραφής", fmt.Sprintf("Είσαι σίγουρος;"), func(b bool) {
			if b {
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
		}, appState.window)
		dlg.Show()
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

func showOwnerEntriesPopup(appState *AppState, owners *[]OwnerDetails, labelContainer fyne.Container, onSave func(string)) {
	log.Printf(">>> landlord: %v\n", owners)
	var owner OwnerDetails
	var selectedFileBytes []byte

	inviSpacer := func(height float32) fyne.CanvasObject {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(1, height))

		return spacer
	}

	firstName := widget.NewEntry()
	firstName.PlaceHolder = "Όνομα"

	lastName := widget.NewEntry()
	lastName.PlaceHolder = "Επώνυμο"

	fathersName := widget.NewEntry()
	fathersName.PlaceHolder = "Όνομα Πατρός"

	afm := xwidget.NewNumericalEntry() // ΑΦΜ
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
				dialog.ShowError(err, appState.window)
				return
			}
		}, appState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})

	homeAdress := widget.NewEntry()
	homeAdress.PlaceHolder = "Διεύθυνση"

	phoneNum := xwidget.NewNumericalEntry()
	phoneNum.PlaceHolder = "Τηλέφωνο"

	email := widget.NewEntry()
	email.PlaceHolder = "e-mail"

	accountInfo := widget.NewEntry()
	accountInfo.PlaceHolder = "Στοιχέια Λογιστή"

	notes := widget.NewEntry()

	containerE9 := container.NewHBox(labelE9, buttonE9)

	form := container.NewPadded(container.NewVBox(
		inviSpacer(10),
		firstName,
		inviSpacer(14),
		lastName,
		inviSpacer(14),
		fathersName,
		inviSpacer(14),
		afm,
		inviSpacer(14),
		adt,
		inviSpacer(14),
		containerE9,
		inviSpacer(14),
		homeAdress,
		inviSpacer(14),
		phoneNum,
		inviSpacer(14),
		email,
		inviSpacer(14),
		accountInfo,
		inviSpacer(14),
		notes,
		inviSpacer(10),
	))
	scrolledForm := container.NewVScroll(form)
	scrolledForm.SetMinSize(fyne.NewSize(400, 300))

	d := dialog.NewCustomConfirm("Enter Owner Details", "Save", "Cancel", scrolledForm, func(ok bool) {
		if ok {
			log.Println("Saving Owner named: ", firstName.Text+" "+lastName.Text)
			if firstName.Text == "" || lastName.Text == "" {
				dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), appState.window)
			}
			afm2Uint, err := strconv.ParseUint(afm.Text, 10, 0)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), appState.window)
			}

			owner.FirstName = firstName.Text
			owner.LastName = lastName.Text
			owner.FathersName = fathersName.Text
			owner.AFM = uint(afm2Uint)
			owner.ADT = adt.Text
			owner.E9 = selectedFileBytes
			owner.HomeAddress = homeAdress.Text
			owner.PhoneNumber = phoneNum.Text
			owner.Email = email.Text
			owner.AccountantInfo = accountInfo.Text
			owner.Notes = notes.Text

			*owners = append(*owners, owner)
			onSave(owner.FirstName + " " + owner.LastName)
			log.Println("Updated entry successfully!")
			return
		} else {
			log.Println("User probably clicked cancel.")
			return
		}
	}, appState.window)

	d.Resize(fyne.NewSize(400, 600))
	d.Show()
}

func showRenterEntriesPopup(appState *AppState, renters *[]RenterDetails, labelContainer fyne.Container, onSave func(string)) {
	log.Printf(">>> renters: %v\n", renters)
	var renter RenterDetails
	var selectedFileBytes []byte

	inviSpacer := func(height float32) fyne.CanvasObject {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(1, height))

		return spacer
	}

	firstName := widget.NewEntry()
	firstName.PlaceHolder = "Όνομα"

	lastName := widget.NewEntry()
	lastName.PlaceHolder = "Επώνυμο"

	fathersName := widget.NewEntry()
	fathersName.PlaceHolder = "Όνομα Πατρός"

	afm := xwidget.NewNumericalEntry() // ΑΦΜ
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
				dialog.ShowError(err, appState.window)
				return
			}
		}, appState.window)

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})

	notes := widget.NewEntry()

	containerE9 := container.NewHBox(labelE9, buttonE9)

	form := container.NewPadded(container.NewVBox(
		inviSpacer(10),
		firstName,
		inviSpacer(14),
		lastName,
		inviSpacer(14),
		fathersName,
		inviSpacer(14),
		afm,
		inviSpacer(14),
		adt,
		inviSpacer(14),
		containerE9,
		inviSpacer(14),
		notes,
		inviSpacer(10),
	))
	scrolledForm := container.NewVScroll(form)
	scrolledForm.SetMinSize(fyne.NewSize(400, 300))

	d := dialog.NewCustomConfirm("Στοιχεία Μησθωτή", "Save", "Cancel", scrolledForm,
		func(ok bool) {
			if ok {
				log.Println("Saving Renter named: ", firstName.Text+" "+lastName.Text)
				if firstName.Text == "" || lastName.Text == "" {
					dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), appState.window)
				}
				afmINT, err := strconv.ParseUint(afm.Text, 10, 0)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), appState.window)
				}

				renter.FirstName = firstName.Text
				renter.LastName = lastName.Text
				renter.FathersName = fathersName.Text
				renter.AFM = uint(afmINT)
				renter.ADT = adt.Text
				renter.E9 = selectedFileBytes
				renter.Notes = notes.Text

				*renters = append(*renters, renter)
				onSave(renter.FirstName + " " + renter.LastName)
				log.Println("Updated entry successfully!")
				return
			} else {
				log.Println("User probably clicked cancel.")
				return
			}
		}, appState.window)

	d.Resize(fyne.NewSize(400, 600))
	d.Show()
}

func addRenter(appState *AppState) error {
	var renter RenterDetails
	var selectedFileBytes []byte

	inviSpacer := func(height float32) fyne.CanvasObject {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(1, height))

		return spacer
	}

	entryList, err := getAllEntries(appState.db)
	if err != nil {
		return err
	}

	var opts []string
	for _, e := range entryList {
		opts = append(opts, e.Name) // Need to make sure they have unique names...
	}

	contractSelect := widget.NewSelectEntry(opts)
	contractSelect.PlaceHolder = "Επιλoγή Συμβόλαιου"
	// Find a better solution for this...
	contractSelect.OnChanged = func(s string) {
		found := false
		for _, option := range opts {
			if s == option {
				found = true
				break
			}
		}
		if !found && s != "" {
			contractSelect.SetText(contractSelect.PlaceHolder)
		}
	}

	firstName := newEntryWithLabel("Όνομα")
	lastName := newEntryWithLabel("Επώνυμο")
	fathersName := newEntryWithLabel("Όνομα Πατρός")
	afm := xwidget.NewNumericalEntry()
	afm.PlaceHolder = "Α.Φ.Μ."
	adt := newEntryWithLabel("Α.Δ.Τ.")
	labelE9 := widget.NewLabel("E9")
	buttonE9 := widget.NewButtonWithIcon("Add E9", theme.FileIcon(), func() {
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

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})
	notes := widget.NewEntry()

	containerE9 := container.NewHBox(labelE9, buttonE9)

	form := container.NewPadded(container.NewVBox(
		inviSpacer(10),
		contractSelect,
		inviSpacer(14),
		firstName,
		inviSpacer(14),
		lastName,
		inviSpacer(14),
		fathersName,
		inviSpacer(14),
		afm,
		inviSpacer(14),
		adt,
		inviSpacer(14),
		containerE9,
		inviSpacer(14),
		notes,
		inviSpacer(10),
	))
	scrolledForm := container.NewVScroll(form)
	scrolledForm.SetMinSize(fyne.NewSize(400, 300))

	d := dialog.NewCustomConfirm("Στοιχεία Μησθωτή", "Save", "Cancel", scrolledForm,
		func(ok bool) {
			if ok {
				var selectedEntry Entry
				var set bool
				for _, e := range entryList {
					if e.Name == contractSelect.Text {
						selectedEntry = e
						set = true
					} else {
						set = false
					}
				}
				// not ideal but works
				if !set {
					dialog.ShowInformation("Error", "Cannot find the selected contract.", appState.window)
				}

				log.Println("Saving Renter named: ", firstName.Text+" "+lastName.Text)
				if firstName.Text == "" || lastName.Text == "" {
					dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), appState.window)
				}
				afmINT, err := strconv.ParseUint(afm.Text, 10, 0)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), appState.window)
				}

				renter.FirstName = firstName.Text
				renter.LastName = lastName.Text
				renter.FathersName = fathersName.Text
				renter.AFM = uint(afmINT)
				renter.ADT = adt.Text
				renter.E9 = selectedFileBytes
				renter.Notes = notes.Text

				// TODO: check if it exists already?
				selectedEntry.Renters = append(selectedEntry.Renters, renter)

				err = updateEntry(appState.db, selectedEntry)
				if err != nil {
					dialog.ShowInformation("Error", "Cannot update entry with the new Renter.", appState.window)
				}

				log.Println("Updated entry successfully!")

			} else {
				log.Println("User probably clicked cancel.")
				return
			}
		}, appState.window)

	d.Resize(fyne.NewSize(400, 600))
	d.Show()

	return nil
}

func addOwner(appState *AppState) error {
	var owner OwnerDetails
	var selectedFileBytes []byte

	inviSpacer := func(height float32) fyne.CanvasObject {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(1, height))

		return spacer
	}

	entryList, err := getAllEntries(appState.db)
	if err != nil {
		return err
	}

	var opts []string
	for _, e := range entryList {
		opts = append(opts, e.Name)
	}

	contractSelect := widget.NewSelectEntry(opts)
	contractSelect.PlaceHolder = "Επιλογή Συμβολαίου"
	// Find a better solution for this...
	contractSelect.OnChanged = func(s string) {
		found := false
		for _, option := range opts {
			if s == option {
				found = true
				break
			}
		}
		if !found && s != "" {
			contractSelect.SetText(contractSelect.PlaceHolder)
		}
	}

	firstName := newEntryWithLabel("Όνομα")
	lastName := newEntryWithLabel("Επώνυμο")
	fathersName := newEntryWithLabel("Όνομα Πατρός")
	afm := xwidget.NewNumericalEntry()
	afm.PlaceHolder = "Α.Φ.Μ."
	adt := newEntryWithLabel("Α.Δ.Τ.")

	labelE9 := widget.NewLabel("E9")
	buttonE9 := widget.NewButtonWithIcon("Add E9", theme.FileIcon(), func() {
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

		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".pdf"}))
		dlg.Show()
	})

	homeAdress := newEntryWithLabel("Διεύθυνση")
	phoneNum := xwidget.NewNumericalEntry()
	phoneNum.PlaceHolder = "Τηλέφωνο"
	email := newEntryWithLabel("e-mail")
	accountInfo := newEntryWithLabel("Στοιχέια Λογιστή")
	notes := widget.NewEntry()

	containerE9 := container.NewHBox(labelE9, buttonE9)

	form := container.NewPadded(container.NewVBox(
		inviSpacer(10),
		contractSelect,
		inviSpacer(14),
		firstName,
		inviSpacer(14),
		lastName,
		inviSpacer(14),
		fathersName,
		inviSpacer(14),
		afm,
		inviSpacer(14),
		adt,
		inviSpacer(14),
		containerE9,
		inviSpacer(14),
		homeAdress,
		inviSpacer(14),
		phoneNum,
		inviSpacer(14),
		email,
		inviSpacer(14),
		accountInfo,
		inviSpacer(14),
		notes,
		inviSpacer(10),
	))
	scrolledForm := container.NewVScroll(form)
	scrolledForm.SetMinSize(fyne.NewSize(400, 300))

	d := dialog.NewCustomConfirm("Enter Owner Details", "Save", "Cancel", scrolledForm, func(ok bool) {
		if ok {
			var selectedEntry Entry
			var set bool
			fmt.Println("selected: " + contractSelect.Text)
			for _, e := range entryList {
				if e.Name == contractSelect.Text {
					selectedEntry = e
					set = true
				} else {
					set = false
				}
			}
			if !set {
				dialog.ShowInformation("Error", "Cannot find the selected contract.", appState.window)
				return
			}

			log.Println("Saving Owner named: ", firstName.Text+" "+lastName.Text)
			if firstName.Text == "" || lastName.Text == "" {
				dialog.ShowError(fmt.Errorf("you need to add at least a first and last name"), appState.window)
			}
			afm2Uint, err := strconv.ParseUint(afm.Text, 10, 0)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Α.Φ.Μ. not valid."), appState.window)
			}

			owner.FirstName = firstName.Text
			owner.LastName = lastName.Text
			owner.FathersName = fathersName.Text
			owner.AFM = uint(afm2Uint)
			owner.ADT = adt.Text
			owner.E9 = selectedFileBytes
			owner.HomeAddress = homeAdress.Text
			owner.PhoneNumber = phoneNum.Text
			owner.Email = email.Text
			owner.AccountantInfo = accountInfo.Text
			owner.Notes = notes.Text

			selectedEntry.Owners = append(selectedEntry.Owners, owner)

			err = updateEntry(appState.db, selectedEntry)
			if err != nil {
				dialog.ShowInformation("Error", "Cannot update entry with the new Owner.", appState.window)
			}

			log.Println("Updated entry successfully!")
		} else {
			log.Println("User probably clicked cancel.")
			return
		}
	}, appState.window)

	d.Resize(fyne.NewSize(400, 600))
	d.Show()

	return nil
}
