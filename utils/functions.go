package utils

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

const gameTrackerDBPath = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"

const saveFileDirectory = "/mnt/SDCARD/Saves/"

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}

	return common.RomDirectory
}

func GetArchiveRoot() string {
	if IsDev() {
		return os.Getenv("ARCHIVE_DIRECTORY")
	}

	return "/mnt/SDCARD/Roms/.Archive"
}

func GetCollectionDirectory() string {
	if IsDev() {
		_ = MakeDirectoryIfNotExist(os.Getenv("COLLECTION_DIRECTORY"))
		return os.Getenv("COLLECTION_DIRECTORY")
	}

	_ = MakeDirectoryIfNotExist(common.CollectionDirectory)
	return common.CollectionDirectory
}

func GetSaveFileDirectory() string {
	if IsDev() {
		return os.Getenv("SAVE_FILE_DIRECTORY")
	}

	return saveFileDirectory
}

func GetGameTrackerDBPath() string {
	if IsDev() {
		return os.Getenv("GAME_TRACKER_DB_PATH")
	}

	return gameTrackerDBPath
}

func GetFileList(dirPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	return entries, nil
}

func FilterList(itemList []shared.Item, keywords ...string) []shared.Item {
	var filteredItemList []shared.Item

	for _, item := range itemList {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				filteredItemList = append(filteredItemList, item)
				break
			}
		}
	}

	return filteredItemList
}

func InsertIntoSlice(s []string, index int, values ...string) []string {
	if index < 0 {
		index = 0
	}
	if index > len(s) {
		index = len(s)
	}

	return append(s[:index], append(values, s[index:]...)...)
}

func FindExistingArt(selectedFile string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	mediaDir := filepath.Join(romDirectory.Path, ".media")

	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		logger.Info("No media directory found", zap.String("current_directory", romDirectory.Path))
		return "", nil
	}

	artDir := filepath.Join(romDirectory.Path, ".media")
	artList, err := GetFileList(artDir)
	if err != nil {
		logger.Error("failed to list arts", zap.Error(err))
		return "", err
	}

	artFilename := ""

	artFilenameNoExtension := strings.ReplaceAll(selectedFile, filepath.Ext(selectedFile), "")

	for _, art := range artList {
		if strings.ReplaceAll(art.Name(), filepath.Ext(art.Name()), "") == artFilenameNoExtension {
			artFilename = art.Name()
			break
		}
	}

	if artFilename == "" {
		return "", nil
	}

	return filepath.Join(artDir, artFilename), err
}

func FindArt(romDirectory shared.RomDirectory, game shared.Item, downloadType sum.Int[shared.ArtDownloadType]) string {
	logger := common.GetLoggerInstance()

	artDirectory := ""

	if IsDev() {
		romDirectory := strings.ReplaceAll(romDirectory.Path, common.RomDirectory, GetRomDirectory())
		artDirectory = filepath.Join(romDirectory, ".media")
	} else {
		artDirectory = filepath.Join(romDirectory.Path, ".media")
	}

	tag := strings.ReplaceAll(romDirectory.Tag, "(", "")
	tag = strings.ReplaceAll(tag, ")", "")

	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(tag)

	artList, err := client.ListDirectory(section.HostSubdirectory)

	if err != nil {
		logger.Info("Unable to fetch artlist", zap.Error(err))
		return ""
	}

	noExtension := strings.TrimSuffix(game.Filename, filepath.Ext(game.Filename))

	var matched shared.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(noExtension)) {
			matched = art
			break
		}
	}

	if matched.Filename != "" {
		lastSavedArtPath, err := client.DownloadArt(section.HostSubdirectory, artDirectory, matched.Filename, game.Filename)
		if err != nil {
			return ""
		}

		src, err := imaging.Open(lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to open last saved art", zap.Error(err))
			return ""
		}

		dst := imaging.Resize(src, 500, 0, imaging.Lanczos)

		err = imaging.Save(dst, lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to save resized last saved art", zap.Error(err))
			return ""
		}

		return lastSavedArtPath
	}

	return ""
}

