package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"runtime"
	"time"
)

var animationsStop chan struct{}
var animationsCounter int

var fileMutex sync.Mutex
var progressMutex sync.Mutex

var logger *log.Logger
var logFile *os.File

func main() {
	// Get time
	timeStart := time.Now()

	setupLogging()
	defer logFile.Close()

	wemFolder := "tmp/wem"
	bnkFolder := "tmp/bnk"
	pakFolder := "tmp/pak"

	// Gib das aktuelle Arbeitsverzeichnis aus
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Printf("Working directory: %s", workingDirectory)
	}

	// Erstelle die Ausgabeverzeichnisse
	createOutputDirs(bnkFolder)

	processFiles(wemFolder, bnkFolder, pakFolder, timeStart)
}
func setupLogging() {
	// Erstelle Log-Verzeichnis, wenn es nicht existiert
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Printf("ERROR: Could not create logs folder: %v\n", err)
		return
	}

	// Öffne die Log-Datei
	logFile, err = os.OpenFile("logs/Missing-bnks-for-wem.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("ERROR: Could not open log file: %v\n", err)
		return
	}

	// Initialisiere den Logger
	logger = log.New(logFile, "", log.LstdFlags)
}

// Function for logging missing wem files
func logMissingFile(message string) {
	if logger != nil {
		logger.Println(message)
	}
}

// Verzeichnisse erstellen, die im Shell-Script definiert sind
func createOutputDirs(outputDir string) {
	dirs := []string{
		outputDir + "/Content/WwiseAudio/Event/English(US)/",
		outputDir + "/Content/WwiseAudio/Media/English(US)/",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("ERROR: Could not create directory %s (err: %v)", dir, err)
		}
	}
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

