package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var logger *log.Logger
var logFile *os.File

func main() {
	convertFolder := "tmp/toConvert"
	extractFolderBsa := "tmp/bsa_original/sound/voice"

	// Get current time
	timeStart := time.Now()

	processFiles(convertFolder, extractFolderBsa, timeStart)
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

func processFiles(convertFolder string, extractFolderBsa string, timeStart time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32
	var progressMutex sync.Mutex

	var animationsStop chan struct{}
	var animationsCounter int

	maxParallelism := runtime.NumCPU() * 2
	if maxParallelism > 16 {
		maxParallelism = 16
	}
	semaphore := make(chan struct{}, maxParallelism)

	// Create  convertFolder if not already present
	os.MkdirAll(convertFolder, 0755)

	// Count files to process
	dlcs, _ := filepath.Glob(filepath.Join(extractFolderBsa, "*"))
	for _, dlc := range dlcs {
		races, _ := filepath.Glob(filepath.Join(dlc, "*"))
		for _, race := range races {
			variants, _ := filepath.Glob(filepath.Join(race, "*"))
			for _, variant := range variants {
				files, _ := filepath.Glob(filepath.Join(variant, "*"))
				for _, f := range files {
					if strings.EqualFold(filepath.Ext(f), ".mp3") {
						totalFiles++
					}
				}
			}
		}
	}

	// Show information
	fmt.Printf("\n-------------------------------------------------------------------\n")
	fmt.Printf("Files:     	%d Audio files\n", totalFiles)
	fmt.Printf("Status:     Start copying renamed files\n")
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
		races, _ := filepath.Glob(filepath.Join(dlc, "*"))
		for _, race := range races {
			raceName := filepath.Base(race)

			variants, _ := filepath.Glob(filepath.Join(race, "*"))
			for _, variant := range variants {
				variantName := filepath.Base(variant)

				// NOTE: Don't use "*.mp3" for unfilteredFiles as it's case sensitive
				unfilteredFiles, _ := filepath.Glob(filepath.Join(variant, "*"))
				var files []string
				for _, f := range unfilteredFiles {
					if strings.EqualFold(filepath.Ext(f), ".mp3") {
						files = append(files, f)
					}
				}

				for _, file := range files {
					// Rename mp3 file and copy it to 'tmp/toConvert'
					mp3Target := filepath.Join(convertFolder, fmt.Sprintf("%s_%s_%s", raceName, variantName, filepath.Base(file)))
					wg.Add(1)
					semaphore <- struct{}{}
					
					go func(src, dst string) {
						defer wg.Done()
						defer func() { <-semaphore }()
						
						if err := copyFile(src, dst); err != nil {
							fmt.Printf("ERROR: Could not copy %s to %s: %v", src, dst, err)
						}
						atomic.AddInt32(&processedFiles, 1)
					}(file, mp3Target)
				}
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

func copyFile(src, dst string) error {
	// Erstelle Zielverzeichnis falls nicht vorhanden
	os.MkdirAll(filepath.Dir(dst), 0755)

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
