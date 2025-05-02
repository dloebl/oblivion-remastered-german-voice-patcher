package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"log"
)

// Neuer Logger
var logger *log.Logger
var logFile *os.File

// Mapping für die Umbenennungen
var prefixMappings = map[string]struct {
	oldPath string
	newPath string
}{
	"oblivion.esm": {
		"tmp/sound/voice/oblivion.esm/argonier",
		"tmp/sound/voice/oblivion.esm/argonian",
	},
	"oblivion.esm_hochelf": {
		"tmp/sound/voice/oblivion.esm/hochelf",
		"tmp/sound/voice/oblivion.esm/high_elf",
	},
	"oblivion.esm_kaiserlicher": {
		"tmp/sound/voice/oblivion.esm/kaiserlicher",
		"tmp/sound/voice/oblivion.esm/imperial",
	},
	"oblivion.esm_rothwardone": {
		"tmp/sound/voice/oblivion.esm/rothwardone",
		"tmp/sound/voice/oblivion.esm/redguard",
	},
	"knights.esp": {
		"tmp/sound/voice/knights.esp/argonier",
		"tmp/sound/voice/knights.esp/argonian",
	},
	"knights.esp_hochelf": {
		"tmp/sound/voice/knights.esp/hochelf",
		"tmp/sound/voice/knights.esp/high_elf",
	},
	"knights.esp_kaiserlicher": {
		"tmp/sound/voice/knights.esp/kaiserlicher",
		"tmp/sound/voice/knights.esp/imperial",
	},
	"knights.esp_rothwardone": {
		"tmp/sound/voice/knights.esp/rothwardone",
		"tmp/sound/voice/knights.esp/redguard",
	},
	"oblivion.esm_dunkler": {
		"tmp/sound/voice/oblivion.esm/dunkler verführer",
		"tmp/sound/voice/oblivion.esm/dark_seducer",
	},
	"oblivion.esm_goldener": {
		"tmp/sound/voice/oblivion.esm/goldener heiliger",
		"tmp/sound/voice/oblivion.esm/golden_saint",
	},
}

// Mapping für alternative Rassen
var raceAlternatives = map[string][]string{
	"argonian": {"khajiit"},
	"high_elf": {"dark_elf", "wood_elf"},
	"imperial": {"breton"},
	"nord":     {"orc"},
}

// Füge einen Mutex für Dateioperationen hinzu
var fileMutex sync.Mutex

func main() {
	setupLogging()
	defer logFile.Close()

	logAndPrint("Starte Verarbeitung der Audiodateien...")
	
	// Führe Prefix-Änderungen durch
	logAndPrint("Führe Prefix-Änderungen durch...")
	performPrefixChanges()

	// Erstelle MP3s Verzeichnis
	logAndPrint("Erstelle MP3s Verzeichnis...")
	os.MkdirAll("tmp/MP3s", 0755)

	// Verarbeite Dateien
	logAndPrint("Starte Dateiverarbeitung...")
	processFiles()

	logAndPrint("\nVerarbeitung abgeschlossen!")
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
	logFile, err = os.OpenFile("logs/change-prefix-move-mp3s.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Log-Datei: %v\n", err)
		return
	}

	// Initialisiere den Logger
	logger = log.New(logFile, "", log.LstdFlags)
	logAndPrint("Logging initialisiert")
}

// Funktion zum Loggen und gleichzeitigen Ausgeben einer Nachricht
func logAndPrint(message string) {
	fmt.Println(message)
	if logger != nil {
		logger.Println(message)
	}
}

func performPrefixChanges() {
	for _, mapping := range prefixMappings {
		// Prüfe, ob der Quellpfad existiert
		if _, err := os.Stat(mapping.oldPath); err == nil {
			logAndPrint(fmt.Sprintf("Benenne um: %s -> %s", mapping.oldPath, mapping.newPath))
			if err := os.Rename(mapping.oldPath, mapping.newPath); err != nil {
				logAndPrint(fmt.Sprintf("Fehler beim Umbenennen von %s zu %s: %v", mapping.oldPath, mapping.newPath, err))
			}
		} else {
			logAndPrint(fmt.Sprintf("Quellpfad existiert nicht: %s", mapping.oldPath))
		}
	}
}

