package utils

import (
	"fmt"
	romnibus "github.com/UncleJunVIP/ROMnibus/utils"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	gameTrackerDBPath  = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"
	saveFileDirectory  = "/mnt/SDCARD/Saves/"
	RecentlyPlayedFile = "/mnt/SDCARD/.userdata/shared/.minui/recent.txt"
	romnibusDBPath     = "/mnt/SDCARD/Tools/tg5040/Game Manager.pak/resources/ROMnibus.sqlite"
	defaultDirPerm     = 0755
	defaultFilePerm    = 0644
)

func init() {
	_, _ = romnibus.InitDB(GetROMnibusDB())
}

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}
	return common.RomDirectory
}

func GetArchiveRoot(archiveName string) string {
	if !strings.HasPrefix(archiveName, ".") {
		archiveName = "." + archiveName
	}

	return filepath.Join(GetRomDirectory(), archiveName)
}

func GetCollectionDirectory() string {
	dir := common.CollectionDirectory
	if IsDev() {
		dir = os.Getenv("COLLECTION_DIRECTORY")
	}

	_ = EnsureDirectoryExists(dir)
	return dir
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

func GetROMnibusDB() string {
	if IsDev() {
		return os.Getenv("ROMNIBUS_DB_PATH")
	}
	return romnibusDBPath
}

func CreateRomDirectoryFromItem(item shared.Item) shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: item.DisplayName,
		Tag:         item.Tag,
		Path:        item.Path,
	}
}

func FilterList(itemList []shared.Item, keywords ...string) []shared.Item {
	if len(keywords) == 0 {
		return itemList
	}

	var filteredItems []shared.Item
	for _, item := range itemList {
		if matchesAnyKeyword(item.Filename, keywords) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}

func matchesAnyKeyword(filename string, keywords []string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, keyword := range keywords {
		if strings.Contains(lowerFilename, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func InsertIntoSlice(slice []string, index int, values ...string) []string {
	if index < 0 {
		index = 0
	}
	if index > len(slice) {
		index = len(slice)
	}

	return append(slice[:index], append(values, slice[index:]...)...)
}

func getRomFilesRecursive(dirPath string) ([]shared.Item, error) {
	var romFiles []shared.Item

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWDDepth(dirPath, true, -1)

	if err != nil {
		return nil, fmt.Errorf("failed to get rom files: %w", err)
	}

	var selfContainedPaths []string

	for _, item := range fb.Items {
		if strings.Contains(item.Path, ".media") {
			continue
		}

		parent := filepath.Dir(item.Path)
		if slices.Contains(selfContainedPaths, parent) {
			continue
		}

		if item.IsSelfContainedDirectory {
			selfContainedPaths = append(selfContainedPaths, item.Path)
			romFiles = append(romFiles, item)
		}

		if !item.IsDirectory {
			romFiles = append(romFiles, item)
		}
	}

	return romFiles, err
}

func findSaveFile(items []shared.Item, targetFilename string) shared.Item {
	lowerTarget := strings.ToLower(targetFilename)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Filename), lowerTarget) {
			return item
		}
	}
	return shared.Item{}
}

func cleanTag(tag string) string {
	cleaned := strings.ReplaceAll(tag, "(", "")
	return strings.ReplaceAll(cleaned, ")", "")
}

func removeFileExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
