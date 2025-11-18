<div align="center">
    <img src=".github/resources/game_manager_logo.png" width="auto" alt="Game Manager Logo">
<h3 style="font-size: 25px; margin-top:0px; padding-top: 0px;">
    NextUI Game Manager
</h3>

<h4 style="font-size: 18px;">
A one-stop pak for all your game management needs.
</h4>

## [Download this in the Pak Store!](https://github.com/UncleJunVIP/nextui-pak-store)

![GitHub License](https://img.shields.io/github/license/UncleJunVip/nextui-game-manager?style=for-the-badge&color=007C77)
![GitHub Release](https://img.shields.io/github/v/release/UncleJunVIP/nextui-game-manager?sort=semver&style=for-the-badge&color=007C77)
![GitHub Repo stars](https://img.shields.io/github/stars/UncleJunVip/nextui-game-manager?style=for-the-badge&color=007C77)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/UncleJunVIP/nextui-game-manager/total?style=for-the-badge&label=Total%20Downloads&color=007C77)


</div>

---

# Features

- Create / Rename / Delete Collections
- Add / Remove Games from Collections (Single and Multiple Selection)
- Rename ROM
    - Renames Art and Associated Save File
- Download Art from the Libretro Thumbnail Project (Single and Multiple Selection)
    - Can configure what type of art you would like to download in the Game Manager Settings
    - Searches first for exact match and then uses `Jaccard Similarity` with a configurable threshold
    - The Libretro Thumbnail Project has Box Art, Title Screens, Screenshots and Logos
- Delete Art (Single and Multiple Selection)
- Archive ROM (Places ROM and Art if present into a hidden folder)
- Manage ROM Archives (Rename archive folder names and restore archived ROMs)
- Delete ROM (Deletes ROM file and associated Art)
- Global Actions
    - Download all missing art
        - Ability to download by platform
    - Clear recently played list

---

## How do I install Game Manager?

*Preferred Method:* [NextUI Pak Store](https://github.com/UncleJunVIP/nextui-pak-store)

1. Own a TrimUI Brick or Smart Pro and have a SD Card with NextUI configured.
2. Connect your device to a Wi-Fi network. (This is needed for Artwork Download Support)
3. Download the latest Game Manager release from this repo.
4. Unzip the release download.
5. With your device powered off, eject the SD Card and connect it to your computer.
6. Copy the entire Game Manager.pak file to `SD_ROOT/Tools/tg5040`.
7. Reinsert your SD Card into your device.
8. Launch `Game Manager` from the `Tools` menu and enjoy!

✌️
