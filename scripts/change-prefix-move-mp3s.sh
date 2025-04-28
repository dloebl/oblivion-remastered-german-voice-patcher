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

# copy all audio files to the MP3 to WEM input folder
mkdir -p tmp/MP3s/
for dlc in tmp/sound/voice/*
do
    for race in $dlc/*
    do
        for variant in $race/*
        do
            for file in $variant/*.mp3
            do
                cp $file tmp/MP3s/"${race##*/}_${variant##*/}_${file##*/}" &
            done
        done
    done
done