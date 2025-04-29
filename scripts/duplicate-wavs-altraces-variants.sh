#!/bin/bash
# Check if bnk with matching prefix exists and creates a renamed duplicat for variants
# Vars: 1. Race prefix, 2. Gender prefix, 3. Event, 4.File
create_alt_copies() {
	event="${3%.wav}"
	
	# Check for normal, altvoice and beggar variant
	for prefix in "" "_altvoice" "_beggar"
	do
		echo "Check BNK: Play_${1}_${2}${prefix}_${event}.bnk..."

		# Check if bnk matching the wav name exists 
		if [ -f "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/Play_${1}_${2}${prefix}_${event}.bnk" ]; then
			echo "Found BNK: Play_${1}_${2}${prefix}_${event} -> Duplicate wav..."
			cp "$4" "sound2wem/audiotemp/${1}_${2}${prefix}_${event}.wav" &
		fi
	done
}

# If bnk counterpart is found copy wavs to the final audiotemp folder
# Also checks and adds duplicates for variants
for file in sound2wem/audiotempconvert/*.wav
do
	filename="${file##*/}"
	race=$(echo "$filename" | cut -d'_' -f1)
	gender=$(echo "$filename" | cut -d'_' -f2)
	event=$(echo "$filename" | cut -d'_' -f3-)

	create_alt_copies "${race}" "${gender}" "${event}" "$file"
	
	# Check if mp3 has alternative race bnk counterpart
	case "${race}" in
		"argonian")
			create_alt_copies "khajiit" "${gender}" "${event}" "$file"
			;;
		"high_elf")
			create_alt_copies "dark_elf" "${gender}" "${event}" "$file"
			create_alt_copies "wood_elf" "${gender}" "${event}" "$file"
			;;
		"imperial")
			create_alt_copies "breton" "${gender}" "${event}" "$file"
			;;
		"nord")
			create_alt_copies "orc" "${gender}" "${event}" "$file"
			;;
		*)
			;;
	esac
done