package main

import (
	"bytes"
	"encoding/binary"
	"strings"

	//	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <input.wem>\n", os.Args[0])
	}
	wemPath := os.Args[1]
	comps := strings.Split(filepath.Base(wemPath), "_")

	// File is audio for video
	if len(comps) < 2 {
		videoAudio := comps[0]
		// TODO: Match bik to wems
		var mappingTable = map[string]string{
			"abc":    "abc",
		}
		
		if newName, exists := mappingTable[videoAudio]; exists {
			log.Printf("Found mapping entry for '%s'.", videoAudio)

			outWemPath := "german-voices-oblivion-remastered-voxmeld_v0.3.1_P/Content/WwiseAudio/Media/English(US)/" + newName + ".wem"
			wem, err := os.ReadFile(wemPath)
			if err != nil {
				log.Fatalf("Error while reading .wem file: %v\n", err)
			}

			// Move video audio to media folder
			err = os.WriteFile(outWemPath, wem, 0644)
			if err != nil {
				log.Fatalf("Error while writing .wem file: %v\n", err)
			}
		} else {
			log.Printf("Could not find mapping entry for '%s'.", videoAudio)
		}

		return
	}

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
			bnkPath := "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_" + bnkName + ".bnk"
			log.Println("File:", bnkPath)

			bnk, err := os.ReadFile(bnkPath)
			if err != nil {
				log.Printf("Skipping missing BNK: %s (err: %v)", bnkPath, err)
				continue
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

			// Get size of the .wem file
			wemInfo, err := os.Stat(wemPath + ".wem")
			if err != nil {
				log.Fatalf("Failed to read .wem file: %v\n", err)
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
			outBnkPath := "german-voices-oblivion-remastered-voxmeld_v0.3.1_P/Content/WwiseAudio/Event/English(US)/" + filepath.Base(bnkPath)
			err = os.WriteFile(outBnkPath, bnk, 0644)
			if err != nil {
				log.Fatalf("Failed to write modified .bnk file: %v\n", err)
			}
			//fmt.Printf("Modified .bnk file written to: %s\n", outBnkPath)
			// write the .wem file to output folder
			wem, err := os.ReadFile(wemPath + ".wem")
			if err != nil {
				log.Fatalf("Failed to read .wem file: %v\n", err)
			}
			outWemPath := "german-voices-oblivion-remastered-voxmeld_v0.3.1_P/Content/WwiseAudio/Media/English(US)/" + strconv.Itoa(int(id)) + ".wem"
			err = os.WriteFile(outWemPath, wem, 0644)
			if err != nil {
				log.Fatalf("Failed to write .wem file: %v\n", err)
			}
			//fmt.Printf("Extracted .wem file written to: %s\n", outWemPath)
		}
	}
}