func RenameRom(game shared.Item, newFilename string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	oldPath := filepath.Join(romDirectory.Path, game.Filename)
	oldExt := filepath.Ext(game.Filename)
	newPath := filepath.Join(romDirectory.Path, newFilename+oldExt)

	logger.Debug("Renaming Rom", zap.String("oldPath", oldPath), zap.String("newPath", newPath))

	err := MoveFile(oldPath, newPath)
	if err != nil {
		logger.Error("failed to move file", zap.Error(err))
		return "", err
	}

	gameTrackerOldPath := strings.ReplaceAll(oldPath, common.RomDirectory+"/", "")
	gameTrackerNewPath := strings.ReplaceAll(newPath, common.RomDirectory+"/", "")

	logger.Debug("Updating Game Tracker for Rename",
		zap.String("old_path", oldPath), zap.String("new_path", newPath))

	MigrateGameTrackerData(newFilename, gameTrackerOldPath, gameTrackerNewPath)

	RenameSaveFile(strings.ReplaceAll(game.Filename, filepath.Ext(game.Filename), ""), newFilename, romDirectory)

	existingArtFilename, err := FindExistingArt(game.Filename, romDirectory)
	if err != nil {
		logger.Error("failed to find existing art", zap.Error(err))
		return "", err
	} else if existingArtFilename != "" {
		oldArtPath := filepath.Join(romDirectory.Path, ".media", existingArtFilename)
		oldArtExt := filepath.Ext(existingArtFilename)
		newArtPath := filepath.Join(romDirectory.Path, ".media", newFilename+oldArtExt)

		if _, err := os.Stat(oldArtPath); os.IsNotExist(err) {
			logger.Info("No media exists. Skipping...")
			return "", err
		} else {
			err := MoveFile(oldArtPath, newArtPath)
			if err != nil {
				logger.Error("failed to rename existing art", zap.Error(err))
				return "", err
			}
		}
	}
	return newFilename + oldExt, nil
}

func RenameSaveFile(oldFilename string, newFilename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	unwrappedTag := strings.ReplaceAll(romDirectory.Tag, "(", "")
	unwrappedTag = strings.ReplaceAll(unwrappedTag, ")", "")

	saveFileDirectoryWithTag := filepath.Join(saveFileDirectory, unwrappedTag)

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(saveFileDirectoryWithTag, true)
	if err != nil {
		logger.Error("failed to change directory", zap.Error(err))
		return
	}

	var foundSaveFile shared.Item

	for _, item := range fb.Items {
		if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(oldFilename)) {
			foundSaveFile = item
		}
	}

	if foundSaveFile.Filename == "" {
		logger.Info("No save file found. Skipping...")
		return
	}

	oldExt := strings.ReplaceAll(foundSaveFile.Filename, oldFilename, "")
	newPath := filepath.Join(saveFileDirectory, unwrappedTag, newFilename+oldExt)

	err = MoveFile(foundSaveFile.Path, newPath)
	if err != nil {
		logger.Error("failed to rename save file", zap.Error(err))
		return
	}
}

func DeleteArt(filename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	art, err := FindExistingArt(filename, romDirectory)
	if err != nil {
		logger.Error("failed to find existing art", zap.Error(err))
		return
	} else if art == "" {
		logger.Info("No art. Skipping delete.")
		return
	}

	artPath := filepath.Join(romDirectory.Path, ".media", art)
	common.DeleteFile(artPath)
}

func HasGameTrackerData(romFilename string, romDirectory shared.RomDirectory) bool {
	logger := common.GetLoggerInstance()

	db, err := sql.Open("sqlite3", GetGameTrackerDBPath())
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return false
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close game tracker database", zap.Error(err))
		}
	}(db)

	romPath := filepath.Join(strings.ReplaceAll(romDirectory.Path, GetRomDirectory()+"/", ""), romFilename)

	var romID string
	err = db.QueryRow("SELECT id FROM rom WHERE file_path = ?", romPath).Scan(&romID)
	if err != nil {
		logger.Error("Failed to find ROM ID", zap.Error(err))
		return false
	}

	return romID != ""
}

func MigrateGameTrackerData(filename string, oldPath string, newPath string) bool {
	logger := common.GetLoggerInstance()

	db, err := sql.Open("sqlite3", gameTrackerDBPath)
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return false
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close game tracker database", zap.Error(err))
		}
	}(db)

	logger.Debug("Migrating game tracker data", zap.String("filename", filename),
		zap.String("oldPath", oldPath), zap.String("newPath", newPath))

	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}

	var romID string
	err = tx.QueryRow("SELECT id FROM rom WHERE file_path = ?", oldPath).Scan(&romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to find ROM ID", zap.Error(err))
		return false
	}

	if romID == "" {
		logger.Warn("No ROM ID found", zap.String("old_path", oldPath))
		return false
	}

	_, err = tx.Exec("UPDATE rom SET name = ?, file_path = ? WHERE id = ?", filename, newPath, romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to update game tracker Rom name", zap.Error(err))
		return false
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker Rom Name updated successfully")
	return true
}

func ClearGameTracker(romName string, romDirectory shared.RomDirectory) bool {
	logger := common.GetLoggerInstance()

	db, err := sql.Open("sqlite3", GetGameTrackerDBPath())
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return false
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close game tracker database", zap.Error(err))
		}
	}(db)

	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}

	romPath := filepath.Join(strings.ReplaceAll(romDirectory.Path, GetRomDirectory()+"/", ""), romName)

	var romID string
	err = tx.QueryRow("SELECT id FROM rom WHERE file_path = ?", romPath).Scan(&romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to find ROM ID", zap.Error(err))
		return false
	}

	if romID == "" {
		logger.Warn("No ROM ID found", zap.String("fullpath", romPath), zap.String("name", romName))
		return false
	}

	_, err = tx.Exec("DELETE FROM play_activity WHERE rom_id = ?", romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to delete play activity", zap.Error(err))
		return false
	}

	_, err = tx.Exec("DELETE FROM rom WHERE id = ?", romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to delete rom", zap.Error(err))
		return false
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker data cleared successfully")
	return true
}

