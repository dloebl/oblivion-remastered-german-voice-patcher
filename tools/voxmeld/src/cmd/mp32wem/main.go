// mp32wem converts audio files into Wwise Opus (.wem) files without needing
// the Audiokinetic Wwise authoring tool.
//
// Background: Oblivion Remastered stores its voice lines as OPUS_WEM
// (fmt tag 0x3041), which - unlike Wwise Vorbis - is plain, standard Opus.
// A WEM is just a RIFF container holding
//
//	fmt   0x3041, 1 channel, 48000 Hz, 18 bytes of extra data
//	hash  16 bytes (only used by the authoring tool)
//	seek  one uint16 per Opus packet, holding that packet's size in bytes
//	data  the raw Opus packets, concatenated, no Ogg framing
//
// So the conversion is: encode with ffmpeg/libopus to Ogg Opus, unwrap the
// Ogg pages back into packets, and write the RIFF container around them.
//
// Usage mirrors the sound2wem tool it replaces:
//
//	mp32wem.exe "tmp\MP3s\*"
package main

import (
	"bytes"
	"encoding/binary"
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

const (
	sampleRate   = 48000
	frameSamples = 960 // 20 ms at 48 kHz
)

// Defaults, overridable with -bitrate and -gain.
//
// bitrate: the sources are 64 kbit/s mono MP3, so nothing above that can be
// recovered - the only thing more bits buy is that this second lossy stage
// adds less damage on top. Measured signal-to-distortion against the decoded
// source: 64k = 17.9 dB, 96k = 21.4 dB, 128k = 25.4 dB.
//
// gain: the German Oblivion was mastered louder than the Remaster. Measured
// over 200 randomly picked lines against their English counterparts, the
// German files run +4.75 dB (median +4.6) hotter, and 92 % of them are louder.
// The volume settings inside the BNKs are tuned for the English levels and we
// keep them as they are, so without this correction German dialogue sits too
// hot against music and effects.
var (
	bitrate   = flag.String("bitrate", "96k", "Opus Zielbitrate, z.B. 64k, 96k, 128k")
	gainDB    = flag.Float64("gain", -4.6, "Pegelkorrektur in dB, 0 = unveraendert")
	outputDir = flag.String("out", "tmp/wem", "Zielordner fuer die .wem Dateien")
	ffmpegBin = flag.String("ffmpeg", "", "Pfad zu ffmpeg.exe (leer = automatisch suchen)")
)

var (
	converted int32
	failed    int32
)

// oggStream is the result of unwrapping an Ogg Opus bitstream.
type oggStream struct {
	packets     [][]byte // audio packets only, OpusHead/OpusTags removed
	preSkip     uint16
	finalGranue uint64
}

// parseOgg walks the Ogg pages and reassembles the packets inside them.
func parseOgg(data []byte) (*oggStream, error) {
	st := &oggStream{}
	var pending []byte // packet spanning multiple pages
	pos := 0

	for pos+27 <= len(data) {
		if !bytes.Equal(data[pos:pos+4], []byte("OggS")) {
			return nil, fmt.Errorf("no OggS capture pattern at offset %d", pos)
		}
		granule := binary.LittleEndian.Uint64(data[pos+6 : pos+14])
		segCount := int(data[pos+26])
		segTableOff := pos + 27
		if segTableOff+segCount > len(data) {
			return nil, fmt.Errorf("truncated segment table at offset %d", pos)
		}
		segTable := data[segTableOff : segTableOff+segCount]

		bodyOff := segTableOff + segCount
		bodyLen := 0
		for _, s := range segTable {
			bodyLen += int(s)
		}
		if bodyOff+bodyLen > len(data) {
			return nil, fmt.Errorf("truncated page body at offset %d", pos)
		}
		body := data[bodyOff : bodyOff+bodyLen]

		// A packet is made of consecutive segments and ends with a segment
		// shorter than 255 bytes.
		segPos := 0
		for _, s := range segTable {
			pending = append(pending, body[segPos:segPos+int(s)]...)
			segPos += int(s)
			if s < 255 {
				st.packets = append(st.packets, pending)
				pending = nil
			}
		}

		st.finalGranue = granule
		pos = bodyOff + bodyLen
	}

	if len(st.packets) < 3 {
		return nil, fmt.Errorf("stream holds only %d packets", len(st.packets))
	}
	head := st.packets[0]
	if len(head) < 12 || !bytes.HasPrefix(head, []byte("OpusHead")) {
		return nil, fmt.Errorf("first packet is not an OpusHead")
	}
	st.preSkip = binary.LittleEndian.Uint16(head[10:12])

	// Drop OpusHead and OpusTags - only the audio packets go into the WEM.
	st.packets = st.packets[2:]
	return st, nil
}

// buildWem assembles the RIFF/WAVE container around the Opus packets.
func buildWem(st *oggStream) ([]byte, error) {
	var data bytes.Buffer
	seek := make([]byte, 0, len(st.packets)*2)
	for _, p := range st.packets {
		if len(p) > 0xFFFF {
			return nil, fmt.Errorf("packet of %d bytes does not fit the uint16 seek table", len(p))
		}
		seek = binary.LittleEndian.AppendUint16(seek, uint16(len(p)))
		data.Write(p)
	}

	// The granule position of the last page counts decoded samples including
	// the encoder delay, so the real length is granule - preSkip.
	var samples uint32
	if st.finalGranue > uint64(st.preSkip) {
		samples = uint32(st.finalGranue - uint64(st.preSkip))
	}
	if samples == 0 {
		return nil, fmt.Errorf("stream decodes to 0 samples")
	}

	var avgBytesPerSec uint32
	if samples > 0 {
		avgBytesPerSec = uint32(uint64(data.Len()) * sampleRate / uint64(samples))
	}

	var fmtChunk bytes.Buffer
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(0x3041)) // OPUS_WEM
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(1))      // mono
	binary.Write(&fmtChunk, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&fmtChunk, binary.LittleEndian, avgBytesPerSec)
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(0)) // nBlockAlign
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(0)) // wBitsPerSample
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(18))
	// 18 bytes of Wwise specific extra data
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(frameSamples))
	fmtChunk.WriteByte(0x01)
	fmtChunk.WriteByte(0x41)
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(0))
	binary.Write(&fmtChunk, binary.LittleEndian, samples)
	binary.Write(&fmtChunk, binary.LittleEndian, uint32(len(st.packets)))
	binary.Write(&fmtChunk, binary.LittleEndian, st.preSkip)
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(1))

	var body bytes.Buffer
	body.WriteString("WAVE")
	writeChunk(&body, "fmt ", fmtChunk.Bytes())
	writeChunk(&body, "hash", make([]byte, 16))
	writeChunk(&body, "seek", seek)
	writeChunk(&body, "data", data.Bytes())

	var out bytes.Buffer
	out.WriteString("RIFF")
	binary.Write(&out, binary.LittleEndian, uint32(body.Len()))
	out.Write(body.Bytes())
	return out.Bytes(), nil
}

