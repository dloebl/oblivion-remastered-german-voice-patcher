package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Fehler: Keine Eingabedateien angegeben")
		return
	}

	// Konfiguration laden
	execDir, err := os.Executable()
	if err != nil {
		fmt.Println("Fehler beim Ermitteln des Ausführungsverzeichnisses:", err)
		return
	}
	execDir = filepath.Dir(execDir)

	config, err := loadConfig(execDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Temporäres Verzeichnis erstellen
	tempDir := filepath.Join(execDir, "audiotemp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

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
				semaphore <- struct{}{} // Slot belegen
				defer func() { <-semaphore }() // Slot freigeben

				outputFile := filepath.Join(tempDir, filepath.Base(inputFile))
				outputFile = outputFile[:len(outputFile)-len(filepath.Ext(outputFile))] + ".wav"
				
				fmt.Printf("Konvertiere: %s -> %s\n", inputFile, filepath.Base(outputFile))
				
				cmd := exec.Command(config.FfmpegPath, "-hide_banner", "-loglevel", "warning", 
					"-i", inputFile, outputFile)
				if err := cmd.Run(); err != nil {
					fmt.Printf("Fehler bei der Konvertierung von %s: %v\n", inputFile, err)
				} else {
					fmt.Printf("Erfolgreich konvertiert: %s\n", filepath.Base(outputFile))
				}
			}(file)
		}
	}
	wg.Wait()

	fmt.Println("Alle Audio-Dateien konvertiert. Erstelle XML...")

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
	xmlData, _ := xml.MarshalIndent(sources, "", "  ")
	os.WriteFile(wsourcesPath, []byte(xml.Header+string(xmlData)), 0644)
	defer os.Remove(wsourcesPath)


	fmt.Println("Starte Wwise Konvertierung...")

	// Wwise Konvertierung
	cmd := exec.Command(config.WwisePath, "convert-external-source",
		filepath.Join(execDir, config.ProjectName, config.ProjectName+".wproj"),
		"--source-file", wsourcesPath,
		"--output", execDir)  // "--quiet" Flag entfernt

	// Stdout und Stderr an die Konsole weiterleiten
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Fehler bei der Wwise Konvertierung: %v\n", err)
	} else {
		fmt.Println("Wwise Konvertierung erfolgreich abgeschlossen!")
	}
}