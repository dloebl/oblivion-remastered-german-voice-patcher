# Discord Server
Join the Discord server to stay up to date on the future development of this patcher:
https://discord.gg/TVfn6xkBhB

# Important notes
- You have to own both Oblivion Remastered as well as the original German version of Oblivion
- Some voices are still in English. This will be improved in a future version
- The build process is quite slow right now (some parts are still very inefficent)
- Lipsync still uses the English version
- Further development of this mod will take place here on GitHub and mirrored to Nexus Mods: https://www.nexusmods.com/oblivionremastered/mods/1092

# Requirements
- Windows, plus enough free disk space for the temporary files
- Keep the patcher in a short path such as `C:\obre-de\`. Windows caps paths at 260 characters and the extraction tools drop everything past that limit without reporting an error

Wwise is no longer required. The patcher writes the Wwise Opus files itself using the bundled FFmpeg.

# Audio settings
Optional, both can be set in `config\settings.txt`:

| Setting | Default | Meaning |
| --- | --- | --- |
| `OPUS_BITRATE` | `96k` | Opus target bitrate. The sources are 64 kbit/s mono MP3, so more bits recover nothing - they only keep this encoding step from adding damage on top. Drives the mod size: 64k ≈ 3.9 GB, 96k ≈ 5.6 GB, 128k ≈ 7.6 GB |
| `OPUS_GAIN_DB` | `-4.6` | Level correction. The German Oblivion was mastered louder than the Remaster - measured over 200 lines the German files run +4.75 dB hotter than their English counterparts, while the volume settings inside the BNKs are tuned for the English levels. Set to `0` to keep the original level |

# Build steps and installation steps
1. Run the "Create-Mod.bat" script
2. Enter the paths the terminal asks you for and confirm them with 'Enter'
3. Be patient. Building the mod takes a while.
4. Some errors like "panic: open [..]/Event/English(US)/Play_*.bnk are expected. As long as the patcher doesn't get stuck for multiple minutes without anything changing you don't ahve to worry.
5. Install the mod by confirming the prompt at the end or copy the the files from the built "Mod\" folder to your installation of Oblivion Remastered
6. Enjoy Oblivion Remastered with German voices!

# Open Source credits
The following open source software is used during the build process of the mod. A big thank you to the original authors!
- BSArch: https://github.com/TES5Edit/TES5Edit/tree/dev/Tools/BSArchive
- FFmpeg: https://github.com/FFmpeg/FFmpeg
- Repak: https://github.com/trumank/repak

The following tools also have been very useful during development of this mod:
- wwiser: https://github.com/bnnm/wwiser
- sound2wem: https://github.com/EternalLeo/sound2wem
- foobar2000: https://www.foobar2000.org/
- vgmstream: https://github.com/vgmstream/vgmstream
- hexer: https://gitlab.com/hexer/hexer
- ReadyOrNot UE5 modding guide: https://unofficial-modding-guide.com/posts/thebasics/#creating-a-pak-file
