package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Parses a string to d amount of decimals to float
func ParseFloatToXDecimals(n string, d int) (float64, error) {
	if d < 0 || d > 15 {
		return 0, fmt.Errorf("number of decimals must be non-negative and smaller than 15")
	}
	if n == "" {
		return 0, fmt.Errorf("string is empty")
	}
	if len(n) > 15 {
		return 0, fmt.Errorf("string too long")
	}

	v, err := strconv.ParseFloat(n, 64)
	log.Printf("Parsed to float: %f", v)
	if err != nil {
		return 0, fmt.Errorf("cannot parse float: %v", err)
	}

	decimals := math.Pow(10, float64(d))
	rounded := math.Round(v*decimals) / decimals

	return rounded, nil
}

// Valid Coordinates
func IsValidLatitude(f float64) (bool, error) {
	if f > 90 || f < -90 {
		return false, nil
	}

	return true, nil
}

func IsValidLongitude(f float64) (bool, error) {
	if f > 180 || f < -180 {
		return false, nil
	}

	return true, nil
}

// Checks for negative nubmers
func IsNegative(f float64) (bool, error) {
	if f >= 0 {
		return false, nil
	}

	return true, nil
}

// Truncate a float32 to 2 decimals
func TruncateFloatTo2Decimals(f float64) float64 {
	return float64(int(f*100)) / 100
}

func Contains(sl []string, str string) bool {
	for _, v := range sl {
		if v == str {
			return true
		}
	}

	return false
}

func openFile(e Entry, appState *AppState) error {
	if len(e.emisth) == 0 {
		dialog.ShowInformation("Empty file", "No file data!", appState.window)
		return nil
	}

	var guessedExt string

	mimeType := http.DetectContentType(e.emisth)

	extMap := map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"application/pdf": ".pdf",
	}

	if ext, ok := extMap[mimeType]; ok {
		guessedExt = ext
	} else {
		dialog.ShowInformation("Unsupported file", "file type not supported", appState.window)
		return nil
	}

	if fyne.CurrentDevice().IsMobile() {
		storage := fyne.CurrentApp().Storage()

		name := "temp-open-" + e.Name + guessedExt
		writerCloser, err := storage.Create(name)
		if err != nil {
			return err
		}
		defer writerCloser.Close()

		_, err = writerCloser.Write(e.emisth)
		if err != nil {
			return err
		}

		fURI := writerCloser.URI()
		u, err := url.Parse(fURI.String())
		if err != nil {
			return err
		}

		time.AfterFunc(60*time.Second, func() {
			_ = storage.Remove(name)
		})

		err = fyne.CurrentApp().OpenURL(u)
		if err != nil {
			dialog.ShowInformation("Failed to show the file", "File created but failed to open.", appState.window)
			_ = storage.Remove(name)
			return err
		}

		return nil
	} else {
		base := strings.TrimSuffix(filepath.Base(e.Name), guessedExt)
		if base == "" {
			base = "blobfile"
		}
		pattern := base + "-" + guessedExt
		tmpFile, err := os.CreateTemp("", pattern)
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}

		_, err = tmpFile.Write(e.emisth)
		if err != nil {
			return fmt.Errorf("failed to write temp file: %v", err)
		}
		tmpFile.Close()

		fileURI := fmt.Sprintf("file://%s", filepath.ToSlash(tmpFile.Name()))
		u, err := url.Parse(fileURI)
		if err != nil {
			return fmt.Errorf("cannot parse file URI: %v", err)
		}
		err = fyne.CurrentApp().OpenURL(u)
		if err != nil {
			msg := fmt.Sprintf("cannot open file: %s\nError: %v", tmpFile.Name(), err)
			dialog.ShowInformation("Failed to open file", msg, appState.window)
			return err
		}
	}

	return nil
}

func buildList(data []any) *widget.List {
	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			if lii < 0 || lii >= len(data) {
				return
			}
			label, ok := co.(*widget.Label)
			if !ok {
				log.Println("CanvasObject is not *widget.Label! It's: %s\n)", fmt.Sprintf("%T", co))
				return
			}

			switch t := data[lii].(type) {
			case RenterDetails:
				label.SetText(fmt.Sprintf("%s", t.FirstName+" "+t.LastName))
			case OwnerDetails:
				label.SetText(fmt.Sprintf("%s", t.FirstName+" "+t.LastName))
			default:
				log.Println("Unknown type.")
			}
		},
	)

	return list
}