// writeChunk appends a RIFF chunk, padded to an even length as the spec requires.
func writeChunk(w *bytes.Buffer, id string, payload []byte) {
	w.WriteString(id)
	binary.Write(w, binary.LittleEndian, uint32(len(payload)))
	w.Write(payload)
	if len(payload)%2 == 1 {
		w.WriteByte(0)
	}
}

func convert(ffmpeg, inputFile, outputFile string) error {
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", inputFile,
		"-vn",
	}
	if *gainDB != 0 {
		args = append(args, "-af", fmt.Sprintf("volume=%.2fdB", *gainDB))
	}
	args = append(args,
		"-ac", "1",
		"-ar", fmt.Sprint(sampleRate),
		"-c:a", "libopus",
		"-b:a", *bitrate,
		"-vbr", "on",
		"-application", "audio",
		"-f", "opus",
		"-")
	cmd := exec.Command(ffmpeg, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("ffmpeg: %v (%s)", err, msg)
		}
		return fmt.Errorf("ffmpeg: %v", err)
	}

	st, err := parseOgg(stdout.Bytes())
	if err != nil {
		return err
	}
	wem, err := buildWem(st)
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, wem, 0644)
}

func findFfmpeg(execDir string) (string, error) {
	if *ffmpegBin != "" {
		if _, err := os.Stat(*ffmpegBin); err != nil {
			return "", fmt.Errorf("ffmpeg nicht gefunden unter %s", *ffmpegBin)
		}
		return *ffmpegBin, nil
	}
	candidates := []string{
		filepath.Join(execDir, "..", "sound2wem", "ffmpeg-master-latest-win64-gpl-shared", "bin", "ffmpeg.exe"),
		filepath.Join("tools", "sound2wem", "ffmpeg-master-latest-win64-gpl-shared", "bin", "ffmpeg.exe"),
		filepath.Join("sound2wem", "ffmpeg-master-latest-win64-gpl-shared", "bin", "ffmpeg.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("ffmpeg.exe not found - expected it below sound2wem/")
}

func main() {
	start := time.Now()

	flag.Parse()
	patterns := flag.Args()
	if len(patterns) == 0 {
		fmt.Println("Fehler: Keine Eingabedateien angegeben")
		flag.Usage()
		os.Exit(1)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Fehler beim Ermitteln des Ausfuehrungsverzeichnisses: %v\n", err)
		os.Exit(1)
	}
	ffmpeg, err := findFfmpeg(filepath.Dir(exe))
	if err != nil {
		fmt.Printf("Fehler: %v\n", err)
		os.Exit(1)
	}

	var files []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		files = append(files, matches...)
	}
	if len(files) == 0 {
		fmt.Println("Fehler: Keine passenden Dateien gefunden")
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Fehler beim Erstellen von %s: %v\n", (*outputDir), err)
		os.Exit(1)
	}

	fmt.Printf("\n====================== MP32WEM ======================\n")
	fmt.Printf("Quelle:      %s\n", strings.Join(patterns, ", "))
	fmt.Printf("Dateien:     %d Audio-Dateien gefunden\n", len(files))
	fmt.Printf("Ziel:        %s (Wwise Opus, kein Wwise noetig)\n", *outputDir)
	fmt.Printf("Bitrate:     %s\n", *bitrate)
	fmt.Printf("Pegel:       %+.1f dB\n", *gainDB)
	fmt.Printf("-----------------------------------------------------\n")

	workers := runtime.NumCPU()
	fmt.Printf("Verwende %d parallele Prozesse\n", workers)

	stop := make(chan struct{})
	go progress(len(files), start, stop)

	jobs := make(chan string, workers*4)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for in := range jobs {
				base := filepath.Base(in)
				out := filepath.Join(*outputDir, strings.TrimSuffix(base, filepath.Ext(base))+".wem")
				if err := convert(ffmpeg, in, out); err != nil {
					atomic.AddInt32(&failed, 1)
					fmt.Printf("\rFehler bei %s: %v\n", base, err)
				} else {
					atomic.AddInt32(&converted, 1)
				}
			}
		}()
	}
	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()

	close(stop)
	time.Sleep(150 * time.Millisecond)

	fmt.Printf("\n\nFertig in %s\n", time.Since(start).Round(time.Second))
	fmt.Printf("Erstellte .wem Dateien: %d\n", atomic.LoadInt32(&converted))
	fmt.Printf("Fehlgeschlagen:         %d\n", atomic.LoadInt32(&failed))
}

func progress(total int, start time.Time, stop <-chan struct{}) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	spin := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-ticker.C:
			done := int(atomic.LoadInt32(&converted) + atomic.LoadInt32(&failed))
			width := 40
			ratio := float64(done) / float64(total)
			filled := int(ratio * float64(width))
			bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
			if filled < width {
				b := []rune(bar)
				b[filled] = []rune(spin[i%len(spin)])[0]
				bar = string(b)
			}
			i++
			fmt.Printf("\r[%s] %3.0f%% %d/%d Dateien", bar, ratio*100, done, total)
		case <-stop:
			return
		}
	}
}