func ClearSaveStates() {
	// TODO - implement
}

func ArchiveRom(selectedGame shared.Item, romDirectory shared.RomDirectory) error {
	archiveRoot := GetArchiveRoot()

	logger := common.GetLoggerInstance()

	logger.Debug("Archive Start", zap.String("selected_file", selectedGame.Filename), zap.Any("with_ext", selectedGame))

	oldPath := filepath.Join(romDirectory.Path, selectedGame.Filename)
	oldPathSubdirectory := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")
	newPath := filepath.Join(archiveRoot, oldPathSubdirectory, selectedGame.Filename)

	logger.Debug("Archiving Rom", zap.String("oldPath", oldPath), zap.String("newPath", newPath))

	err := MoveFile(oldPath, newPath)
	if err == nil {
		existingArtFilename, err := FindExistingArt(selectedGame.Filename, romDirectory)
		if err != nil {
			logger.Error("failed to find existing art", zap.Error(err))
		} else {
			oldArtPath := filepath.Join(romDirectory.Path, ".media", existingArtFilename)
			newArtPath := filepath.Join(archiveRoot, oldPathSubdirectory, ".media", existingArtFilename)

			err := MoveFile(oldArtPath, newArtPath)
			if err != nil {
				logger.Error("failed to archive existing art", zap.Error(err))
			}
		}
	}

	return err
}

func DeleteRom(game shared.Item, romDirectory shared.RomDirectory) {
	romPath := filepath.Join(romDirectory.Path, game.Filename)
	res := common.DeleteFile(romPath)

	if res {
		DeleteArt(game.Filename, romDirectory)
	}
}

func Nuke(game shared.Item, romDirectory shared.RomDirectory) {
	ClearGameTracker(game.Filename, romDirectory)
	DeleteRom(game, romDirectory)
}

func MoveFile(oldPath, newPath string) error {
	logger := common.GetLoggerInstance()

	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logger.Error("Failed to create destination directory", zap.Error(err))
		return err
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		logger.Error("Failed to move file", zap.Error(err))
		return err
	}

	return nil
}

func MakeDirectoryIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}
	return nil
}

func RenameCollection(collection models.Collection, name string) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	newFileName := name + ".txt"

	filepath.Dir(collection.CollectionFile)
	newPath := filepath.Join(filepath.Dir(collection.CollectionFile), newFileName)

	err := os.Rename(collection.CollectionFile, newPath)
	if err != nil {
		logger.Error("Failed to move file", zap.Error(err))
		return models.Collection{}, err
	}

	collection.DisplayName = name
	collection.CollectionFile = newPath

	return collection, nil
}

func DeleteCollection(collection models.Collection) {
	common.DeleteFile(collection.CollectionFile)
}

func AddCollectionGame(collection models.Collection, game shared.Item) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	if len(collection.Games) == 0 {
		if _, err := os.Stat(collection.CollectionFile); !os.IsNotExist(err) {
			logger.Debug("Collection file already exists. Loading...")

			loadCollection, err := ReadCollection(collection)

			if err != nil {
				return collection, err
			}
			collection = loadCollection
		}
	}

	for _, existingGame := range collection.Games {
		if existingGame.Path == game.Path {
			logger.Debug("Game already exists in collection", zap.String("path", game.Path))
			return collection, nil
		}
	}

	collection.Games = append(collection.Games, game)

	_ = SaveCollection(collection)

	return collection, nil
}

func ReadCollection(collection models.Collection) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	file, err := os.Open(collection.CollectionFile)
	if err != nil {
		logger.Error("failed to open collection", zap.Error(err))
		return collection, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var games []shared.Item

	for scanner.Scan() {
		line := scanner.Text()
		displayName := filepath.Base(line)
		displayName, _ = filebrowser.ItemNameCleaner(displayName, false)
		games = append(games, shared.Item{
			DisplayName: displayName,
			Path:        line,
		})
	}

	if err := scanner.Err(); err != nil {
		logger.Error("failed to read collection", zap.Error(err))
		return collection, err
	}

	collection.Games = games

	return collection, nil
}

func SaveCollection(collection models.Collection) error {
	dir := filepath.Dir(collection.CollectionFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(collection.CollectionFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range collection.Games {

		path := strings.ReplaceAll(line.Path, common.SDCardRoot, "")

		if _, err := writer.WriteString(path + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

func SaveConfig(config *models.Config) error {
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		logger := common.GetLoggerInstance()
		logger.Info("Config file does not exist, creating a blank one")

		file, err := os.Create("config.yml")
		if err != nil {
			return fmt.Errorf("error creating config file: %w", err)
		}
		file.Close()
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	viper.Set("art_download_type", config.ArtDownloadType)
	viper.Set("hide_empty", config.HideEmpty)
	viper.Set("log_level", config.LogLevel)

	return viper.WriteConfigAs("config.yml")
}
