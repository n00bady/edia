package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// App initialization
	log.Printf("Sarting the AgriCoMan App!")
	AppInst, err := InitApp()
	if err != nil {
		log.Printf("error initializing the App: %v", err)
	} 
	defer AppInst.db.Close()

	log.Printf("Constructing the intial view...")

	// Check if it runs on mobile or desktop and construct the apropriate layout
	var body fyne.CanvasObject

	body, err = mainView(AppInst)
	if err != nil {
		log.Fatalf("error constructing main view: %v", err)
	}

	// Set window content and size
	AppInst.window.SetContent(body)
	log.Printf("Window content set.")

	if !fyne.CurrentDevice().IsMobile() {
		log.Printf("It's not a mobile device set size to 600, 500")
		AppInst.window.Resize(fyne.NewSize(600, 750))
	}

	go notify(AppInst)

	log.Printf("Running...")
	// Runing the app
	AppInst.window.ShowAndRun()
}

// App initialization
func InitApp() (*AppState, error) {
	log.Printf("Initializing the application...")
	myApp := app.NewWithID("xyz.n00bady.agricoman")
	myWindow := myApp.NewWindow("AgriCoMan")
	myWindow.SetPadded(false)
	myWindow.SetMaster()

	bgImg := canvas.NewImageFromResource(resourceBackgroundJpg)
	bgImg.FillMode = canvas.ImageFillCover
	bgImg.ScaleMode = canvas.ImageScaleFastest
	overlay := canvas.NewRectangle(color.NRGBA{43, 45, 66, 128})
	background := container.NewStack(bgImg, overlay)

	lg := canvas.NewImageFromResource(resourceLogoPng)
	lg.FillMode = canvas.ImageFillContain
	lg.ScaleMode = canvas.ImageScaleSmooth
	lg.SetMinSize(fyne.NewSize(600, 300))
	logo := container.NewBorder(lg, nil, nil, nil, nil)

	// This is deprecated will be removed for fyne v3.0
	// Don't care!
	myApp.Settings().SetTheme(&MyTheme{base: theme.DarkTheme()})

	log.Printf("Initializing database...")

	dataDir := myApp.Storage().RootURI().Path()
	dbPath := filepath.Join(dataDir, "entries.db")
	log.Printf("Database path: %s", dbPath)

	db, err := initDB(dbPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return nil, err
	}
	if db.Ping() == nil {
		log.Printf("Opening existing database in: %s", dbPath)
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, fmt.Errorf("error opening the database: %v", err)
		}
	}

	log.Printf("App initialized successfully!")

	return &AppState{
		db:     db,
		app:    myApp,
		window: myWindow,
		bg:     background,
		logo:   logo,
	}, nil
}
