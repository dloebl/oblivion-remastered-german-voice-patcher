package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"strings"
)

// prüfeBsaarchExe sucht nach der BSArch.exe im Programmverzeichnis
func prüfeBsaarchExe() (string, error) {
	// Bestimme den Pfad des ausführenden Programms
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("konnte den Programmpfad nicht ermitteln: %w", err)
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
		return "", fmt.Errorf("BSArch.exe wurde nicht im Programmverzeichnis gefunden: %s", exeDir)
	} else if err != nil {
		return "", fmt.Errorf("fehler beim Überprüfen von BSArch.exe: %w", err)
	}
	
	// Führe einen einfachen Testbefehl aus, um die Funktionalität zu überprüfen
	testCmd := exec.Command(bsaarchPath, "-h")
	err = testCmd.Start()
	if err != nil {
		return "", fmt.Errorf("BSArch.exe kann nicht ausgeführt werden: %w", err)
	}
	// Beende den Testprozess
	testCmd.Process.Kill()
	
	return bsaarchPath, nil
}

// extrahiereBsa führt die BSArch.exe aus und entpackt die angegebene BSA-Datei
func extrahiereBsa(bsaarchPath, quelldatei, zielpfad string) error {
	// Bereite den Befehl vor
	cmd := exec.Command(bsaarchPath, "unpack", quelldatei, zielpfad, "-mt")
	
	// Erfasse die Ausgabe
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("BSArch.exe Fehler: %w\nAusgabe: %s", err, string(output))
	}
	
	return nil
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
	fmt.Printf("\r[%s] %3.0f%% %d/%d", balken, prozent*100, aktuellerFortschritt, gesamtAnzahl)
}

func main() {
	// Startzeit erfassen
	startZeit := time.Now()
	
	// Parse Kommandozeilenargumente
	parallel := flag.Int("p", runtime.NumCPU(), "Anzahl der parallel zu verarbeitenden Dateien")
	outputDir := flag.String("o", "", "Ausgabeverzeichnis für entpackte Dateien")
	flag.Parse()
	
	// Überprüfe Ausgabeverzeichnis
	if *outputDir == "" {
		fmt.Println("Fehler: Kein Ausgabeverzeichnis angegeben (-o)")
		fmt.Println("Verwendung: bsaextract -o AUSGABEVERZEICHNIS [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		os.Exit(1)
	}
	
	// Überprüfe, ob Dateien angegeben wurden
	dateien := flag.Args()
	if len(dateien) == 0 {
		fmt.Println("Fehler: Keine BSA-Dateien angegeben")
		fmt.Println("Verwendung: bsaextract -o AUSGABEVERZEICHNIS [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		os.Exit(1)
	}
	
	// Überprüfe Dateierweiterungen
	for _, datei := range dateien {
		if filepath.Ext(datei) != ".bsa" {
			fmt.Printf("Fehler: %s ist keine .bsa-Datei\n", datei)
			fmt.Println("Dieses Programm akzeptiert nur Dateien mit der Endung .bsa")
			os.Exit(1)
		}
		
		// Prüfe zusätzlich, ob die Datei existiert
		if _, err := os.Stat(datei); os.IsNotExist(err) {
			fmt.Printf("Fehler: Datei %s existiert nicht\n", datei)
			os.Exit(1)
		}
	}
	
	// Stelle sicher, dass das Ausgabeverzeichnis existiert
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Fehler beim Erstellen des Ausgabeverzeichnisses %s: %v\n", *outputDir, err)
		os.Exit(1)
	}
	
	// Finde BSArch.exe
	bsaarchPath, err := prüfeBsaarchExe()
	if err != nil {
		fmt.Printf("Fehler: %v\n", err)
		fmt.Println("Bitte stellen Sie sicher, dass die BSArch.exe im selben Verzeichnis wie dieses Programm vorhanden ist.")
		os.Exit(1)
	}
	
	fmt.Printf("BSArch.exe gefunden: %s\n", bsaarchPath)
	fmt.Printf("Extrahiere %d Dateien nach %s (max. %d parallel)\n", len(dateien), *outputDir, *parallel)
	
	// Parallelverarbeitung einrichten
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *parallel)
	
	// Speichere Erfolge und Fehler
	var erfolge []string
	var fehler []string
	var mutex sync.Mutex  // Schützt die Slices
	
	// Fortschrittsbalken-Variablen
	var erledigt int32
	var fortschrittsMutex sync.Mutex
	gesamtAnzahl := len(dateien)
	
	// Zeige initialen Fortschrittsbalken
	fmt.Println("Starte Extraktionen:")
	
	// Starte Animation im Hintergrund
	animationsStopp := make(chan struct{})
	var animationsZähler int
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				fortschrittsMutex.Lock()
				animationsZähler++
				zeichneAnimiertenFortschrittsbalken(int(atomic.LoadInt32(&erledigt)), gesamtAnzahl, startZeit, animationsZähler)
				fortschrittsMutex.Unlock()
			case <-animationsStopp:
				return
			}
		}
	}()
	
	// Verarbeite alle Dateien
	for _, datei := range dateien {
		wg.Add(1)
		go func(dateiname string) {
			defer wg.Done()
			
			// Semaphore erwerben
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Führe Extraktion durch
			err := extrahiereBsa(bsaarchPath, dateiname, *outputDir)
			
			// Ergebnis speichern (thread-sicher)
			mutex.Lock()
			
			if err != nil {
				fehler = append(fehler, dateiname)
			} else {
				erfolge = append(erfolge, dateiname)
			}
			
			// Aktualisiere Fortschrittsbalken
			atomic.AddInt32(&erledigt, 1)
			mutex.Unlock()
		}(datei)
	}
	
	// Warte auf Abschluss aller Extraktionen
	wg.Wait()
	
	// Animationsschleife stoppen
	close(animationsStopp)
	time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird
	
	// Zeige finalen Fortschrittsbalken
	fortschrittsMutex.Lock()
	zeichneAnimiertenFortschrittsbalken(gesamtAnzahl, gesamtAnzahl, startZeit, animationsZähler)
	fortschrittsMutex.Unlock()
	
	// Zeile nach Fortschrittsbalken
	fmt.Println("\n")
	
	// Berechne die Gesamtzeit
	gesamtZeit := time.Since(startZeit)
	
	// Zeige Zusammenfassung an
	fmt.Println("\n=== Zusammenfassung ===")
	fmt.Printf("Erfolgreich extrahierte Dateien (%d):\n", len(erfolge))
	for i, datei := range erfolge {
		fmt.Printf("%d. %s\n", i+1, datei)
	}
	
	if len(fehler) > 0 {
		fmt.Printf("\nFehlgeschlagene Extraktionen (%d):\n", len(fehler))
		for i, datei := range fehler {
			fmt.Printf("%d. %s\n", i+1, datei)
		}
		fmt.Printf("\nGesamtzeit: %s\n", gesamtZeit)
		os.Exit(1)
	} else {
		fmt.Println("\nAlle Extraktionen erfolgreich abgeschlossen!")
		fmt.Printf("Gesamtzeit: %s\n", gesamtZeit)
	}
}
