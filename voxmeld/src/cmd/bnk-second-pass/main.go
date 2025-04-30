package main

import (
	"bytes"
	"encoding/binary"

	//	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <org_input.bnk>\n", os.Args[0])
	}
	bnkPath := os.Args[1]

	bnk, err := os.ReadFile(bnkPath)
	if err != nil {
		panic(err)
	}

	pattern := []byte{0x01, 0x00, 0x14, 0x00}  // Codec: OPUS_WEM
	newCodec := []byte{0x01, 0x00, 0x04, 0x00} // Codec: VORBIS
	// Find the pattern in the file:
	// Quick and dirty approach to patch the BNKs
	pos := bytes.Index(bnk, pattern)
	if pos == -1 {
		pos = bytes.Index(bnk, newCodec)
		if pos == -1 {
			panic("Pattern not found")
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
	wemPath := "german-voices-oblivion-remastered-voxmeld_v0.3.2_P/Content/WwiseAudio/Media/English(US)/" + strconv.Itoa(int(id)) + ".wem"
	// Get size of the .wem file
	wemInfo, err := os.Stat(wemPath)
	if err != nil {
		log.Fatalf("Skipping BNK second pass. No matching WEM found in output: %v\n", err)
	}
	wemSize := uint32(wemInfo.Size())
	// Update the codex to VORBIS
	copy(bnk[pos:pos+4], newCodec)
	// Update file size (4 bytes after dummy byte and ID)
	fileSizeOffset := pos + 9
	if fileSizeOffset+4 > len(bnk) {
		log.Fatalf("Not enough data to update file size in .bnk")
	}
	binary.LittleEndian.PutUint32(bnk[fileSizeOffset:fileSizeOffset+4], wemSize)
	// write the modified .bnk file to the output folder
	outBnkPath := "german-voices-oblivion-remastered-voxmeld_v0.3.2_P/Content/WwiseAudio/Event/English(US)/" + filepath.Base(bnkPath)
	err = os.WriteFile(outBnkPath, bnk, 0644)
	if err != nil {
		log.Fatalf("Failed to write modified .bnk file: %v\n", err)
	}
}
