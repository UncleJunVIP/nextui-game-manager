package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
)

func GetArchiveFileListBasic() ([]string, error) {
	entries, err := GetFileList(GetRomDirectory())
	if err != nil {
		return nil, err
	}

	var archiveFolders []string
	for _, folder := range entries {
		folderName := folder.Name()
		if err := EnsureDirectoryExists(filepath.Join(GetRomDirectory(), folderName)); err == nil {
			if strings.HasPrefix(folderName, ".") {
				if folderName != ".media" {
					archiveFolders = append(archiveFolders, folderName)
				}
			}
		}
	}

	return archiveFolders, nil
}

func GetArchiveFileList() ([]string, error) {
	archiveFolders, err := GetArchiveFileListBasic()

	if err != nil {
		return nil, err
	}

	if archiveFolders == nil {
		archiveFolders = append(archiveFolders, ".Archive")
	}
	return archiveFolders, nil
}

func ArchiveRom(selectedGame shared.Item, romDirectory shared.RomDirectory, archiveName string) error {
	logger := common.GetLoggerInstance()

	sourcePath := filepath.Join(romDirectory.Path, selectedGame.Filename)
	destinationPath := buildArchivePath(selectedGame.Filename, romDirectory, archiveName)

	logger.Debug("Archiving ROM", zap.String("from", sourcePath), zap.String("to", destinationPath))

	if err := MoveFile(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("failed to archive ROM: %w", err)
	}

	archiveArtFile(selectedGame.Filename, romDirectory, archiveName, logger)
	return nil
}

func RestoreRom(selectedGame shared.Item, romDirectory shared.RomDirectory, archive shared.RomDirectory) error {
	logger := common.GetLoggerInstance()

	sourcePath := filepath.Join(romDirectory.Path, selectedGame.Filename)
	destinationPath := buildRestorePath(selectedGame.Filename, romDirectory, archive)

	logger.Debug("Restoring ROM", zap.String("from", sourcePath), zap.String("to", destinationPath))

	if err := MoveFile(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("failed to restore ROM: %w", err)
	}

	restoreArtFile(selectedGame.Filename, romDirectory, archive, logger)
	return nil
}

func CleanArchiveName(archive string) string {
	return strings.TrimPrefix(archive, ".")
}

func DeleteArchive(archive shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()
	res, err := deleteArchiveRecursive(archive.Path, 0)

	if err != nil {
		logger.Error("Failed to traverse archive", zap.Error(err))
		return res, err
	}

	if res != "" {
		return res, nil
	}

	removeErr := os.RemoveAll(archive.Path)

	if removeErr != nil {
		return "", removeErr
	}

	return "", nil
}

func deleteArchiveRecursive(currentDirectory string, currentDepth int) (string, error) {
	logger := common.GetLoggerInstance()
	if currentDepth > 10 {
		return "Max Depth Exceeded", nil
	}

	entries, err := GetFileList(currentDirectory)

	if err != nil {
		logger.Error("Failed to traverse archive", zap.Error(err))
		return "", err
	}

	for _, file := range entries {
		if !file.IsDir() {
			return file.Name(), nil
		}

		res, recurseErr := deleteArchiveRecursive(filepath.Join(currentDirectory, file.Name()), currentDepth+1)

		if recurseErr != nil {
			return "", recurseErr
		}

		if res != "" {
			return res, nil
		}
	}

	return "", nil
}

func PrepArchiveName(archive string) string {
	if !strings.HasPrefix(archive, ".") {
		return "." + archive
	}
	return archive
}

func buildArchivePath(filename string, romDirectory shared.RomDirectory, archiveName string) string {
	archiveRoot := GetArchiveRoot(archiveName)
	subdirectory := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")
	return filepath.Join(archiveRoot, subdirectory, filename)
}

func buildRestorePath(filename string, romDirectory shared.RomDirectory, archive shared.RomDirectory) string {
	subdirectory := strings.ReplaceAll(romDirectory.Path, archive.Path, "")
	return filepath.Join(GetRomDirectory(), subdirectory, filename)
}

func archiveArtFile(filename string, romDirectory shared.RomDirectory, archiveName string, logger *slog.Logger) {
	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil || artPath == "" {
		return
	}

	archiveRoot := GetArchiveRoot(archiveName)
	subdirectory := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")
	destinationPath := filepath.Join(archiveRoot, subdirectory, ".media", filepath.Base(artPath))

	if err := MoveFile(artPath, destinationPath); err != nil {
		logger.Error("Failed to archive art file", zap.Error(err))
	}
}

func restoreArtFile(filename string, romDirectory shared.RomDirectory, archive shared.RomDirectory, logger *slog.Logger) {
	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil || artPath == "" {
		return
	}

	subdirectory := strings.ReplaceAll(romDirectory.Path, archive.Path, "")
	destinationPath := filepath.Join(GetRomDirectory(), subdirectory, ".media", filepath.Base(artPath))

	if err := MoveFile(artPath, destinationPath); err != nil {
		logger.Error("Failed to restore art file", zap.Error(err))
	}
}
