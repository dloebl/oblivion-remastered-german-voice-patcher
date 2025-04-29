#!/bin/bash
# Checks each bnk file for a counterpart in the folder with wem files
for bnkfile in "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/"*
do
	filename="${bnkfile##*/}"
	
    if [ ! -f "german-voices-oblivion-remastered-voxmeld_v0.3.0_P\Content\WwiseAudio\Event\English(US)/${filename}" ]; then
		# No bnk file was found that matches the name of a remaster bnk file, add missing bnk name to log file
		echo "${filename}.bnk" >> missing.txt
	fi

	
done