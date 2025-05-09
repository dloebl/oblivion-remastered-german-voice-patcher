package main

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	//	"fmt"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

var VERSION = "v0.4.3"

// Globale Variable für übersprungene BNK-Dateien
var totalSkippedBnks int32

// Neue globale Variable für erstellte BNK-Dateien
var totalCreatedBnks int32

// Verzeichnisse erstellen, die im Shell-Script definiert sind
func createOutputDirs() {
	dirs := []string{
		"german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Event/English(US)/",
		"german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Media/English(US)/",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("Warnung: Konnte Verzeichnis nicht erstellen: %s (err: %v)", dir, err)
		}
	}
}

func create_bnk(bnkName string, bnkPath string, bnk []byte, wemPath string, isVideo string) {
	pattern := []byte{0x01, 0x00, 0x14, 0x00} // Codec: OPUS_WEM
	if isVideo == "true" {
		pattern = []byte{0x01, 0x00, 0x01, 0x00} // Codec: PCM
	}
	newCodec := []byte{0x01, 0x00, 0x04, 0x00} // Codec: VORBIS
	// Find the pattern in the file:
	// Quick and dirty approach to patch the BNKs
	pos := bytes.Index(bnk, pattern)
	if pos == -1 {
		pos = bytes.Index(bnk, newCodec)
		if pos == -1 {
			log.Fatalf("Pattern not found")
		}
	}
	// Read the values that we need
	//codec := bnk[pos : pos+4]
	//dummy := bnk[pos+4]
	id := binary.LittleEndian.Uint32(bnk[pos+5 : pos+9])
	//fileSize := binary.LittleEndian.Uint32(bnk[pos+9 : pos+13])

	/* debug output
	fmt.Printf("Codec:      %02X %02X %02X %02X\n", codec[0], codec[1], codec[2], codec[3])
	fmt.Printf("Dummy:      %02X\n", dummy)
	fmt.Printf("ID:         %d\n", id)
	fmt.Printf("File Size:  %d bytes\n", fileSize)
	*/

	// Get size of the .wem file
	wemInfo, err := os.Stat(wemPath + ".wem")
	if err != nil {
		log.Fatalf("Failed to read .wem file: %v\n", err)
	}
	wemSize := uint32(wemInfo.Size())
	// Update the codex to VORBIS
	copy(bnk[pos:pos+4], newCodec)
	// Update file size (4 bytes after dummy byte and ID)
	fileSizeOffset := pos + 9
	if fileSizeOffset+4 > len(bnk) {
		log.Fatalf("Not enough data to update file size in .bnk")
	}
	binary.LittleEndian.PutUint32(bnk[fileSizeOffset:fileSizeOffset+4], wemSize)
	// write the modified .bnk file to the output folder
	outBnkPath := "german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Event/English(US)/" + filepath.Base(bnkPath)
	if isVideo == "true" {
		outBnkPath = "german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Event/" + filepath.Base(bnkPath)
	}
	err = os.WriteFile(outBnkPath, bnk, 0644)
	if err != nil {
		log.Fatalf("Failed to write modified .bnk file: %v\n", err)
	}
	//fmt.Printf("Modified .bnk file written to: %s\n", outBnkPath)
	// write the .wem file to output folder
	wem, err := os.ReadFile(wemPath + ".wem")
	if err != nil {
		log.Fatalf("Failed to read .wem file: %v\n", err)
	}
	outWemPath := "german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Media/English(US)/" + strconv.Itoa(int(id)) + ".wem"
	if isVideo == "true" {
		outWemPath = "german-voices-oblivion-remastered-voxmeld_" + VERSION + "_P/Content/WwiseAudio/Media/" + strconv.Itoa(int(id)) + ".wem"
	}
	err = os.WriteFile(outWemPath, wem, 0644)
	if err != nil {
		log.Fatalf("Failed to write .wem file: %v\n", err)
	}

	// Zähler für erstellte BNK-Dateien erhöhen
	atomic.AddInt32(&totalCreatedBnks, 1)
}

