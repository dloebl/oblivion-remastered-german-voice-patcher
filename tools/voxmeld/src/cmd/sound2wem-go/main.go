package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dloebl/wemenc/pkg/wemenc"
)

type Config struct {
	FfmpegPath string `json:"ffmpegPath"`
	Bitrate    string `json:"bitrate"`
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
	if config.FfmpegPath == "" {
		config.FfmpegPath = filepath.Join(execDir, "ffmpeg-master-latest-win64-gpl-shared", "bin", "ffmpeg.exe")
	}
	if config.Bitrate == "" {
		config.Bitrate = "64k"
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
	percent := 0.0
	if amountTotal > 0 {
		percent = float64(currentProgress) / float64(amountTotal)
	}

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

	// Output Verzeichnis (Direkt in tmp/wem, wie von Create-Mod.bat erwartet)
	outputDir := filepath.Join(execDir, "..", "..", "tmp", "wem")
	os.MkdirAll(outputDir, 0755)

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
	fmt.Printf("\n====================== SOUND2WEM (wemenc) ======================\n")
	fmt.Printf("Files:     	%d files found\n", totalFiles)
	fmt.Printf("Source:     %s\n", strings.Join(os.Args[1:], ", "))
	fmt.Printf("Status:     Starting to convert files to .wem format using wemenc (Opus)\n")
	fmt.Printf("---------------------------------------------------------------\n")

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

	fmt.Printf("Converting files to .wem with %d parallel processes...\n", numCPU)

	for _, pattern := range os.Args[1:] {
		matches, _ := filepath.Glob(pattern)
		for _, file := range matches {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(inputFile string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				outputFile := filepath.Join(outputDir, filepath.Base(inputFile))
				outputFile = outputFile[:len(outputFile)-len(filepath.Ext(outputFile))] + ".wem"

				inFile, err := os.Open(inputFile)
				if err != nil {
					fmt.Printf("\nERROR: Could not open %s: %v\n", inputFile, err)
					return
				}
				defer inFile.Close()

				outFile, err := os.Create(outputFile)
				if err != nil {
					fmt.Printf("\nERROR: Could not create %s: %v\n", outputFile, err)
					return
				}
				defer outFile.Close()

				opt := wemenc.EncodeOptions{
					Codec:      wemenc.CodecOpus,
					Bitrate:    config.Bitrate,
					FFmpegPath: config.FfmpegPath,
				}

				if err := wemenc.EncodeToWEM(inFile, outFile, opt); err != nil {
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

	// Berechne die Gesamtzeit
	timeTotal := time.Since(timeStart)
	fmt.Printf("\n\nSuccessfully converted all files to .wem!\n")
	fmt.Printf("Time total: %s\n", timeTotal)
}
