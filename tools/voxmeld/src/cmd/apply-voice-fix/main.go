package main

import (
	"bufio"
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


var animationsStop chan struct{}
var animationsCounter int

var progressMutex sync.Mutex

var logger *log.Logger
var logFile *os.File

var loggerDev1 *log.Logger
var logFileDev1 *os.File

var loggerDev2 *log.Logger
var logFileDev2 *os.File

func main() {
	setupLogging()
	defer logFile.Close()
	defer logFileDev1.Close()
	defer logFileDev2.Close()
	
	logAndPrint("Voice Fix Logging initialisiert", nil)

	
	// Show header
	fmt.Printf("\n====================== Apply voice fix ======================\n")
	fmt.Printf("Status:      		Start processing files\n")
	fmt.Printf("-------------------------------------------------------------------\n")

	// Get current time
	timeStart := time.Now()
	processFiles(timeStart)

	logAndPrint("\nVerarbeitung abgeschlossen!", nil)
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
	logFile, err = os.OpenFile("logs/apply-voice-fix.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Log-Datei: %v\n", err)
		return
	}

	// Erstelle Dev-Log-Verzeichnis, wenn es nicht existiert
	err = os.MkdirAll("logs/dev", 0755)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen des Log-Verzeichnisses: %v\n", err)
		return
	}

	// Öffne die Dev Log-Datei
	logFileDev1, err = os.OpenFile("logs/dev/voice-fix-missing-mp3.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Log-Datei: %v\n", err)
		return
	}

	// Öffne die Dev Log-Datei
	logFileDev2, err = os.OpenFile("logs/dev/voice-fix-remaster-path-not-found.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Log-Datei: %v\n", err)
		return
	}

	// Initialisiere die Logger
	logger = log.New(logFile, "", log.LstdFlags)
	loggerDev1 = log.New(logFileDev1, "", log.LstdFlags)
	loggerDev2 = log.New(logFileDev2, "", log.LstdFlags)
}

// Logging function
func logAndPrint(message string, alternativeLogger *log.Logger) {
	if alternativeLogger != nil {
		alternativeLogger.Println(message)
	} else {
		fmt.Println(message)
		if logger != nil {
			logger.Println(message)
		}
	}
}

