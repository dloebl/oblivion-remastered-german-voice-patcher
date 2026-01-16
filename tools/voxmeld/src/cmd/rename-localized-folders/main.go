package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var animationsStop chan struct{}
var animationsCounter int

var progressMutex sync.Mutex

var logger *log.Logger
var logFile *os.File

func main() {
	setupLogging()
	defer logFile.Close()

	extractFolderBsa := "tmp/bsa_original/sound/voice"

	// Get current time
	timeStart := time.Now()

	renameLocalizedFolders(extractFolderBsa, timeStart)
}

// Funktion zum Einrichten des Loggings
func setupLogging() {
	// Erstelle Log-Verzeichnis, wenn es nicht existiert
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen des Log-Verzeichnisses: %v\n", err)
		return
	}

	// Öffne die Log-Datei
	logFile, err = os.OpenFile("logs/rename-localized-folders.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Log-Datei: %v\n", err)
		return
	}

	// Initialisiere den Logger
	logger = log.New(logFile, "", log.LstdFlags)
}

// Funktion zum Loggen und gleichzeitigen Ausgeben einer Nachricht
func logAndPrint(message string) {
	fmt.Println(message)
	if logger != nil {
		logger.Println(message)
	}
}

func renameLocalizedFolders(extractFolderBsa string, timeStart time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32

	maxParallelism := runtime.NumCPU() * 2
	if maxParallelism > 16 {
		maxParallelism = 16
	}
	semaphore := make(chan struct{}, maxParallelism)

	language := "german"

	file, err := os.Open(filepath.Join("custom", language, "folders.txt"))
	if err != nil {
		fmt.Println("ERROR: Could not open file:", err)
		return
	}
	defer file.Close()

	folderRenameMappings := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		row := scanner.Text()

		if strings.HasPrefix(row, "::") {
			// Row is comment
			continue
		}

		rowParts := strings.SplitN(row, "=", 2)
		if len(rowParts) != 2 {
			fmt.Println("Invalid row (no '=' character found):", row)
			continue
		}

		localizedFolderName := strings.TrimSpace(rowParts[0])
		englishFolderName := strings.TrimSpace(rowParts[1])

		folderRenameMappings[localizedFolderName] = englishFolderName
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR: Could not read file:", err)
		return
	}

	dlcs, _ := filepath.Glob(filepath.Join(extractFolderBsa, "*"))
	for _, dlc := range dlcs {
		for localizedFolderName, _ := range folderRenameMappings {
			oldPath := filepath.Join(dlc, localizedFolderName)
			if _, err := os.Stat(oldPath); err == nil {
				totalFiles++
			}
		}
	}

	if totalFiles > 0 {
		// Show information
		fmt.Printf("\n-------------------------------------------------------------------\n")
		fmt.Printf("Folders:     %d folders\n", totalFiles)
		fmt.Printf("Status:      Start renaming folders\n")
		fmt.Printf("-------------------------------------------------------------------\n")

		// Start animation in the background
		animationsStop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					progressMutex.Lock()
					animationsCounter++
					updateAnimatedProgressBar(
						int(atomic.LoadInt32(&processedFiles)),
						totalFiles,
						timeStart,
						animationsCounter)
					progressMutex.Unlock()
				case <-animationsStop:
					return
				}
			}
		}()

		for _, dlc := range dlcs {
			for localizedFolderName, englishFolderName := range folderRenameMappings {
				oldPath := filepath.Join(dlc, localizedFolderName)
				if _, err := os.Stat(oldPath); err == nil {
					newPath := filepath.Join(dlc, englishFolderName)

					// Rename folder to english language variant
					wg.Add(1)
					semaphore <- struct{}{}

					go func(src, dst string) {
						defer wg.Done()
						defer func() { <-semaphore }()

						if err := os.Rename(src, dst); err != nil {
							logAndPrint(fmt.Sprintf("ERROR: Could not rename folder %s to %s: %v", src, dst, err))
						}
						atomic.AddInt32(&processedFiles, 1)
					}(oldPath, newPath)
				}
			}
		}

		wg.Wait()

		// Animationsschleife stoppen
		close(animationsStop)
		time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

		// Zeige finalen Fortschrittsbar
		progressMutex.Lock()
		updateAnimatedProgressBar(
			int(atomic.LoadInt32(&processedFiles)),
			totalFiles,
			timeStart,
			animationsCounter)
		progressMutex.Unlock()
	}
}

// updateAnimatedProgressBar stellt einen animierten Fortschrittsbar in der Konsole dar
func updateAnimatedProgressBar(currentProgress, amountTotal int, timeStart time.Time, animationsCounter int) {
	width := 40 // Breite des Balkens in Zeichen

	// Berechne Prozentsatz
	percent := float64(currentProgress) / float64(amountTotal)

	// Berechne Anzahl der filleden Zeichen
	filled := int(percent * float64(width))

	// Animations-Zeichen
	animationSymbols := []string{"|", "/", "-", "\\"}
	animationSymbol := animationSymbols[animationsCounter%len(animationSymbols)]

	// ASCII-Ladebar Zeichen
	filledChar := "#"

	// Erstelle den Ladebar
	bar := strings.Repeat(filledChar, filled) + strings.Repeat("-", width-filled)

	// Erstelle einen eingebetteten Animations-Cursor im Ladebar
	if filled < width {
		position := filled
		barRunes := []rune(bar)
		barRunes[position] = []rune(animationSymbol)[0]
		bar = string(barRunes)
	}

	// Lösche die aktuelle Zeile und zeige nur den Balken ohne verstrichene Zeit an
	fmt.Printf("\r[%s] %3.0f%% %d/%d files processed",
		bar, percent*100, currentProgress, amountTotal)
}
