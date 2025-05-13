#!/bin/bash
# Checks each bnk file for a counterpart in the folder with wem files
txtpath="logs/dev"
txtname="missing-bnk.txt"

if [ -d "${txtpath}" ]; then
	if [ -f "${txtpath}/${txtname}" ]; then
		rm "${txtpath}/${txtname}"
	fi
else
	mkdir "${txtpath}"
fi

for bnkfile in "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/"*
do
	filename="${bnkfile##*/}"
	
    if [ ! -f "tmp\bnk\Content\WwiseAudio\Event\English(US)/${filename}" ]; then
		# No bnk file was found that matches the name of a remaster bnk file, add missing bnk name to log file
		echo "Not found: ${filename}"
		echo "${filename}" >> "${txtpath}/${txtname}"
	else
		echo "Found: ${filename}"
	fi
done