func processFiles(timeStart time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32

	maxParallelism := runtime.NumCPU() * 2
	if maxParallelism > 16 {
		maxParallelism = 16
	}
	semaphore := make(chan struct{}, maxParallelism)

	language := "german"

	file, err := os.Open(filepath.Join("custom", language, "raceAlternatives.txt"))
	if err != nil {
		fmt.Println("ERROR: Could not open file:", err)
		return
	}
	defer file.Close()

	raceAlternativeMappings := make(map[string]string)

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

		remasterName := strings.TrimSpace(rowParts[0])
		originalName := strings.TrimSpace(rowParts[1])

		raceAlternativeMappings[remasterName] = originalName
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR: Could not read file:", err)
		return
	}

	// Zähle zuerst alle Dateien
	bsas , _ := filepath.Glob("tmp/bsa_remaster/*")
	for _, bsa := range bsas {
		dlcs, _ := filepath.Glob(filepath.Join(bsa, "sound/voice", "*"))
		for _, dlc := range dlcs {
			races, _ := filepath.Glob(filepath.Join(dlc, "*"))
			for _, race := range races {
				variants, _ := filepath.Glob(filepath.Join(race, "*"))
				for _, variant := range variants {

					prefixes := []string{"", "altvoice", "beggar"}
					for _, prefix := range prefixes {
						alternativeFolder := filepath.Join(variant, prefix)
						if _, err := os.Stat(alternativeFolder); err == nil {
							files, err := filepath.Glob(filepath.Join(alternativeFolder, "*.mp3"))
							if err != nil {
								logAndPrint(fmt.Sprintf("Fehler beim Auflisten von Dateien in %s: %v", alternativeFolder, err), logger)
								continue
							}

							for _, file := range files {
								bsaName := filepath.Base(bsa)
								dlcName := filepath.Base(dlc)
								raceName := filepath.Base(race)
								variantName := filepath.Base(variant)
								fileName := filepath.Base(file)
								remasterFile := filepath.Join("tmp/bsa_remaster", bsaName, "sound/voice", dlcName, raceName, variantName, prefix, fileName)

								if _, err := os.Stat(remasterFile); err != nil {
									logAndPrint(fmt.Sprintf("ERROR: Could not find Oblivion Remastered file to replace: %s", remasterFile), logger)
									logAndPrint(remasterFile, loggerDev2)
									continue
								}

								// Remove prefixes from file name
								fileAlternativeName := strings.Replace(fileName, "_altvoice_", "_", -1)
								fileAlternativeName = strings.Replace(fileAlternativeName, "_beggar_", "_", -1)

								// Replace race with alternative (NOTE: Could be different for other languages)
								raceAlternativeName := raceName
								for remasterName, originalName := range raceAlternativeMappings {
									if raceAlternativeName == remasterName {
										raceAlternativeName = originalName
										break
									}
								}

								// Check if matching language version of audio can be found
								languageFile := filepath.Join("tmp/bsa_original/sound/voice", dlcName, raceAlternativeName, variantName, fileAlternativeName)

								if _, err := os.Stat(languageFile); err == nil {
									totalFiles++
								} else {
									logAndPrint(fmt.Sprintf("%s", remasterFile), loggerDev1)
								}
							}
						}
					}
				}
			}
		}
	}

	// Show information
	fmt.Printf("\n-------------------------------------------------------------------\n")
	fmt.Printf("Files:     		%d files\n", totalFiles)
	fmt.Printf("Status:      	Start applying voice fix\n")
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

	for _, bsa := range bsas {
		dlcs, _ := filepath.Glob(filepath.Join(bsa, "sound/voice", "*"))
		for _, dlc := range dlcs {
			races, _ := filepath.Glob(filepath.Join(dlc, "*"))
			for _, race := range races {
				variants, _ := filepath.Glob(filepath.Join(race, "*"))
				for _, variant := range variants {
					prefixes := []string{"", "altvoice", "beggar"}
					for _, prefix := range prefixes {
						alternativeFolder := filepath.Join(variant, prefix)
						if _, err := os.Stat(alternativeFolder); err == nil {
							files, err := filepath.Glob(filepath.Join(alternativeFolder, "*.mp3"))
							if err != nil {
								continue
							}
							
							for _, file := range files {
								bsaName := filepath.Base(bsa)
								dlcName := filepath.Base(dlc)
								raceName := filepath.Base(race)
								variantName := filepath.Base(variant)
								fileName := filepath.Base(file)
								remasterFile := filepath.Join("tmp/bsa_remaster", bsaName, "sound/voice", dlcName, raceName, variantName, prefix, fileName)

								if _, err := os.Stat(remasterFile); err != nil {
									continue
								}

								// Remove prefixes from file name
								fileAlternativeName := strings.Replace(fileName, "_altvoice_", "_", -1)
								fileAlternativeName = strings.Replace(fileAlternativeName, "_beggar_", "_", -1)

								// Replace race folder name of remaster with the one of original
								raceAlternativeName := raceName
								for remasterName, originalName := range raceAlternativeMappings {
									if raceAlternativeName == remasterName {
										raceAlternativeName = originalName
										break
									}
								}

								// Check if language version of audio can be found
								languageFile := filepath.Join("tmp/bsa_original/sound/voice", dlcName, raceAlternativeName, variantName, fileAlternativeName)

								if _, err := os.Stat(languageFile); err == nil {
									wg.Add(1)
									semaphore <- struct{}{}

									go func(src, dst string) {
										defer wg.Done()
										defer func() { <-semaphore }()

										if err := copyFile(src, dst); err != nil {
											logAndPrint(fmt.Sprintf("ERROR: Could not copy %s to %s: %v", src, dst, err), logger)
										}
										atomic.AddInt32(&processedFiles, 1)
									}(languageFile, remasterFile)
								}
							}
						}
					}
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

func copyFile(src, dst string) error {
	info, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return fmt.Errorf("ERROR: Target path %s does not exist", dst)
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("ERROR: Target path %s does not include file", dst)
	}

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