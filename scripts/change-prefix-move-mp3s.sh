#!/bin/bash
# Change the prefix to the english variant - important for further processing
mv tmp/sound/voice/oblivion.esm/argonier tmp/sound/voice/oblivion.esm/argonian
mv tmp/sound/voice/oblivion.esm/hochelf tmp/sound/voice/oblivion.esm/high_elf
mv tmp/sound/voice/oblivion.esm/kaiserlicher tmp/sound/voice/oblivion.esm/imperial
mv tmp/sound/voice/oblivion.esm/rothwardone tmp/sound/voice/oblivion.esm/redguard
# copy all audio files to the MP3 to WEM input folder
mkdir -p tmp/MP3s/
for race in tmp/sound/voice/oblivion.esm/*
do
    for variant in $race/*
    do
        for file in $variant/*.mp3
        do
            cp $file tmp/MP3s/"${race##*/}_${variant##*/}_${file##*/}" &
        done
    done
done