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
1. Install the Audiokinetic Wwise Launcher: https://www.audiokinetic.com/en/wwise/overview/. You'll have to create an Audiokinetic account for this. The free trial version is sufficient - you don't have to purchase a license for this to work
2. Start the "Wwise Launcher", login and install the latest version of Wwise. You can unselect all optional features - we just need Wwise
3. Install the Unreal Engine 5 through the Epic Games Store launcher. This is required to unpack the .pak file from Oblivion Remastered (you'll have to create an Epic Games account for this)

# Build steps and installation steps
1. Open "paths.bat" with a text editor and update the three paths at the beginning of the file with the updated ones from the requirements that you just installed
2. Run the "build_oblivion_remastered_german_voices.bat" script
3. Be patient. It takes about 60 minutes to build this mod.
4. Some errors like "panic: open [..]/Event/English(US)/Play_*.bnk are expected and just mean that a German voice file couldn't be mapped to a BNK
5. Copy the the files from the built "ModFiles\" folder to your installation of Oblivion Remastered
6. Enjoy Oblivion Remastered with German voices!

# Open Source credits
The following open source software is used during the build process of the mod. A big thank you to the original authors!
- BSArch: https://github.com/TES5Edit/TES5Edit/tree/dev/Tools/BSArchive
- busybox-w32: https://github.com/rmyorston/busybox-w32
- FFmpeg: https://github.com/FFmpeg/FFmpeg

The following tools also have been very useful during development of this mod:
- wwiser: https://github.com/bnnm/wwiser
- sound2wem: https://github.com/EternalLeo/sound2wem
- foobar2000: https://www.foobar2000.org/
- vgmstream: https://github.com/vgmstream/vgmstream
- hexer: https://gitlab.com/hexer/hexer
- ReadyOrNot UE5 modding guide: https://unofficial-modding-guide.com/posts/thebasics/#creating-a-pak-file
