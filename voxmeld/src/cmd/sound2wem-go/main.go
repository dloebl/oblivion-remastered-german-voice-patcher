package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	WwisePath   string `json:"wwisePath"`
	FfmpegPath  string `json:"ffmpegPath"`
	ProjectName string `json:"projectName"`
	Conversion  string `json:"conversion"`
}

type ExternalSourcesList struct {
	XMLName       xml.Name `xml:"ExternalSourcesList"`
	SchemaVersion string   `xml:"SchemaVersion,attr"`
	Root          string   `xml:"Root,attr"`
	Sources       []Source `xml:"Source"`
}

type Source struct {
	Path       string `xml:"Path,attr"`
	Conversion string `xml:"Conversion,attr"`
}

// Globale Variablen für Fortschrittsanzeige
var verarbeiteteAudios int32
var gesamtAnzahlAudios int32

func loadConfig(execDir string) (*Config, error) {
	configPath := filepath.Join(execDir, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Lesen der config.json: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("Fehler beim Parsen der config.json: %v", err)
	}

	// Setze Standardwerte falls leer
	if config.WwisePath == "" {
		config.WwisePath = os.Getenv("WWISEROOT") + "\\Authoring\\x64\\Release\\bin\\WwiseConsole.exe"
	}
	if config.FfmpegPath == "" {
		config.FfmpegPath = filepath.Join(execDir, "ffmpeg-master-latest-win64-gpl-shared", "bin", "ffmpeg.exe")
	}
	if config.ProjectName == "" {
		config.ProjectName = "wavtowemscript"
	}
	if config.Conversion == "" {
		config.Conversion = "Vorbis Quality High"
	}

	return &config, nil
}

// Funktion zum Ausgeben einer Nachricht
func printMessage(message string) {
	fmt.Println(message)
}

// zeichneAnimiertenFortschrittsbalken stellt einen animierten Fortschrittsbalken in der Konsole dar
func zeichneAnimiertenFortschrittsbalken(aktuellerFortschritt, gesamtAnzahl int, startZeit time.Time, animationsZähler int) {
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
	fmt.Printf("\r[%s] %3.0f%% %d/%d Dateien verarbeitet",
		balken, prozent*100, aktuellerFortschritt, gesamtAnzahl)
}

