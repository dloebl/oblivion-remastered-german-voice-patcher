#!/bin/bash
# this is there the new BNKs go
mkdir -p "german-voices-oblivion-remastered-voxmeld_v0.3.0_P/Content/WwiseAudio/Event/English(US)/"
# this is there the new WEMs go
mkdir -p "german-voices-oblivion-remastered-voxmeld_v0.3.0_P/Content/WwiseAudio/Media/English(US)/"
# loop over all WEMs that we managed to convert to
find sound2wem/Windows/ -name "*.wem" | sed "s/\.wem$//" | xargs -P 64 -I {} ./voxmeld/voxmeld.exe {}
# second pass on original BNKs
find "tmp/pak/OblivionRemastered/Content/WwiseAudio/Event/English(US)/" -name "*.bnk" | xargs -P 64 -I {} ./voxmeld/bnk-second-pass.exe {}