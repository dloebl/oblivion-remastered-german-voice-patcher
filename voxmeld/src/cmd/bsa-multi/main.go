package main

import (
	"bufio"
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
	outputDir := flag.String("o", "", "Ausgabeverzeichnis für alle entpackten Dateien")
	maxRetries := flag.Int("retries", 3, "Anzahl der Wiederholungsversuche für fehlgeschlagene Extraktionen")
	
	// Definiere spezifische Ausgabeverzeichnisse für nummerierte Dateien
	maxDirs := 20 // Maximale Anzahl von spezifischen Ausgabeverzeichnissen
	outputDirs := make([]*string, maxDirs)
	for i := 1; i <= maxDirs; i++ {
		outputDirs[i-1] = flag.String(fmt.Sprintf("o%d", i), "", fmt.Sprintf("Ausgabeverzeichnis für die %d. Datei", i))
	}
	
	flag.Parse()
	
	// Überprüfe, ob Dateien angegeben wurden
	dateien := flag.Args()
	if len(dateien) == 0 {
		fmt.Println("Fehler: Keine BSA-Dateien angegeben")
		fmt.Println("Verwendung: bsa-multi.exe -o AUSGABEVERZEICHNIS [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		fmt.Println("Oder: bsa-multi.exe -o1 AUSGABEVERZEICHNIS1 -o2 AUSGABEVERZEICHNIS2 ... [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		os.Exit(1)
	}
	
	// Überprüfe Ausgabeverzeichnis(se)
	hatAusgabeverzeichnis := false
	if *outputDir != "" {
		hatAusgabeverzeichnis = true
	} else {
		// Prüfe, ob spezifische Ausgabeverzeichnisse angegeben wurden
		for i := 0; i < len(dateien) && i < maxDirs; i++ {
			if *outputDirs[i] != "" {
				hatAusgabeverzeichnis = true
				break
			}
		}
	}
	
	if !hatAusgabeverzeichnis {
		fmt.Println("Fehler: Kein Ausgabeverzeichnis angegeben (-o oder -o1, -o2, ...)")
		fmt.Println("Verwendung: bsa-multi.exe -o AUSGABEVERZEICHNIS [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		fmt.Println("Oder: bsa-multi.exe -o1 AUSGABEVERZEICHNIS1 -o2 AUSGABEVERZEICHNIS2 ... [-p ANZAHL_PARALLEL] DATEI1.bsa DATEI2.bsa ...")
		os.Exit(1)
	}
	
	// Überprüfe Dateierweiterungen
	var gültigeDateien []string
	var nichtExistierendeDateien []string
	
	for _, datei := range dateien {
		if filepath.Ext(datei) != ".bsa" {
			fmt.Printf("Fehler: %s ist keine .bsa-Datei\n", datei)
			fmt.Println("Dieses Programm akzeptiert nur Dateien mit der Endung .bsa")
			os.Exit(1)
		}
		
		// Prüfe, ob die Datei existiert
		if _, err := os.Stat(datei); os.IsNotExist(err) {
			fmt.Printf("Warnung: Datei %s existiert nicht. Diese wird übersprungen.\n", datei)
			nichtExistierendeDateien = append(nichtExistierendeDateien, datei)
		} else {
			gültigeDateien = append(gültigeDateien, datei)
		}
	}
	
	// Falls keine gültigen Dateien gefunden wurden
	if len(gültigeDateien) == 0 {
		fmt.Println("Fehler: Keine der angegebenen BSA-Dateien existiert")
		os.Exit(1)
	}
	
	// Aktualisiere die Liste der zu verarbeitenden Dateien
	dateien = gültigeDateien
	
	// Stelle sicher, dass alle nötigen Ausgabeverzeichnisse existieren
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			fmt.Printf("Fehler beim Erstellen des Ausgabeverzeichnisses %s: %v\n", *outputDir, err)
			os.Exit(1)
		}
	}
	
	for i := 0; i < len(dateien) && i < maxDirs; i++ {
		if *outputDirs[i] != "" {
			if err := os.MkdirAll(*outputDirs[i], 0755); err != nil {
				fmt.Printf("Fehler beim Erstellen des Ausgabeverzeichnisses %s: %v\n", *outputDirs[i], err)
				os.Exit(1)
			}
		}
	}
	
	// Finde BSArch.exe
	bsaarchPath, err := prüfeBsaarchExe()
	if err != nil {
		fmt.Printf("Fehler: %v\n", err)
		fmt.Println("Bitte stellen Sie sicher, dass die BSArch.exe im selben Verzeichnis wie dieses Programm vorhanden ist.")
		os.Exit(1)
	}
	
	fmt.Printf("BSArch.exe gefunden: %s\n", bsaarchPath)
	fmt.Printf("Extrahiere %d Dateien (max. %d parallel)\n", len(dateien), *parallel)
	
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
	var animationsZähler int
	
	// Zeige initialen Fortschrittsbalken
	fmt.Println("Starte Extraktionen:")
	
	// Verarbeite alle Dateien
	prozessiereDateien := func(dateienZuVerarbeiten []string, dateienIndizes []int, istWiederholung bool) []string {
		var fehlgeschlagen []string
		
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
					zeichneAnimiertenFortschrittsbalken(int(atomic.LoadInt32(&erledigt)), gesamtAnzahl, startZeit, animationsZähler)
					fortschrittsMutex.Unlock()
				case <-animationsStopp:
					return
				}
			}
		}()
		
		for i, datei := range dateienZuVerarbeiten {
			wg.Add(1)
			go func(index int, dateiIndex int, dateiname string) {
				defer wg.Done()
				
				// Semaphore erwerben
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				// Bestimme das Zielverzeichnis
				zielverzeichnis := *outputDir
				// Wenn ein spezifisches Ausgabeverzeichnis für diese Datei vorhanden ist, verwende es
				if dateiIndex < maxDirs && *outputDirs[dateiIndex] != "" {
					zielverzeichnis = *outputDirs[dateiIndex]
				}
				
				// Führe Extraktion durch, aber nur wenn ein Zielverzeichnis definiert ist
				if zielverzeichnis == "" {
					mutex.Lock()
					fehler = append(fehler, fmt.Sprintf("%s (kein Ausgabeverzeichnis angegeben)", dateiname))
					if !istWiederholung {
						fehlgeschlagen = append(fehlgeschlagen, dateiname)
					}
					mutex.Unlock()
				} else {
					err := extrahiereBsa(bsaarchPath, dateiname, zielverzeichnis)
					
					// Ergebnis speichern (thread-sicher)
					mutex.Lock()
					
					if err != nil {
						fehler = append(fehler, fmt.Sprintf("%s (%v)", dateiname, err))
						if !istWiederholung {
							fehlgeschlagen = append(fehlgeschlagen, dateiname)
						}
					} else {
						erfolge = append(erfolge, fmt.Sprintf("%s -> %s", dateiname, zielverzeichnis))
					}
					
					mutex.Unlock()
				}
				
				// Aktualisiere Fortschrittsbalken
				atomic.AddInt32(&erledigt, 1)
			}(i, dateienIndizes[i], datei)
		}
		
		// Warte auf Abschluss aller Extraktionen
		wg.Wait()
		
		// Animationsschleife stoppen
		close(animationsStopp)
		time.Sleep(200 * time.Millisecond) // Kurz warten, damit die Animation sauber beendet wird
		
		return fehlgeschlagen
	}

	// Erste Durchführung mit allen Dateien
	dateienIndizes := make([]int, len(dateien))
	for i := range dateienIndizes {
		dateienIndizes[i] = i
	}
	
	fehlgeschlagen := prozessiereDateien(dateien, dateienIndizes, false)
	
	// Wiederholungsversuche für fehlgeschlagene Dateien
	wiederholungszähler := 0
	for wiederholungszähler < *maxRetries && len(fehlgeschlagen) > 0 {
		wiederholungszähler++
		fmt.Printf("\n\nFehler bei %d Dateien festgestellt. Warte 3 Sekunden vor dem Wiederholungsversuch...\n", 
			len(fehlgeschlagen))
		
		// Warte 3 Sekunden vor dem Wiederholungsversuch
		for countdown := 3; countdown > 0; countdown-- {
			fmt.Printf("\rWiederholungsversuch startet in %d Sekunden...", countdown)
			time.Sleep(1 * time.Second)
		}
		
		fmt.Printf("\rWiederholungsversuch %d von %d für %d fehlgeschlagene Dateien...\n", 
			wiederholungszähler, *maxRetries, len(fehlgeschlagen))
		
		// Zurücksetzen des Fortschrittsbalkens für die Wiederholungsversuche
		erledigt = 0
		gesamtAnzahl = len(fehlgeschlagen)
		
		// Erstelle die entsprechenden Indizes
		fehlgeschlagenIndizes := make([]int, len(fehlgeschlagen))
		for i, fehlDatei := range fehlgeschlagen {
			for j, originalDatei := range dateien {
				if fehlDatei == originalDatei {
					fehlgeschlagenIndizes[i] = j
					break
				}
			}
		}
		
		// Führe die fehlgeschlagenen Dateien erneut aus
		fehlgeschlagen = prozessiereDateien(fehlgeschlagen, fehlgeschlagenIndizes, true)
	}
	
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
	
	// Zeige übersprungene nicht existierende Dateien
	if len(nichtExistierendeDateien) > 0 {
		fmt.Printf("\nÜbersprungene nicht existierende Dateien (%d):\n", len(nichtExistierendeDateien))
		for i, datei := range nichtExistierendeDateien {
			fmt.Printf("%d. %s\n", i+1, datei)
		}
	}
	
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
		
		// Wenn nach allen Wiederholungsversuchen immer noch Fehler bestehen, pausiere das Programm
		if len(fehlgeschlagen) > 0 {
			fmt.Println("\n\nNach allen Wiederholungsversuchen bestehen immer noch Fehler.")
			fmt.Println("Drücken Sie Enter, um das Programm zu beenden...")
			
			// Warte auf Benutzereingabe
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
		
		os.Exit(1)
	} else {
		if len(nichtExistierendeDateien) > 0 {
			fmt.Println("\nAlle vorhandenen Extraktionen erfolgreich abgeschlossen!")
			fmt.Printf("(Es wurden %d nicht existierende Dateien übersprungen)\n", len(nichtExistierendeDateien))
		} else {
			fmt.Println("\nAlle Extraktionen erfolgreich abgeschlossen!")
		}
		fmt.Printf("Gesamtzeit: %s\n", gesamtZeit)
	}
}