// TODO: Bekommt eine WEM datei und schaut ob deren Name zum zusammenstellen eines pfade zu einer BNK Datei benutzt werden kann.
func processWemFile(wemPath string) {
	comps := strings.Split(filepath.Base(wemPath), "_")

	// Zähler für übersprungene BNK-Dateien
	skippedBnks := 0

	// File is audio for video
	if comps[0] == "scripted" {
		bnkName := filepath.Base(wemPath)

		bnkPath := "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/" + bnkName + ".bnk"
		// Log-Ausgabe entfernen
		// log.Println("File:", bnkPath)

		bnk, err := os.ReadFile(bnkPath)
		if err != nil {
			// Log-Ausgabe entfernen
			// log.Printf("Skipping missing BNK: %s (err: %v)", bnkPath, err)
			skippedBnks++
			return
		}

		create_bnk(bnkName, bnkPath, bnk, wemPath, "true")
	} else {
		raceComb := comps[0]

		// Case for high_elf, dark_seducer and holy_saint as they have an underscore in their name
		if comps[1] != "f" && comps[1] != "m" {
			raceComb += "_" + comps[1]
		}

		var races []string
		var variants []string
		races = append(races, raceComb)
		switch raceComb {
		case "argonian":
			races = append(races, "khajiit")
			break
		case "high_elf":
			races = append(races, "dark_elf")
			races = append(races, "wood_elf")
			break
		case "imperial":
			races = append(races, "breton")
			break
		case "nord":
			races = append(races, "orc")
			break
		}
		variants = append(variants, "")
		variants = append(variants, "altvoice")
		variants = append(variants, "beggar")

		for _, race := range races {
			for _, variant := range variants {
				variantComp := comps[1]
				restComp := strings.Join(comps[2:], "_")
				// Case for high_elf, dark_seducer and holy_saint as they have an underscore in their name
				if comps[1] != "f" && comps[1] != "m" {
					variantComp = comps[2]
					restComp = strings.Join(comps[3:], "_")
				}

				bnkName := race + "_" + variantComp + "_"
				if variant != "" {
					bnkName += variant + "_"
				}
				bnkName += restComp

				bnkPath := "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + ".bnk"
				// Log-Ausgabe entfernen
				// log.Println("File:", bnkPath)

				bnk, err := os.ReadFile(bnkPath)
				if err != nil {
					bnkPath = "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + "_sid.bnk"
					// Log-Ausgabe entfernen
					// log.Printf("Skipping missing BNK: %s (err: %v)", bnkPath, err)
					bnk, err = os.ReadFile(bnkPath)
					if err != nil {
						skippedBnks++
						continue
					}
				}
				create_bnk(bnkName, bnkPath, bnk, wemPath, "false")
			}
		}
	}

	// Globalen Zähler für übersprungene BNKs aktualisieren
	if skippedBnks > 0 {
		atomic.AddInt32(&totalSkippedBnks, int32(skippedBnks))
	}
}

// zeichneAnimiertenFortschrittsbalken stellt einen animierten Fortschrittsbalken in der Konsole dar
func zeichneAnimiertenFortschrittsbalken(aktuellerFortschritt, gesamtAnzahl int, startZeit time.Time, animationsZähler int, skippedBnks int32) {
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

	// Lösche die aktuelle Zeile und zeige den Balken an
	fmt.Printf("\r[%s] %3.0f%% %d/%d WEMs | Erstellt: %d BNKs| Übersprungene: %d BNKs",
		balken, prozent*100, aktuellerFortschritt, gesamtAnzahl,
		atomic.LoadInt32(&totalCreatedBnks), skippedBnks)
}

