package main

import (
	"bufio"
	"flag"
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

// checkBsaarchExe sucht nach der BSArch.exe im Programmverzeichnis
func checkBsaarchExe() (string, error) {
	// Bestimme den Pfad des ausführenden Programms
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("ERROR: Could not get directory of executable: %w", err)
	}

	// Bestimme das Verzeichnis des Programms
	exeDir := filepath.Dir(exePath)

	// Definiere den Namen der BSArch.exe gemäß Betriebssystem
	bsaarchName := "BSArch.exe"
	if runtime.GOOS != "windows" {
		bsaarchName = "bsaarch"
	}

	// Erstelle den vollständigen Pfad zur BSArch.exe
	bsaarchPath := filepath.Join(exeDir, bsaarchName)

	// Prüfe, ob die Datei existiert
	_, err = os.Stat(bsaarchPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("ERROR: Could not find BSArch.exe at: %s", exeDir)
	} else if err != nil {
		return "", fmt.Errorf("ERROR: An error occured while checking BSArch.exe: %w", err)
	}

	// Führe einen einfachen Testbefehl aus, um die Funktionalität zu überprüfen
	testCmd := exec.Command(bsaarchPath, "-h")
	err = testCmd.Start()
	if err != nil {
		return "", fmt.Errorf("ERROR: Can not excecute BSArch.exe: %w", err)
	}
	// Beende den Testprozess
	testCmd.Process.Kill()

	return bsaarchPath, nil
}

