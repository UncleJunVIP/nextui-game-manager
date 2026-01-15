package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/shared"
)

func GetFileList(dirPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}
	return entries, nil
}

func DoesFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, defaultDirPerm)
	}
	return nil
}

func MoveFile(sourcePath, destinationPath string) error {
	logger := common.GetLoggerInstance()

	if err := EnsureDirectoryExists(filepath.Dir(destinationPath)); err != nil {
		logger.Error("Failed to create destination directory", zap.Error(err))
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		logger.Error("Failed to move file", zap.String("from", sourcePath), zap.String("to", destinationPath), zap.Error(err))
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, destinationPath, err)
	}

	return nil
}

func DeleteRom(game shared.Item, romDirectory shared.RomDirectory) {
	romPath := filepath.Join(romDirectory.Path, game.Filename)
	if common.DeleteFile(romPath) {
		DeleteArt(game.Filename, romDirectory)
	}
}

func Nuke(game shared.Item, romDirectory shared.RomDirectory) {
	ClearGameTracker(game.Filename, romDirectory)
	DeleteRom(game, romDirectory)
}
