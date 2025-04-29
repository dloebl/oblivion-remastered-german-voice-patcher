#!/bin/bash
# Checks each bnk file for a counterpart in the folder with wem files
for bnkfile in "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/"*
do
	filename="${bnkfile##*/}"
	filename="${filename/Play_/}"
	filename="${filename%.bnk}"
	
    if [ ! -f "sound2wem/WindowsFinal/${filename}.wem" ]; then
		# No wem file was found that matches the name of a bnk file, add missing wem name to log file
		echo "${filename}.wem" >> missing.txt
	fi
done