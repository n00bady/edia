## About
Μια cross-platform εφαρμογή διαχείρισης μισθωμάτων και συμβολαίων αγροτικής γης.
Αναπτύχθηκε σε GO και το framework Fyne.io για την αποθήκευσή χρησιμοποιεί sqlite.

## Build Instructions
### Requirements
- go >=1.25
- fyne.io
- android-ndk (see fyne.io documentation for android builds)
### Build
- For local desktop build `go build .`
- For android build `fyne package -os android --app-id xyz.n00bady.agricoman -icon assets/icon.png -release`
### Tests
Just run `go test`
