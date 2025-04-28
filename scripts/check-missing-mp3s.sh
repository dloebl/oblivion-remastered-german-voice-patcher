#!/bin/bash
mkdir -p tmp/MP3s/
# Checks each bnk file for a counterpart in the folder with mp3 files to convert
for bnkfile in "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/"*
do
	filename="${bnkfile##*/}"
	filename="${filename/Play_/}"
	filename="${filename%.bnk}"
	
    if [ ! -f "tmp/MP3s/${filename}.mp3" ]; then
		# No mp3 file was found that matches the name of a bnk file, add missing mp3 name to log file
		echo "${filename}.mp3" >> missing.txt
	fi
done