func processFiles(wemFolder string, bnkFolder string, pakFolder string, timeStart time.Time) {
	var wg sync.WaitGroup
	var totalFiles int
	var processedFiles int32

	maxParallelism := runtime.NumCPU() * 2
	if maxParallelism > 16 {
		maxParallelism = 16
	}
	semaphore := make(chan struct{}, maxParallelism)

	// Wenn ein Argument übergeben wurde, verarbeite nur diese eine Datei
	wems, err := filepath.Glob(filepath.Join(wemFolder, "*"))
	if err != nil {
		fmt.Printf("ERROR: Could not look up files in %s: %v", wemFolder, err)
		return
	}

	if len(os.Args) > 1 {
		wems = []string{os.Args[1]}
	}

	fmt.Printf("Counting .bnk files to create...")
		
	for _, wemPathWithType := range wems {
		wemPath := strings.TrimSuffix(wemPathWithType, filepath.Ext(wemPathWithType))
		comps := strings.Split(filepath.Base(wemPath), "_")

		if len(comps) < 3 {
			// Incorrect naming
			logMissingFile(comps[0] + ".wem")
			continue
		}

		// File is audio for video
		if comps[0] == "scripted" {
			bnkName := filepath.Base(wemPath)
			bnkPath := filepath.Join(pakFolder, "OblivionRemastered/Content/WwiseAudio/Event", bnkName + ".bnk")
		
			_, err := os.ReadFile(bnkPath)
			if err != nil {
				logMissingFile(bnkName + ".wem")
				continue
			}

			totalFiles++
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

			variantCounter := 0
			originalFile := ""
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

					// Save original file name for potentially missing files
					if race == raceComb && variant == "" {
						originalFile = bnkName
					}

					bnkPath := pakFolder + "/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + ".bnk"
				
					_, err := os.ReadFile(bnkPath)
					if err != nil {
						// Check for '_sid' variant
						bnkPath = pakFolder + "/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + "_sid.bnk"
					
						_, err = os.ReadFile(bnkPath)
						if err != nil {
							variantCounter += 1
							
							if variantCounter == len(variants) * len(races){
								logMissingFile(originalFile + ".wem")
							}

							continue
						}
					}

					totalFiles++
				}
			}
		}
	}

	// Show step
	fmt.Printf("\n====================== VOXMELD ======================\n")
	fmt.Printf("Source:     	%s\n", wemFolder)
	fmt.Printf("Files:     	%d .wem files found\n", totalFiles)
	fmt.Printf("Status:     	Start creating .bnk files\n")
	fmt.Printf("-----------------------------------------------------\n")

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

	for _, wemPathWithType := range wems {
		wemPath := strings.TrimSuffix(wemPathWithType, filepath.Ext(wemPathWithType))
		comps := strings.Split(filepath.Base(wemPath), "_")

		if len(comps) < 3 {
			// Incorrect naming
			continue
		}

		// File is audio for video
		if comps[0] == "scripted" {
			bnkName := filepath.Base(wemPath)
			bnkPath := filepath.Join(pakFolder, "OblivionRemastered/Content/WwiseAudio/Event", bnkName + ".bnk")
		
			bnk, err := os.ReadFile(bnkPath)
			if err != nil {
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{}

			go func(file []byte, fileName, filePath, fileFolder, srcPath string, isVideo bool) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if err := create_bnk(file, fileName, filePath, fileFolder, srcPath, isVideo); err != nil {
					fmt.Printf("Fehler beim Erstellen von %s : %v", fileName, err)
				}
				atomic.AddInt32(&processedFiles, 1)
			}(bnk, bnkName, bnkPath, bnkFolder, wemPath, true)
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
					bnkPath := pakFolder + "/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + ".bnk"
				
					bnk, err := os.ReadFile(bnkPath)
					if err != nil {
						// Check for '_sid' variant
						bnkPath = pakFolder + "/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + "_sid.bnk"
					
						bnk, err = os.ReadFile(bnkPath)
						if err != nil {
							continue
						}
					}

					wg.Add(1)
					semaphore <- struct{}{}

					go func(file []byte, fileName, filePath, fileFolder, srcPath string, isVideo bool) {
						defer wg.Done()
						defer func() { <-semaphore }()

						if err := create_bnk(file, fileName, filePath, fileFolder, srcPath, isVideo); err != nil {
							fmt.Printf("Fehler beim Erstellen von %s : %v", fileName, err)
						}
						atomic.AddInt32(&processedFiles, 1)
					}(bnk, bnkName, bnkPath, bnkFolder, wemPath, false)
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

	// Calculate time total
	timeTotal := time.Since(timeStart)
	fmt.Printf("\n\nAll .bnk files have been created in %s\n", timeTotal)
}

func create_bnk(bnk []byte, bnkName string, bnkPath string, bnkFolder string, wemPath string, isVideo bool) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	pattern := []byte{0x01, 0x00, 0x14, 0x00} // Codec: OPUS_WEM
	if isVideo == true {
		pattern = []byte{0x01, 0x00, 0x01, 0x00} // Codec: PCM
	}
	newCodec := []byte{0x01, 0x00, 0x04, 0x00} // Codec: VORBIS
	// Find the pattern in the file:
	// Quick and dirty approach to patch the BNKs
	pos := bytes.Index(bnk, pattern)
	if pos == -1 {
		pos = bytes.Index(bnk, newCodec)
		if pos == -1 {
			return fmt.Errorf("Pattern not found: %s", bnkName)
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
		return fmt.Errorf("Failed to read .wem file: %v", err)
	}
	wemSize := uint32(wemInfo.Size())
	// Update the codex to VORBIS
	copy(bnk[pos:pos+4], newCodec)
	// Update file size (4 bytes after dummy byte and ID)
	fileSizeOffset := pos + 9
	if fileSizeOffset+4 > len(bnk) {
		return errors.New("Not enough data to update file size in .bnk")
	}
	binary.LittleEndian.PutUint32(bnk[fileSizeOffset:fileSizeOffset+4], wemSize)
	// write the modified .bnk file to the output folder
	outBnkPath := filepath.Join(bnkFolder, "Content/WwiseAudio/Event/English(US)", filepath.Base(bnkPath))
	if isVideo == true {
		outBnkPath = filepath.Join(bnkFolder, "Content/WwiseAudio/Event", filepath.Base(bnkPath))
	}
	err = os.WriteFile(outBnkPath, bnk, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write modified .bnk file: %v", err)
	}
	//fmt.Printf("Modified .bnk file written to: %s\n", outBnkPath)
	// write the .wem file to output folder
	wem, err := os.ReadFile(wemPath + ".wem")
	if err != nil {
		return fmt.Errorf("Failed to read .wem file: %v", err)
	}
	outWemPath := filepath.Join(bnkFolder, "Content/WwiseAudio/Media/English(US)", strconv.Itoa(int(id)) + ".wem")
	if isVideo == true {
		outWemPath = filepath.Join(bnkFolder, "Content/WwiseAudio/Media", strconv.Itoa(int(id)) + ".wem")
	}
	err = os.WriteFile(outWemPath, wem, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write .wem file: %v", err)
	}

	return err
}
