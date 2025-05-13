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
var processedAudioFiles int32
var totalAudioFiles int32

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

// Print message function
func printMessage(message string) {
	fmt.Println(message)
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

func main() {
	// Startzeit erfassen
	timeStart := time.Now()

	if len(os.Args) < 2 {
		printMessage("ERROR: No input files given")
		return
	}

	// Konfiguration laden
	execDir, err := os.Executable()
	if err != nil {
		printMessage(fmt.Sprintf("ERROR: Could not get execution path: %v", err))
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
		printMessage("Creating new Wwise project...")
		cmd := exec.Command(config.WwisePath, "create-new-project",
			filepath.Join(projectPath, config.ProjectName+".wproj"),
			"--quiet")
		if err := cmd.Run(); err != nil {
			printMessage(fmt.Sprintf("ERROR: Could not create Wwsie project: %v", err))
			return
		}
	}

	// Temporäres Verzeichnis erstellen
	tempDir := filepath.Join(execDir, "..", "..", "tmp", "wav")
	os.MkdirAll(tempDir, 0755)
	// defer os.RemoveAll(tempDir)

	var wg sync.WaitGroup
	var totalFiles int
	var progressMutex sync.Mutex
	var animationsCounter int

	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)
	
	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		totalFiles += len(matches)
	}
	atomic.StoreInt32(&totalAudioFiles, int32(totalFiles))

	// Zeige Step an.
	fmt.Printf("\n====================== SOUND2WEM ======================\n")
	fmt.Printf("Source:     %s\n", strings.Join(os.Args[1:], ", "))
	fmt.Printf("Files:     	%d files found\n", totalFiles)
	fmt.Printf("Status:     Starting to convert files to .wav and .wem format\n")
	fmt.Printf("-------------------------------------------------------\n")

	// Start animation in the background
	animationsStop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progressMutex.Lock()
				animationsCounter++
				updateAnimatedProgressBar(
					int(atomic.LoadInt32(&processedAudioFiles)),
					int(atomic.LoadInt32(&totalAudioFiles)),
					timeStart,
					animationsCounter)
				progressMutex.Unlock()
			case <-animationsStop:
				return
			}
		}
	}()

	fmt.Printf("Starting to convert files to .wav with %d parallel processes...\n", numCPU)

	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		for _, file := range matches {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(inputFile string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				outputFile := filepath.Join(tempDir, filepath.Base(inputFile))
				outputFile = outputFile[:len(outputFile)-len(filepath.Ext(outputFile))] + ".wav"

				cmd := exec.Command(config.FfmpegPath, "-hide_banner", "-loglevel", "warning",
					"-i", inputFile, outputFile)
				if err := cmd.Run(); err != nil {
					fmt.Printf("\nERROR: Could not convert %s to .wem: %v\n", inputFile, err)
				}

				// Upate progress
				atomic.AddInt32(&processedAudioFiles, 1)
			}(file)
		}
	}
	wg.Wait()

	// Stop animation loop
	close(animationsStop)
	time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

	// Show final progress bar
	progressMutex.Lock()
	updateAnimatedProgressBar(
		int(atomic.LoadInt32(&processedAudioFiles)),
		int(atomic.LoadInt32(&totalAudioFiles)),
		timeStart,
		animationsCounter)
	progressMutex.Unlock()

	fmt.Println("\n\nAll audio files have been converted to .wav! Creating XML...")

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
		printMessage(fmt.Sprintf("ERROR: Could not create XML file: %v", err))
		os.Exit(1)
	}
	os.WriteFile(wsourcesPath, []byte(xml.Header+string(xmlData)), 0644)
	defer os.Remove(wsourcesPath)

	printMessage("Starting to convert files to .wem...")

	// Wwise Konvertierung
	cmd := exec.Command(config.WwisePath, "convert-external-source",
		filepath.Join(execDir, config.ProjectName, config.ProjectName+".wproj"),
		"--source-file", wsourcesPath,
		"--output", filepath.Join(execDir, "..", "..", "tmp"),
		"--quiet")

	// Pipe für die Standardausgabe erstellen
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		printMessage(fmt.Sprintf("ERROR: Could not create Stdout-Pipe: %v", err))
		return
	}

	// Pipe für die Fehlerausgabe erstellen
	stderr, err := cmd.StderrPipe()
	if err != nil {
		printMessage(fmt.Sprintf("ERROR: Could not create Stderr-Pipe: %v", err))
		return
	}

	// Kommando im Hintergrund starten
	if err := cmd.Start(); err != nil {
		printMessage(fmt.Sprintf("Error: Could not start converting files: %v", err))
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
				fmt.Printf("\rConverting files... %s", spinner[i])
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
		printMessage(fmt.Sprintf("\n\rERROR: Could not convert .wav files: %v", err))
	} else {
		done <- true
		printMessage(fmt.Sprintf("\n\rSuccessfully converted all files to .wem!"))
	}

	// Berechne die Gesamtzeit
	timeTotal := time.Since(timeStart)
	fmt.Printf("\nTime total: %s\n", timeTotal)
}