// extractBsa führt die BSArch.exe aus und entpackt die angegebene BSA-Datei
func extractBsa(bsaarchPath, srcFile, dstDirectory string) error {
	// Bereite den Befehl vor
	cmd := exec.Command(bsaarchPath, "unpack", srcFile, dstDirectory, "-mt")

	// Erfasse die Ausgabe
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("BSArch.exe Error: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// updateAniamtedProgressBar stellt einen animierten Fortschrittsbar in der Konsole dar
func updateAniamtedProgressBar(currentProgress, amountTotal int, timeStart time.Time, animationCounter int) {
	width := 40 // width of the bar in chars

	// Berechne percentsatz
	percent := float64(currentProgress) / float64(amountTotal)

	// Berechne Anzahl der filleden Zeichen
	filled := int(percent * float64(width))

	// Animations-Zeichen
	animationSymbols := []string{"|", "/", "-", "\\"}
	animationSymbol := animationSymbols[animationCounter%len(animationSymbols)]

	// ASCII-Ladebar Zeichen
	filledChars := "#"

	// Erstelle den Ladebar
	bar := strings.Repeat(filledChars, filled) + strings.Repeat("-", width-filled)

	// Erstelle einen eingebetteten Animations-Cursor im Ladebar
	if filled < width {
		position := filled
		barRunes := []rune(bar)
		barRunes[position] = []rune(animationSymbol)[0]
		bar = string(barRunes)
	}

	// Lösche die aktuelle Zeile und zeige den bar an
	fmt.Printf("\r[%s] %3.0f%% %d/%d files extracted.", bar, percent*100, currentProgress, amountTotal)
}

func main() {
	// timeStart erfassen
	timeStart := time.Now()

	// Parse Kommandozeilenargumente
	parallel := flag.Int("p", runtime.NumCPU(), "Anzahl der parallel zu verarbeitenden Files")
	outputDir := flag.String("o", "", "Ausgabeverzeichnis für alle entpackten Files")
	maxRetries := flag.Int("retries", 3, "Anzahl der Wiederholungsversuche für failede Extraktionen")

	// Definiere spezifische Ausgabeverzeichnisse für nummerierte Files
	maxDirs := 20 // Maximale Anzahl von spezifischen Ausgabeverzeichnissen
	outputDirs := make([]*string, maxDirs)
	for i := 1; i <= maxDirs; i++ {
		outputDirs[i-1] = flag.String(fmt.Sprintf("o%d", i), "", fmt.Sprintf("Output directory for the %d. file", i))
	}

	flag.Parse()

	// Überprüfe, ob Files angegeben wurden
	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("ERROR: No .bsa files given")
		fmt.Println("Usage: bsa-multi.exe -o OUTPUT_DIRECTORY [-p AMOUNT_PARALLEL] FILE1.bsa FILE2.bsa ...")
		fmt.Println("Or: bsa-multi.exe -o1 OUTPUT_DIRECTORY1 -o2 OUTPUT_DIRECTORY2 ... [-p AMOUNT_PARALLEL] FILE1.bsa FILE2.bsa ...")
		os.Exit(1)
	}

	// Überprüfe Ausgabeverzeichnis(se)
	hasOutputDirectory := false
	if *outputDir != "" {
		hasOutputDirectory = true
	} else {
		// Prüfe, ob spezifische Ausgabeverzeichnisse angegeben wurden
		for i := 0; i < len(files) && i < maxDirs; i++ {
			if *outputDirs[i] != "" {
				hasOutputDirectory = true
				break
			}
		}
	}

	if !hasOutputDirectory {
		fmt.Println("ERROR: No output directory given (-o or -o1, -o2, ...)")
		fmt.Println("Usage: bsa-multi.exe -o OUTPUT_DIRECTORY [-p AMOUNT_PARALLEL] FILE1.bsa FILE2.bsa ...")
		fmt.Println("Or: bsa-multi.exe -o1 OUTPUT_DIRECTORY1 -o2 OUTPUT_DIRECTORY2 ... [-p AMOUNT_PARALLEL] FILE1.bsa FILE2.bsa ...")
		os.Exit(1)
	}

	// Überprüfe Dateierweiterungen
	var validFiles []string
	var skippedFiles []string

	for _, file := range files {
		if filepath.Ext(file) != ".bsa" {
			fmt.Printf("ERROR: %s is not a .bsa file\n", file)
			fmt.Println("This tool only accepts .bsa files")
			os.Exit(1)
		}

		// Prüfe, ob die Datei existiert
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Warning: File %s was not found. Skipping.\n", file)
			skippedFiles = append(skippedFiles, file)
		} else {
			validFiles = append(validFiles, file)
		}
	}

	// Falls keine gültigen Files gefunden wurden
	if len(validFiles) == 0 {
		fmt.Println("ERROR: None of the given .bsa files exist")
		os.Exit(1)
	}

	// Aktualisiere die Liste der zu verarbeitenden Files
	files = validFiles

	// Stelle sicher, dass alle nötigen Ausgabeverzeichnisse existieren
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			fmt.Printf("ERROR: Could not create output directory %s: %v\n", *outputDir, err)
			os.Exit(1)
		}
	}

	for i := 0; i < len(files) && i < maxDirs; i++ {
		if *outputDirs[i] != "" {
			if err := os.MkdirAll(*outputDirs[i], 0755); err != nil {
				fmt.Printf("ERROR: Could not create output directory %s: %v\n", *outputDirs[i], err)
				os.Exit(1)
			}
		}
	}

	// Finde BSArch.exe
	bsaarchPath, err := checkBsaarchExe()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		fmt.Println("Please make sure BSArch.exe is in the same directory!")
		os.Exit(1)
	}

	// Zeige Status an
	fmt.Printf("\n====================== BSA-MULTI ======================\n")
	fmt.Printf("BSArch:     %s\n", bsaarchPath)
	fmt.Printf("Files:     	%d .bsa files found\n", len(files))
	fmt.Printf("Parallel:   %d extractions at the same time\n", *parallel)
	fmt.Printf("Status:     Start extraction of .bsa files\n")
	fmt.Printf("-------------------------------------------------------\n")

	// Parallelverarbeitung einrichten
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *parallel)

	// Speichere successful und Fehler
	var successful []string
	var errors []string
	var mutex sync.Mutex // Schützt die Slices

	// Fortschrittsbar-Variablen
	var processed int32
	var progressMutex sync.Mutex
	amountTotal := len(files)
	var animationCounter int

	processFiles := func(filesToProcess []string, fileIndices []int, isRetry bool) []string {
		var failed []string

		// Starte Animation im Hintergrund
		animationsStop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					progressMutex.Lock()
					animationCounter++
					updateAniamtedProgressBar(int(atomic.LoadInt32(&processed)), amountTotal, timeStart, animationCounter)
					progressMutex.Unlock()
				case <-animationsStop:
					return
				}
			}
		}()

		for i, file := range filesToProcess {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(index int, fileIndex int, fileName string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// Bestimme das Zielverzeichnis
				targetDirectory := *outputDir
				// Wenn ein spezifisches Ausgabeverzeichnis für diese Datei vorhanden ist, verwende es
				if fileIndex < maxDirs && *outputDirs[fileIndex] != "" {
					targetDirectory = *outputDirs[fileIndex]
				}

				// Führe Extraktion durch, aber nur wenn ein Zielverzeichnis definiert ist
				if targetDirectory == "" {
					mutex.Lock()
					errors = append(errors, fmt.Sprintf("%s (no output directory set)", fileName))
					if !isRetry {
						failed = append(failed, fileName)
					}
					mutex.Unlock()
				} else {
					err := extractBsa(bsaarchPath, fileName, targetDirectory)

					// Ergebnis speichern (thread-sicher)
					mutex.Lock()

					if err != nil {
						errors = append(errors, fmt.Sprintf("%s (%v)", fileName, err))
						if !isRetry {
							failed = append(failed, fileName)
						}
					} else {
						successful = append(successful, fmt.Sprintf("%s -> %s", fileName, targetDirectory))
					}

					mutex.Unlock()
				}

				// Aktualisiere Fortschrittsbar
				atomic.AddInt32(&processed, 1)
			}(i, fileIndices[i], file)
		}

		// Warte auf Abschluss aller Extraktionen
		wg.Wait()

		// Animationsschleife stoppen
		close(animationsStop)
		time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird

		return failed
	}

	// Erste Durchführung mit allen Files
	fileIndices := make([]int, len(files))
	for i := range fileIndices {
		fileIndices[i] = i
	}

	failed := processFiles(files, fileIndices, false)

	// Wiederholungsversuche für failede Files
	repeatCounter := 0
	for repeatCounter < *maxRetries && len(failed) > 0 {
		repeatCounter++
		fmt.Printf("\n\nError for %d files have been detected. Wait 3 seconds before trying again...\n",
			len(failed))

		// Warte 3 Sekunden vor dem Wiederholungsversuch
		for countdown := 3; countdown > 0; countdown-- {
			fmt.Printf("\rRetry attempt starts in %d seconds...", countdown)
			time.Sleep(1 * time.Second)
		}

		fmt.Printf("\rRetry attempt %d of %d for %d failed files...\n",
			repeatCounter, *maxRetries, len(failed))

		// Zurücksetzen des Fortschrittsbars für die Wiederholungsversuche
		processed = 0
		amountTotal = len(failed)

		// Erstelle die entsprechenden Indizes
		failedIndizes := make([]int, len(failed))
		for i, failedFile := range failed {
			for j, originalFile := range files {
				if failedFile == originalFile {
					failedIndizes[i] = j
					break
				}
			}
		}

		// Führe die faileden Files erneut aus
		failed = processFiles(failed, failedIndizes, true)
	}

	// Zeige finalen Fortschrittsbar
	progressMutex.Lock()
	updateAniamtedProgressBar(amountTotal, amountTotal, timeStart, animationCounter)
	progressMutex.Unlock()

	// Zeile nach Fortschrittsbar
	fmt.Println("\n")

	// Berechne die Gesamtzeit
	timeTotal := time.Since(timeStart)

	// Zeige Zusammenfassung an
	fmt.Println("\n=== Result ===")

	// Zeige übersprungene nicht existierende Files
	if len(skippedFiles) > 0 {
		fmt.Printf("\nSkipped files (%d):\n", len(skippedFiles))
		for i, file := range skippedFiles {
			fmt.Printf("%d. %s\n", i+1, file)
		}
	}

	fmt.Printf("Extracted files (%d):\n", len(successful))
	for i, file := range successful {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	if len(errors) > 0 {
		fmt.Printf("\nFailed files (%d):\n", len(errors))
		for i, file := range errors {
			fmt.Printf("%d. %s\n", i+1, file)
		}
		fmt.Printf("\nTime total: %s\n", timeTotal)

		// Wenn nach allen Wiederholungsversuchen immer noch Fehler bestehen, pausiere das Programm
		if len(failed) > 0 {
			fmt.Println("\n\nERROR: Could not process file.")
			fmt.Println("Press Enter to close the programm...")

			// Warte auf Benutzereingabe
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}

		os.Exit(1)
	} else {
		if len(skippedFiles) > 0 {
			fmt.Println("\nSuccessfully extracted all files!")
			fmt.Printf("(%d files were skipped)\n", len(skippedFiles))
		} else {
			fmt.Println("\nSuccessfully extracted all files!")
		}
		fmt.Printf("Time total: %s\n", timeTotal)
	}
}
