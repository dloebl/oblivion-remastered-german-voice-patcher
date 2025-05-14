package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"log"
	"runtime"
	"strings"
	"time"
)

var animationsStop chan struct{}
var animationsCounter int

var progressMutex sync.Mutex

var logger *log.Logger
var logFile *os.File

func main() {
	logAndPrint("\nNext step: Apply replace fix")
	logAndPrint("\nInitialize Logging...")

	setupLogging()
	defer logFile.Close()

	execDir, err := os.Executable()
	if err != nil {
		logAndPrint(fmt.Sprintf("ERROR: Could not find execution directory: %v", err))
		return
	}
	execDir = filepath.Dir(execDir)

	// Show header
	fmt.Printf("\n====================== Apply replace fix ======================\n")
	fmt.Printf("Status:      			Preparing files\n")
	fmt.Printf("-------------------------------------------------------------------\n")

	// Get current time
	timeStart := time.Now()

	extractFolder := filepath.Join(execDir, "..", "..", "tmp", "bsa_original")
	customFolder := filepath.Join(execDir, "..", "..", "custom")

	logAndPrint("\nStarting process...")
	processFiles(extractFolder, customFolder, timeStart)

	logAndPrint("\nFinished replace fix step!")
}

// Funktion zum Einrichten des Loggings
func setupLogging() {
	// Erstelle Log-Verzeichnis, wenn es nicht existiert
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("ERROR: Could not create logs directory: %v\n", err)
		return
	}

	// Öffne die Log-Datei
	logFile, err = os.OpenFile("logs/apply-replace-fix.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("ERROR: Could not open log file: %v\n", err)
		return
	}

	// Initialisiere die Logger
	logger = log.New(logFile, "", log.LstdFlags)
}

// Funktion zum Loggen und gleichzeitigen Ausgeben einer Nachricht
func logAndPrint(message string) {
	fmt.Println(message)
	if logger != nil {
		logger.Println(message)
	}
}

func processFiles(extractFolder, customFolder string, timeStart time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32

	maxParallelism := runtime.NumCPU() * 2
	if maxParallelism > 16 {
		maxParallelism = 16
	}
	semaphore := make(chan struct{}, maxParallelism)

	// TODO: Listen to setting instead once introduced
	language := "german"

	file, err := os.Open(filepath.Join("custom", language, "replacements.txt"))
	if err != nil {
		fmt.Println("ERROR: Could not open file:", err)
		return
	}
	defer file.Close()

	fileReplaceMappings := make(map[string]string)

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

		fileToReplace := strings.TrimSpace(rowParts[0])
		fileToReplaceWith := strings.TrimSpace(rowParts[1])

		fileReplaceMappings[fileToReplace] = fileToReplaceWith
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR: Could not read file:", err)
		return
	}

	// Check for optional dlc
	if _, err := os.Stat(filepath.Join(extractFolder, "sound/voice/dlchorsearmor.esp/")); err == nil {
		fileOptional, err := os.Open(filepath.Join("custom", language, "replacements-optional.txt"))
		if err != nil {
			fmt.Println("ERROR: Could not open file:", err)
			return
		}
		defer fileOptional.Close()

		scanner = bufio.NewScanner(fileOptional)
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

			fileToReplace := strings.TrimSpace(rowParts[0])
			fileToReplaceWith := strings.TrimSpace(rowParts[1])

			fileReplaceMappings[fileToReplace] = fileToReplaceWith
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("ERROR: Could not read file:", err)
			return
		}
	}

	totalFiles = len(fileReplaceMappings)

	// Show information
	fmt.Printf("\n-------------------------------------------------------------------\n")
	fmt.Printf("Files:     		%d files\n", totalFiles)
	fmt.Printf("Status:      	Start replacing files\n")
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

	for old, new := range fileReplaceMappings {
		fileToReplace := filepath.Join(extractFolder, old)
		fileToReplaceWith := filepath.Join(customFolder, new)

		if _, err := os.Stat(fileToReplaceWith); err != nil {
			// If no custom file available, replace with other game file
			// Use if there is a valid replacement file from game for untranslated/broken files
			fileToReplaceWith = filepath.Join(extractFolder, new)

			if _, err := os.Stat(fileToReplaceWith); err != nil {
				// If no custom file available, replace with other game file
				// Use if there is a valid replacement file from game for untranslated/broken files
				logAndPrint(fmt.Sprintf("ERROR: Could not find file to use as replacement '%s': %v", fileToReplaceWith, err))
				atomic.AddInt32(&processedFiles, 1)
			}
		}

		// Beginn processing file
		wg.Add(1)
		semaphore <- struct{}{}

		go func(dst, src string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := copyFile(dst, src); err != nil {
				logAndPrint(fmt.Sprintf("ERROR: Could not replace '%s' with '%s': %v", dst, src, err))
			}
			atomic.AddInt32(&processedFiles, 1)
		}(fileToReplace, fileToReplaceWith)
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

	// Calculate time total
	timeTotal := time.Since(timeStart)
	fmt.Printf("\n\nAll replace fixes have been applied in %s\n", timeTotal)
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

func copyFile(dst, src string) error {
	// Erstelle Zielverzeichnis falls nicht vorhanden
	os.MkdirAll(filepath.Dir(dst), 0755)

	if src == dst {
		return errors.New("ERROR: Source and destination must be different files")
	}

	dstDir := filepath.Dir(dst)
	if info, err := os.Stat(dstDir); err != nil {
		if os.IsNotExist(err) {
			return errors.New("ERROR: Destination directory does not exist")
		}
		return err
	} else if !info.IsDir() {
		return errors.New("ERROR: Destination path's parent is not a directory")
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}