func main() {
	// Startzeit erfassen
	startZeit := time.Now()

	// Gib das aktuelle Arbeitsverzeichnis aus
	aktuellesVerzeichnis, err := os.Getwd()
	if err == nil {
		log.Printf("Arbeitsverzeichnis: %s", aktuellesVerzeichnis)
	}

	// Erstelle die Ausgabeverzeichnisse
	createOutputDirs()

	// Wenn ein Argument übergeben wurde, verarbeite nur diese eine Datei
	if len(os.Args) > 1 {
		wemPath := os.Args[1]
		processWemFile(wemPath)

		// Zeige Gesamtzeit an
		gesamtZeit := time.Since(startZeit)
		fmt.Printf("\n\nVerarbeitung abgeschlossen in %s\n", gesamtZeit)
		fmt.Printf("Erstellte BNK-Dateien: %d\n", atomic.LoadInt32(&totalCreatedBnks))
		fmt.Printf("Übersprungene BNK-Dateien: %d\n", atomic.LoadInt32(&totalSkippedBnks))
		return
	}

	// Prüfe, ob das Verzeichnis existiert, bevor es durchsucht wird
	wemDir := "sound2wem/Windows/"
	if _, err := os.Stat(wemDir); os.IsNotExist(err) {
		log.Fatalf("Fehler: Verzeichnis %s existiert nicht im aktuellen Arbeitsverzeichnis %s",
			wemDir, aktuellesVerzeichnis)
	}

	// Andernfalls durchsuche das Verzeichnis nach allen WEM-Dateien und verarbeite sie parallel
	var wemFiles []string
	err = filepath.WalkDir(wemDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".wem") {
			// Entferne die .wem-Endung, wie im Shell-Script
			wemFiles = append(wemFiles, strings.TrimSuffix(path, ".wem"))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Fehler beim Durchsuchen des Verzeichnisses: %v", err)
	}

	// Zeige Step an.
	fmt.Printf("\n====================== VOXMELD ======================\n")
	fmt.Printf("Quelle:      %s\n", wemDir)
	fmt.Printf("Dateien:     %d WEM-Dateien gefunden\n", len(wemFiles))
	fmt.Printf("Status:      Starte Konvertierung von WEM zu BNK\n")
	fmt.Printf("-----------------------------------------------------\n")

	// Fortschrittsbalken-Variablen
	var erledigt int32
	var fortschrittsMutex sync.Mutex
	gesamtAnzahl := len(wemFiles)
	var animationsZähler int

	// Starte Animation im Hintergrund
	animationsStopp := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fortschrittsMutex.Lock()
				animationsZähler++
				zeichneAnimiertenFortschrittsbalken(
					int(atomic.LoadInt32(&erledigt)),
					gesamtAnzahl,
					startZeit,
					animationsZähler,
					atomic.LoadInt32(&totalSkippedBnks))
				fortschrittsMutex.Unlock()
			case <-animationsStopp:
				return
			}
		}
	}()

	// Parallelisierung mit Worker-Pool
	numWorkers := runtime.NumCPU() // Statt fester Anzahl 64 die Anzahl der CPU-Kerne verwenden
	fmt.Printf("Verwende %d Worker-Threads (basierend auf CPU-Kernen)\n", numWorkers)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, numWorkers)

	for _, wemPath := range wemFiles {
		wg.Add(1)
		semaphore <- struct{}{} // Blockiert, wenn alle Worker beschäftigt sind
		go func(path string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Worker freigeben
			processWemFile(path)

			// Aktualisiere Fortschrittsbalken
			atomic.AddInt32(&erledigt, 1)
		}(wemPath)
	}

	wg.Wait() // Warte auf Abschluss aller Worker

	// Animationsschleife stoppen
	close(animationsStopp)
	time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

	// Zeige finalen Fortschrittsbalken
	fortschrittsMutex.Lock()
	zeichneAnimiertenFortschrittsbalken(
		gesamtAnzahl,
		gesamtAnzahl,
		startZeit,
		animationsZähler,
		atomic.LoadInt32(&totalSkippedBnks))
	fortschrittsMutex.Unlock()

	// Berechne die Gesamtzeit
	gesamtZeit := time.Since(startZeit)
	fmt.Printf("\n\nAlle Dateien verarbeitet in %s\n", gesamtZeit)
	fmt.Printf("Erstellte BNK-Dateien: %d\n", atomic.LoadInt32(&totalCreatedBnks))
	fmt.Printf("Übersprungene BNK-Dateien: %d\n", atomic.LoadInt32(&totalSkippedBnks))
}