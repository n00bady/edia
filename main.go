package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    _ "github.com/mattn/go-sqlite3"
)

// An entry is a placeholder for the actual structure that
// will hold all the data about each Rent and Contracts for 
// farm land
type Entry struct {
    ID        int
    Content   string
    Timestamp time.Time
}

func main() {
    // Initialize Fyne app
    myApp := app.New()
    myWindow := myApp.NewWindow("Entry Saver")

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
    entry := widget.NewEntry()
    entry.SetPlaceHolder("Type your entry here...")

    // Save button
    saveBtn := widget.NewButton("Save", func() {
        if entry.Text != "" {
            newEntry := Entry{
                Content:   entry.Text,
                Timestamp: time.Now(),
            }
            err := saveEntry(db, newEntry)
            if err != nil {
                log.Printf("Error saving entry: %v", err)
                return
            }
            entry.SetText("")
            log.Printf("Saved entry: %s", newEntry.Content)
        }
    })

    // Layout
    content := container.NewVBox(
        widget.NewLabel("Enter your text:"),
        entry,
        saveBtn,
    )

    // Set window content and size
    myWindow.SetContent(content)
    myWindow.Resize(fyne.NewSize(300, 300))

    // Runing the app starts here
    myWindow.ShowAndRun()
}

func initDB(dbPath string) (*sql.DB, error) {
    // Make sure that the directory exists (old android version need this)
    err := os.MkdirAll(filepath.Dir(dbPath), 0755)
    if err != nil {
        return nil, fmt.Errorf("error creating directory: %v", err)
    }

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("error opening database: %v", err)
    }

    createTableSQL := `
        CREATE TABLE IF NOT EXISTS entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            content TEXT NOT NULL,
            timestamp DATETIME NOT NULL
        );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, fmt.Errorf("error creating table: %v", err)
    }

    return db, nil
}

func saveEntry(db *sql.DB, entry Entry) error {
    insertSQL := `
        INSERT INTO entries (content, timestamp) 
        VALUES (?, ?)`
    result, err := db.Exec(insertSQL, entry.Content, entry.Timestamp)
    if err != nil {
        return fmt.Errorf("error inserting entry: %v", err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("error getting last insert ID: %v", err)
    }
    entry.ID = int(id)

    return nil
}
