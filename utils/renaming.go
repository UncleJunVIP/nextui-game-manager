package utils

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"path/filepath"
	"strings"
)

func renameSaveFile(oldFilename, newFilename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	tag := cleanTag(romDirectory.Tag)

	saveDir := filepath.Join(GetSaveFileDirectory(), tag)
	fb := filebrowser.NewFileBrowser(logger)

	if err := fb.CWD(saveDir, true); err != nil {
		logger.Error("Failed to access save directory", zap.String("dir", saveDir), zap.Error(err))
		return
	}

	saveFile := findSaveFile(fb.Items, oldFilename)
	if saveFile.Filename == "" {
		logger.Info("No save file found to rename")
		return
	}

	ext := strings.ReplaceAll(saveFile.Filename, removeFileExtension(oldFilename), "")
	newSavePath := filepath.Join(saveDir, newFilename+ext)

	if err := MoveFile(saveFile.Path, newSavePath); err != nil {
		logger.Error("Failed to rename save file", zap.Error(err))
	}
}

func RenameCollection(collection models.Collection, name string) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	newPath := filepath.Join(filepath.Dir(collection.CollectionFile), name+".txt")

	if err := os.Rename(collection.CollectionFile, newPath); err != nil {
		logger.Error("Failed to rename collection file", zap.Error(err))
		return models.Collection{}, fmt.Errorf("failed to rename collection: %w", err)
	}

	collection.DisplayName = name
	collection.CollectionFile = newPath
	return collection, nil
}

func renameCollectionEntries(game shared.Item, oldDisplayName string, romDirectory shared.RomDirectory, collectionMap map[string][]models.Collection) {
	collections := findCollectionsContainingGame(collectionMap, game)
	for _, collection := range collections {
		updateCollectionGamePath(collection, oldDisplayName, game, romDirectory)
		SaveCollection(collection)
	}
}

func RenameRom(game shared.Item, newFilename string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	oldPath := filepath.Join(romDirectory.Path, game.Filename)
	newPath := buildNewRomPath(romDirectory.Path, newFilename, game.Filename)

	logger.Debug("Renaming ROM", zap.String("from", oldPath), zap.String("to", newPath))

	if err := MoveFile(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename ROM file: %w", err)
	}

	renameAssociatedFile(game.Filename, newFilename, newPath, ".cue")
	renameAssociatedFile(game.Filename, newFilename, newPath, ".m3u")

	updateGameTrackerForRename(game.Filename, newFilename, romDirectory, logger)
	renameSaveFile(game.Filename, newFilename, romDirectory)
	// renameCollectionEntries(game, game.Filename, romDirectory) TODO need to finish this functionality
	renameArtFile(game.Filename, newFilename, romDirectory, logger)

	return filepath.Base(newPath), nil
}

func buildNewRomPath(romDirectoryPath, newFilename, oldFilename string) string {
	ext := filepath.Ext(oldFilename)
	return filepath.Join(romDirectoryPath, newFilename+ext)
}

func renameAssociatedFile(oldFilename string, newFilename string, newPath string, extension string) {
	logger := common.GetLoggerInstance()

	oldAssociatedFilename := removeFileExtension(oldFilename) + extension
	oldAssociatedPath := filepath.Join(newPath, oldAssociatedFilename)

	if !DoesFileExists(oldAssociatedPath) {
		return
	}

	newAssociatedFilename := newFilename + extension
	newAssociatedPath := filepath.Join(newPath, newAssociatedFilename)

	if err := MoveFile(oldAssociatedPath, newAssociatedPath); err != nil {
		logger.Error("Failed to rename associated file",
			zap.String("from", oldAssociatedPath),
			zap.String("to", newAssociatedPath),
			zap.String("extension", extension),
			zap.Error(err))
	}
}