func main() {
	// Startzeit erfassen
	startZeit := time.Now()

	if len(os.Args) < 2 {
		printMessage("Fehler: Keine Eingabedateien angegeben")
		return
	}

	// Konfiguration laden
	execDir, err := os.Executable()
	if err != nil {
		printMessage(fmt.Sprintf("Fehler beim Ermitteln des Ausführungsverzeichnisses: %v", err))
		return
	}
	execDir = filepath.Dir(execDir)

	config, err := loadConfig(execDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Wwise Projekt erstellen, falls es nicht existiert
	projectPath := filepath.Join(execDir, config.ProjectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		printMessage("Erstelle neues Wwise Projekt...")
		cmd := exec.Command(config.WwisePath, "create-new-project",
			filepath.Join(projectPath, config.ProjectName+".wproj"),
			"--quiet")
		if err := cmd.Run(); err != nil {
			printMessage(fmt.Sprintf("Fehler beim Erstellen des Wwise Projekts: %v", err))
			return
		}
	}

	// Temporäres Verzeichnis erstellen
	tempDir := filepath.Join(execDir, "audiotemp")
	os.MkdirAll(tempDir, 0755)
	// defer os.RemoveAll(tempDir)

	// Anzahl der zu verarbeitenden Dateien bestimmen
	var gesamtAnzahl int
	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		gesamtAnzahl += len(matches)
	}
	atomic.StoreInt32(&gesamtAnzahlAudios, int32(gesamtAnzahl))

	// Zeige Step an.
	fmt.Printf("\n====================== SOUND2WEM ======================\n")
	fmt.Printf("Quelle:      %s\n", strings.Join(os.Args[1:], ", "))
	fmt.Printf("Dateien:     %d Audio-Dateien gefunden\n", gesamtAnzahl)
	fmt.Printf("Status:      Starte Konvertierung von Audio zu WEM\n")
	fmt.Printf("-------------------------------------------------------\n")

	// Fortschrittsbalken-Variablen
	var fortschrittsMutex sync.Mutex
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
					int(atomic.LoadInt32(&verarbeiteteAudios)),
					int(atomic.LoadInt32(&gesamtAnzahlAudios)),
					startZeit,
					animationsZähler)
				fortschrittsMutex.Unlock()
			case <-animationsStopp:
				return
			}
		}
	}()

	// Parallel Audio-Dateien konvertieren
	var wg sync.WaitGroup
	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)

	fmt.Printf("Starte Konvertierung mit %d parallelen Prozessen...\n", numCPU)

	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		for _, file := range matches {
			wg.Add(1)
			go func(inputFile string) {
				defer wg.Done()
				semaphore <- struct{}{}        // Slot belegen
				defer func() { <-semaphore }() // Slot freigeben

				outputFile := filepath.Join(tempDir, filepath.Base(inputFile))
				outputFile = outputFile[:len(outputFile)-len(filepath.Ext(outputFile))] + ".wav"

				cmd := exec.Command(config.FfmpegPath, "-hide_banner", "-loglevel", "warning",
					"-i", inputFile, outputFile)
				if err := cmd.Run(); err != nil {
					// Reduziertes Logging
					fmt.Printf("\nFehler bei der Konvertierung von %s: %v\n", inputFile, err)
				}

				// Fortschritt aktualisieren
				atomic.AddInt32(&verarbeiteteAudios, 1)
			}(file)
		}
	}
	wg.Wait()

	// Animationsschleife stoppen
	close(animationsStopp)
	time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

	// Zeige finalen Fortschrittsbalken
	fortschrittsMutex.Lock()
	zeichneAnimiertenFortschrittsbalken(
		int(atomic.LoadInt32(&verarbeiteteAudios)),
		int(atomic.LoadInt32(&gesamtAnzahlAudios)),
		startZeit,
		animationsZähler)
	fortschrittsMutex.Unlock()

	fmt.Println("\n\nAlle Audio-Dateien konvertiert. Erstelle XML...")

	// WSources XML erstellen
	sources := ExternalSourcesList{
		SchemaVersion: "1",
		Root:          tempDir,
	}

	files, _ := filepath.Glob(filepath.Join(tempDir, "*.wav"))
	for _, file := range files {
		sources.Sources = append(sources.Sources, Source{
			Path:       filepath.Base(file),
			Conversion: config.Conversion,
		})
	}

	// XML speichern
	wsourcesPath := filepath.Join(execDir, "list.wsources")
	xmlData, err := xml.MarshalIndent(sources, "", "  ")
	if err != nil {
		printMessage(fmt.Sprintf("Fehler beim Erstellen der XML-Daten: %v", err))
		os.Exit(1)
	}
	os.WriteFile(wsourcesPath, []byte(xml.Header+string(xmlData)), 0644)
	defer os.Remove(wsourcesPath)

	printMessage("Starte Wwise Konvertierung...")

	// Wwise Konvertierung
	cmd := exec.Command(config.WwisePath, "convert-external-source",
		filepath.Join(execDir, config.ProjectName, config.ProjectName+".wproj"),
		"--source-file", wsourcesPath,
		"--output", execDir,
		"--quiet")

	// Pipe für die Standardausgabe erstellen
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		printMessage(fmt.Sprintf("Fehler beim Erstellen der Stdout-Pipe: %v", err))
		return
	}

	// Pipe für die Fehlerausgabe erstellen
	stderr, err := cmd.StderrPipe()
	if err != nil {
		printMessage(fmt.Sprintf("Fehler beim Erstellen der Stderr-Pipe: %v", err))
		return
	}

	// Kommando im Hintergrund starten
	if err := cmd.Start(); err != nil {
		printMessage(fmt.Sprintf("Fehler beim Starten der Wwise Konvertierung: %v", err))
		return
	}

	// Animation für den Fortschrittsindikator
	done := make(chan bool)
	go func() {
		spinner := []string{"-", "\\", "|", "/"}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\rKonvertierung läuft... %s", spinner[i])
				i = (i + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Ausgaben in Echtzeit verarbeiten
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("\r%s\n", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("\r%s\n", scanner.Text())
		}
	}()

	// Auf Beendigung warten
	if err := cmd.Wait(); err != nil {
		done <- true
		printMessage(fmt.Sprintf("\n\rFehler bei der Wwise Konvertierung: %v", err))
	} else {
		done <- true
		printMessage(fmt.Sprintf("\n\rWwise Konvertierung erfolgreich abgeschlossen!"))
	}

	// Berechne die Gesamtzeit
	gesamtZeit := time.Since(startZeit)
	fmt.Printf("\nGesamtzeit: %s\n", gesamtZeit)
}
