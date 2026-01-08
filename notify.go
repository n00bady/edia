package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func notify(appState *AppState) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	checkEndDateNotification(appState)

	for range ticker.C {
		checkEndDateNotification(appState)
	}
}

func checkEndDateNotification(appState *AppState) {
	entries, err := getAllEntries(appState.db)
	if err != nil {
		log.Println("Error getting all entries from db!")
	}

	var notifications []string
	for _, e := range entries {
		endTime, err := time.Parse("02-01-2006", e.End)
		if err != nil {
			log.Printf("Error parsing date for entry %d: %v", e.ID, err)
			continue
		}
		daysLeft := int(math.Ceil(time.Until(endTime).Hours() / 24))

		if daysLeft <= 30 && daysLeft > 0 {
			notifications = append(notifications, fmt.Sprintf("%s ends in %d days (%s)", e.Name, daysLeft, e.End))
		}

		if len(notifications) > 0 {
			content := strings.Join(notifications, "\n")
			appState.app.SendNotification(&fyne.Notification{
				Title:   "End dates approaching!",
				Content: content,
			})
			log.Println("Notification for end dates send!")
		}
	}
}
