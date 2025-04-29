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

# Check if remaster .bsa extract folder includes mp3 or variant of mp3 and replace it with german version
# This should fix cut off dialogs
# Vars: 1. Dlc prefix, 2. Race prefix, 3. Gender prefix, 4. File
check_and_copy_remaster_bsa() {
	for prefix in "" "/altvoice" "/beggar"
	do
		# Check if remaster .bsa extract includes the mp3s or a variant of it
		if [ -f "german-voices-oblivion-remastered-voxmeld_v0.2.2_P/sound/voice/${1}/${2}/${3}${prefix}/${4##*/}" ]; then
			echo "Copy variant: ${1}/${2}/${3}${prefix}/${4##*/}..."
			cp "$4" "german-voices-oblivion-remastered-voxmeld_v0.2.2_P/sound/voice/${1}/${2}/${3}${prefix}/${4##*/}" &
		fi
	done
}

mkdir -p tmp/MP3s/
for dlc in tmp/sound/voice/*
do
	for race in $dlc/*
	do
		for variant in $race/*
		do
			for file in $variant/*.mp3
			do
				# Copy file to convert folder
				echo "Copy: ${race##*/}_${variant##*/}_${file##*/}..."
				cp "$file" "tmp/MP3s/${race##*/}_${variant##*/}_${file##*/}" &

				# Check for alternative variants and copy them to .bsa extract folder
				check_and_copy_remaster_bsa "${dlc##*/}" "${race##*/}" "${variant##*/}" "$file" 
				
				# Check if mp3 has alternative race
				case "${race##*/}" in
                    "argonian")
						check_and_copy_remaster_bsa "${dlc##*/}" "khajiit" "${variant##*/}" "$file" 
                        ;;
                    "high_elf")
						check_and_copy_remaster_bsa "${dlc##*/}" "dark_elf" "${variant##*/}" "$file" 
						check_and_copy_remaster_bsa "${dlc##*/}" "wood_elf" "${variant##*/}" "$file" 
                        ;;
                    "imperial")
						check_and_copy_remaster_bsa "${dlc##*/}" "breton" "${variant##*/}" "$file" 
                        ;;
                    "nord")
						check_and_copy_remaster_bsa "${dlc##*/}" "orc" "${variant##*/}" "$file" 
                        ;;
                    *)
                        ;;
                esac
			done
		done
	done
done