package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// App initialization
	log.Printf("Sarting the EDIA App!")
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
		AppInst.window.Resize(fyne.NewSize(600, 650))
	}

	log.Printf("Running...")
	// Runing the app
	AppInst.window.ShowAndRun()
}

// App initialization
func InitApp() (*AppState, error) {
	log.Printf("Initializing the application...")
	myApp := app.NewWithID("xyz.n00bady.edia")
	myWindow := myApp.NewWindow("EDIA")

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
		window: myWindow,
	}, nil
}
