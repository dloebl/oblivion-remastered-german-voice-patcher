package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Neuer Logger
var logger *log.Logger
var logFile *os.File

// Mapping für die Umbenennungen
var prefixMappings = map[string]struct {
	oldPath string
	newPath string
}{
	"oblivion.esm_argonier": {
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
	"oblivion.esm_dunkler": {
		"tmp/sound/voice/oblivion.esm/dunkler verführer",
		"tmp/sound/voice/oblivion.esm/dark_seducer",
	},
	"oblivion.esm_goldener": {
		"tmp/sound/voice/oblivion.esm/goldener heiliger",
		"tmp/sound/voice/oblivion.esm/golden_saint",
	},
	"knights.esp_argonier": {
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
	"dlcvilelair.esp_hochelf": {
		"tmp/sound/voice/dlcvilelair.esp/hochelf",
		"tmp/sound/voice/dlcvilelair.esp/high_elf",
	},
	"dlcvilelair.esp_kaiserlicher": {
		"tmp/sound/voice/dlcvilelair.esp/kaiserlicher",
		"tmp/sound/voice/dlcvilelair.esp/imperial",
	},
	"dlcthievesden.esp_argonier": {
		"tmp/sound/voice/dlcthievesden.esp/argonier",
		"tmp/sound/voice/dlcthievesden.esp/argonian",
	},
	"dlcthievesden.esp_hochelf": {
		"tmp/sound/voice/dlcthievesden.esp/hochelf",
		"tmp/sound/voice/dlcthievesden.esp/high_elf",
	},
	"dlcthievesden.esp_kaiserlicher": {
		"tmp/sound/voice/dlcthievesden.esp/kaiserlicher",
		"tmp/sound/voice/dlcthievesden.esp/imperial",
	},
	"dlcthievesden.esp_rothwardone": {
		"tmp/sound/voice/dlcthievesden.esp/rothwardone",
		"tmp/sound/voice/dlcthievesden.esp/redguard",
	},
	"dlcorrery.esp_hochelf": {
		"tmp/sound/voice/dlcorrery.esp/hochelf",
		"tmp/sound/voice/dlcorrery.esp/high_elf",
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

// Globale Variablen für Fortschrittsanzeige
var animationsStopp chan struct{}
var animationsZähler int
var fortschrittsMutex sync.Mutex

func main() {
	setupLogging()
	defer logFile.Close()

	// Startzeit erfassen
	startZeit := time.Now()

	// Zeige Header an
	fmt.Printf("\n====================== AUDIO DATEI VERARBEITUNG ======================\n")
	fmt.Printf("Status:      Starte Verarbeitung der Audiodateien\n")
	fmt.Printf("-------------------------------------------------------------------\n")

	// Führe Prefix-Änderungen durch
	fmt.Println("Führe Prefix-Änderungen durch...")
	performPrefixChanges()

	// Erstelle MP3s Verzeichnis
	fmt.Println("Erstelle MP3s Verzeichnis...")
	os.MkdirAll("tmp/MP3s", 0755)

	// Verarbeite Dateien
	fmt.Println("Starte Dateiverarbeitung...")
	processFiles(startZeit)

	fmt.Println("\nVerarbeitung abgeschlossen!")
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
	fmt.Println("Logging initialisiert")
}

// Funktion zum Loggen und gleichzeitigen Ausgeben einer Nachricht
func logAndPrint(message string) {
	// Temporär deaktiviert: Keine Ausgabe für "Copy variant:"
	if !strings.Contains(message, "Copy variant:") {
		fmt.Println(message)
		if logger != nil {
			logger.Println(message)
		}
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
		if _, err := os.Stat(targetPath); err == nil {
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

// zeichneAnimiertenFortschrittsbalken stellt einen animierten Fortschrittsbalken in der Konsole dar
func zeichneAnimiertenFortschrittsbalken(aktuellerFortschritt, gesamtAnzahl int, startZeit time.Time, animationsZähler int, aktuelleAktion string) {
	breite := 40 // Breite des Balkens in Zeichen

	// Berechne Prozentsatz
	prozent := float64(aktuellerFortschritt) / float64(gesamtAnzahl)

	// Berechne Anzahl der gefüllten Zeichen
	gefüllt := int(prozent * float64(breite))

	// Animations-Zeichen
	animationsSymbole := []string{"|", "/", "-", "\\"}
	animationSymbol := animationsSymbole[animationsZähler%len(animationsSymbole)]

	// ASCII-Ladebalken Zeichen
	gefülltZeichen := "#"
	leerZeichen := "-"

	// Erstelle den Ladebalken
	balken := strings.Repeat(gefülltZeichen, gefüllt) + strings.Repeat(leerZeichen, breite-gefüllt)

	// Erstelle einen eingebetteten Animations-Cursor im Ladebalken
	if gefüllt < breite {
		position := gefüllt
		balkenRunes := []rune(balken)
		balkenRunes[position] = []rune(animationSymbol)[0]
		balken = string(balkenRunes)
	}

	// Lösche die aktuelle Zeile und zeige nur den Balken ohne verstrichene Zeit an
	fmt.Printf("\r[%s] %3.0f%% %d/%d Dateien verarbeitet",
		balken, prozent*100, aktuellerFortschritt, gesamtAnzahl)
}

func processFiles(startZeit time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32
	var aktuelleAktion string = "Zähle Dateien"

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

	// Aktualisiere Aktion nach dem Zählen
	aktuelleAktion = "Kopiere MP3-Dateien"

	// Zeige Informationen
	fmt.Printf("\n-------------------------------------------------------------------\n")
	fmt.Printf("Dateien:     %d Audio-Dateien gefunden\n", totalFiles)
	fmt.Printf("Status:      Starte Kopieren und Umbenennen der Dateien\n")
	fmt.Printf("-------------------------------------------------------------------\n")

	// Starte Animation im Hintergrund
	animationsStopp = make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fortschrittsMutex.Lock()
				animationsZähler++
				zeichneAnimiertenFortschrittsbalken(
					int(atomic.LoadInt32(&processedFiles)),
					totalFiles,
					startZeit,
					animationsZähler,
					aktuelleAktion)
				fortschrittsMutex.Unlock()
			case <-animationsStopp:
				return
			}
		}
	}()

	// Verarbeite die Dateien
	for _, dlc := range dlcs {
		dlcName := filepath.Base(dlc)
		aktuelleAktion = fmt.Sprintf("Kopiere Dateien aus: %s", dlcName)

		races, _ := filepath.Glob(filepath.Join(dlc, "*"))
		for _, race := range races {
			raceName := filepath.Base(race)
			aktuelleAktion = fmt.Sprintf("Kopiere Dateien: %s/%s", dlcName, raceName)

			variants, _ := filepath.Glob(filepath.Join(race, "*"))
			for _, variant := range variants {
				variantName := filepath.Base(variant)
				aktuelleAktion = fmt.Sprintf("Kopiere: %s/%s/%s", dlcName, raceName, variantName)

				files, _ := filepath.Glob(filepath.Join(variant, "*.mp3"))
				for _, file := range files {
					// Kopiere Datei in den MP3s-Ordner
					mp3Target := filepath.Join("tmp/MP3s", fmt.Sprintf("%s_%s_%s", raceName, variantName, filepath.Base(file)))
					wg.Add(1)
					go func(src, dst string) {
						defer wg.Done()
						if err := copyFile(src, dst); err != nil {
							logAndPrint(fmt.Sprintf("Fehler beim Kopieren von %s nach %s: %v", src, dst, err))
						}
						atomic.AddInt32(&processedFiles, 1)
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

	// Aktualisiere Aktion nach der Verarbeitung
	aktuelleAktion = "Dateien erfolgreich kopiert"

	// Animationsschleife stoppen
	close(animationsStopp)
	time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

	// Zeige finalen Fortschrittsbalken
	fortschrittsMutex.Lock()
	zeichneAnimiertenFortschrittsbalken(
		int(atomic.LoadInt32(&processedFiles)),
		totalFiles,
		startZeit,
		animationsZähler,
		aktuelleAktion)
	fortschrittsMutex.Unlock()
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