func checkAndCopyRemaster(dlc, race, variant, file string, wg *sync.WaitGroup) {
	prefixes := []string{"", "/altvoice", "/beggar"}
	for _, prefix := range prefixes {
		race = strings.Replace(race, "_", " ", -1)
		targetPath := filepath.Join(
			"ModFiles/Content/Dev/ObvData/Data/sound/voice",
			dlc, race, variant+prefix, filepath.Base(file),
		)
		
		// Prüfe ob Zieldatei existiert
		if _, err := os.Stat(filepath.Dir(targetPath)); err == nil {
			message := fmt.Sprintf("Copy variant: %s/%s/%s%s/%s...", dlc, race, variant, prefix, filepath.Base(file))
			logAndPrint(message)
			wg.Add(1)
			go func(src, dst string) {
				defer wg.Done()
				if err := copyFile(src, dst); err != nil {
					logAndPrint(fmt.Sprintf("Fehler beim Kopieren von %s nach %s: %v", src, dst, err))
				}
			}(file, targetPath)
		}
	}
}

func processFiles() {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32
	var mu sync.Mutex

	// Zähle zuerst alle Dateien
	dlcs, _ := filepath.Glob("tmp/sound/voice/*")
	for _, dlc := range dlcs {
		races, _ := filepath.Glob(filepath.Join(dlc, "*"))
		for _, race := range races {
			variants, _ := filepath.Glob(filepath.Join(race, "*"))
			for _, variant := range variants {
				files, _ := filepath.Glob(filepath.Join(variant, "*.mp3"))
				totalFiles += len(files)
			}
		}
	}

	// Funktion zum Aktualisieren des Fortschritts
	updateProgress := func() {
		mu.Lock()
		current := atomic.AddInt32(&processedFiles, 1)
		percentage := float64(current) / float64(totalFiles) * 100
		fmt.Printf("\rFortschritt: %.2f%% (%d/%d Dateien)", percentage, current, totalFiles)
		mu.Unlock()
	}

	// Verarbeite die Dateien
	for _, dlc := range dlcs {
		races, _ := filepath.Glob(filepath.Join(dlc, "*"))
		for _, race := range races {
			variants, _ := filepath.Glob(filepath.Join(race, "*"))
			for _, variant := range variants {
				files, _ := filepath.Glob(filepath.Join(variant, "*.mp3"))
				for _, file := range files {
					raceName := filepath.Base(race)
					variantName := filepath.Base(variant)
					dlcName := filepath.Base(dlc)

					// Kopiere Datei in den MP3s-Ordner
					mp3Target := filepath.Join("tmp/MP3s", fmt.Sprintf("%s_%s_%s", raceName, variantName, filepath.Base(file)))
					wg.Add(1)
					go func(src, dst string) {
						defer wg.Done()
						if err := copyFile(src, dst); err != nil {
							logAndPrint(fmt.Sprintf("Fehler beim Kopieren von %s nach %s: %v", src, dst, err))
						}
						updateProgress()
					}(file, mp3Target)

					// Prüfe Varianten und kopiere in BSA-Extraktordner
					checkAndCopyRemaster(dlcName, raceName, variantName, file, &wg)

					// Prüfe alternative Rassen
					if alternatives, ok := raceAlternatives[raceName]; ok {
						for _, altRace := range alternatives {
							checkAndCopyRemaster(dlcName, altRace, variantName, file, &wg)
						}
					}
				}
			}
		}
	}

	wg.Wait()
	logAndPrint("\nVerarbeitung abgeschlossen!")
}

func copyFile(src, dst string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

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