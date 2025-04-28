#!/bin/bash
# Change the prefix to the english variant - important for further processing
mv tmp/sound/voice/oblivion.esm/argonier tmp/sound/voice/oblivion.esm/argonian
mv tmp/sound/voice/oblivion.esm/hochelf tmp/sound/voice/oblivion.esm/high_elf
mv tmp/sound/voice/oblivion.esm/kaiserlicher tmp/sound/voice/oblivion.esm/imperial
mv tmp/sound/voice/oblivion.esm/rothwardone tmp/sound/voice/oblivion.esm/redguard
# Same for Knights DLC
mv tmp/sound/voice/knights.esp/argonier tmp/sound/voice/knights.esp/argonian
mv tmp/sound/voice/knights.esp/hochelf tmp/sound/voice/knights.esp/high_elf
mv tmp/sound/voice/knights.esp/kaiserlicher tmp/sound/voice/knights.esp/imperial
mv tmp/sound/voice/knights.esp/rothwardone tmp/sound/voice/knights.esp/redguard
# Same for Shivering Isles DLC
mv tmp/sound/voice/oblivion.esm/dunkler* tmp/sound/voice/oblivion.esm/dark_seducer
mv tmp/sound/voice/oblivion.esm/goldener* tmp/sound/voice/oblivion.esm/golden_saint

# Check if bnk with matching prefix exists and copy to the MP3 to WEM input folder
check_and_copy_mp3() {
	filename="${3##*/}"
	filename="${filename%.mp3}"
	counter=0
	
    echo "Check: ${1}_${2}_${3}..."
	
	# Check for normal, altvoice and beggar variant
	for prefix in "" "_altvoice" "_beggar"
	do
		# Check if bnk matching the mp3s name exists to prevent unused mp3 copies
		if [ -f "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_${1}_${2}${prefix}_${filename}.bnk" ]; then
			echo "Found BNK: (${1}_${2}${prefix}_${3}) -> Copy MP3..."
			cp "$4" "tmp/MP3s/${1}_${2}${prefix}_${3}" &
		else
			counter=$((counter + 1))
		fi
	done
	
	# Check if mp3 was used, write to log file if not
	if [ "$counter" -eq 3 ]; then
		echo "No matching bnk found!"
		echo "${1}_${2}_${3}" >> notused.txt
	fi
}

# If bnk counterpart is found copy mp3s to the folder with files to convert
# Also add files for alt races that don't have individual VO 
mkdir -p tmp/MP3s/
for dlc in tmp/sound/voice/*
do
	for race in $dlc/*
	do
		for variant in $race/*
		do
			for file in $variant/*.mp3
			do
				# Check if mp3 has bnk counterpart
				check_and_copy_mp3 "${race##*/}" "${variant##*/}" "${file##*/}" "$file"
				
				# Check if mp3 has alternative race bnk counterpart
				case "${race##*/}" in
                    "argonian")
                        check_and_copy_mp3 "khajiit" "${variant##*/}" "${file##*/}" "$file"
                        ;;
                    "high_elf")
                        check_and_copy_mp3 "dark_elf" "${variant##*/}" "${file##*/}" "$file"
                        check_and_copy_mp3 "wood_elf" "${variant##*/}" "${file##*/}" "$file"
                        ;;
                    "imperial")
                        check_and_copy_mp3 "breton" "${variant##*/}" "${file##*/}" "$file"
                        ;;
                    "nord")
                        check_and_copy_mp3 "orc" "${variant##*/}" "${file##*/}" "$file"
                        ;;
                    *)
                        ;;
                esac
			done
		done
	done
done

wait