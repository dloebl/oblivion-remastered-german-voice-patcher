package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Neuer Logger
var logger *log.Logger
var logFile *os.File

type Config struct {
	WwisePath   string `json:"wwisePath"`
	FfmpegPath  string `json:"ffmpegPath"`
	ProjectName string `json:"projectName"`
	Conversion  string `json:"conversion"`
}

type ExternalSourcesList struct {
	XMLName       xml.Name `xml:"ExternalSourcesList"`
	SchemaVersion string   `xml:"SchemaVersion,attr"`
	Root         string   `xml:"Root,attr"`
	Sources      []Source `xml:"Source"`
}

type Source struct {
	Path       string `xml:"Path,attr"`
	Conversion string `xml:"Conversion,attr"`
}

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

// Funktion zum Einrichten des Loggings
func setupLogging() {
	// Erstelle Log-Verzeichnis, wenn es nicht existiert
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen des Log-Verzeichnisses: %v\n", err)
		return
	}

	// Öffne die Log-Datei
	logFile, err = os.OpenFile("logs/sound2wem.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

func main() {
	// Logging einrichten
	setupLogging()
	defer logFile.Close()

	if len(os.Args) < 2 {
		logAndPrint("Fehler: Keine Eingabedateien angegeben")
		return
	}

	// Konfiguration laden
	execDir, err := os.Executable()
	if err != nil {
		logAndPrint(fmt.Sprintf("Fehler beim Ermitteln des Ausführungsverzeichnisses: %v", err))
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
		logAndPrint("Erstelle neues Wwise Projekt...")
		cmd := exec.Command(config.WwisePath, "create-new-project",
			filepath.Join(projectPath, config.ProjectName+".wproj"),
			"--quiet")
		if err := cmd.Run(); err != nil {
			logAndPrint(fmt.Sprintf("Fehler beim Erstellen des Wwise Projekts: %v", err))
			return
		}
	}

	// Temporäres Verzeichnis erstellen
	tempDir := filepath.Join(execDir, "audiotemp")
	os.MkdirAll(tempDir, 0755)
	// defer os.RemoveAll(tempDir)

	// Parallel Audio-Dateien konvertieren
	var wg sync.WaitGroup
	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)

	logAndPrint(fmt.Sprintf("Starte Konvertierung mit %d parallelen Prozessen...", numCPU))

	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		for _, file := range matches {
			wg.Add(1)
			go func(inputFile string) {
				defer wg.Done()
				semaphore <- struct{}{} // Slot belegen
				defer func() { <-semaphore }() // Slot freigeben

				outputFile := filepath.Join(tempDir, filepath.Base(inputFile))
				outputFile = outputFile[:len(outputFile)-len(filepath.Ext(outputFile))] + ".wav"
				
				logAndPrint(fmt.Sprintf("Konvertiere: %s -> %s", inputFile, filepath.Base(outputFile)))
				
				cmd := exec.Command(config.FfmpegPath, "-hide_banner", "-loglevel", "warning", 
					"-i", inputFile, outputFile)
				if err := cmd.Run(); err != nil {
					logAndPrint(fmt.Sprintf("Fehler bei der Konvertierung von %s: %v", inputFile, err))
				} else {
					logAndPrint(fmt.Sprintf("Erfolgreich konvertiert: %s", filepath.Base(outputFile)))
				}
			}(file)
		}
	}
	wg.Wait()

	logAndPrint("Alle Audio-Dateien konvertiert. Erstelle XML...")

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
		logAndPrint(fmt.Sprintf("Fehler beim Erstellen der XML-Daten: %v", err))
		os.Exit(1)
	}
	os.WriteFile(wsourcesPath, []byte(xml.Header+string(xmlData)), 0644)
	defer os.Remove(wsourcesPath)


	logAndPrint("Starte Wwise Konvertierung...")

	// Wwise Konvertierung
	cmd := exec.Command(config.WwisePath, "convert-external-source",
		filepath.Join(execDir, config.ProjectName, config.ProjectName+".wproj"),
		"--source-file", wsourcesPath,
		"--output", execDir,
		"--quiet")

	// Pipe für die Standardausgabe erstellen
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logAndPrint(fmt.Sprintf("Fehler beim Erstellen der Stdout-Pipe: %v", err))
		return
	}

	// Pipe für die Fehlerausgabe erstellen
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logAndPrint(fmt.Sprintf("Fehler beim Erstellen der Stderr-Pipe: %v", err))
		return
	}

	// Kommando im Hintergrund starten
	if err := cmd.Start(); err != nil {
		logAndPrint(fmt.Sprintf("Fehler beim Starten der Wwise Konvertierung: %v", err))
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
		logAndPrint(fmt.Sprintf("\n\rFehler bei der Wwise Konvertierung: %v", err))
	} else {
		done <- true
		logAndPrint(fmt.Sprintf("\n\rWwise Konvertierung erfolgreich abgeschlossen!"))
	}
}