package utils

import (
	"bufio"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func updateCollectionGamePath(collection models.Collection, oldDisplayName string, game shared.Item, romDirectory shared.RomDirectory) {
	romDirectoryStub := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")

	for i, item := range collection.Games {
		if item.DisplayName == oldDisplayName {
			collection.Games[i].Path = filepath.Join(romDirectoryStub, game.Filename)
		}
	}
}

func DeleteCollection(collection models.Collection) {
	common.DeleteFile(collection.CollectionFile)
}

func AddCollectionGames(collection models.Collection, games []shared.Item) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	if DoesFileExists(collection.CollectionFile) {
		logger.Debug("Loading existing collection")

		if loadedCollection, err := ReadCollection(collection); err == nil {
			collection = loadedCollection
		} else {
			return collection, fmt.Errorf("failed to load existing collection: %w", err)
		}
	}

	for _, game := range games {
		if GameExistsInCollection(collection.Games, game) {
			logger.Debug("Game already exists in collection", zap.String("path", game.Path))
			continue
		}
		collection.Games = append(collection.Games, game)
	}

	return collection, SaveCollection(collection)
}

func GameExistsInCollection(games []shared.Item, targetGame shared.Item) bool {
	for _, game := range games {
		if strings.Contains(strings.ToLower(game.Path), strings.ToLower(targetGame.DisplayName)) {
			return true
		}
	}
	return false
}

func ReadCollection(collection models.Collection) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	file, err := os.Open(collection.CollectionFile)
	if err != nil {
		logger.Error("Failed to open collection file", zap.String("file", collection.CollectionFile), zap.Error(err))
		return collection, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	var games []shared.Item
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		displayName := strings.ReplaceAll(filepath.Base(line), filepath.Ext(line), "")

		games = append(games, shared.Item{
			DisplayName: displayName,
			Path:        line,
		})
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Failed to read collection file", zap.Error(err))
		return collection, fmt.Errorf("failed to read collection: %w", err)
	}

	collection.Games = games
	return collection, nil
}

func SaveCollection(collection models.Collection) error {
	if err := EnsureDirectoryExists(filepath.Dir(collection.CollectionFile)); err != nil {
		return fmt.Errorf("failed to create collection directory: %w", err)
	}

	file, err := os.OpenFile(collection.CollectionFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePerm)
	if err != nil {
		return fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, game := range collection.Games {
		path := normalizeCollectionGamePath(game)
		if _, err := writer.WriteString(path + "\n"); err != nil {
			return fmt.Errorf("failed to write collection entry: %w", err)
		}
	}

	return nil
}

func normalizeCollectionGamePath(game shared.Item) string {
	path := strings.ReplaceAll(game.Path, GetRomDirectory()+"/", "/Roms/")

	if game.IsMultiDiscDirectory {
		path = filepath.Join(path, game.DisplayName+".m3u")
	}

	return path
}

func findCollectionsContainingGame(game shared.Item, logger *zap.Logger) []models.Collection {
	fb := filebrowser.NewFileBrowser(logger)
	if err := fb.CWD(GetCollectionDirectory(), false); err != nil {
		return nil
	}

	var collections []models.Collection
	for _, item := range fb.Items {
		collection := models.Collection{
			DisplayName:    item.DisplayName,
			CollectionFile: item.Path,
		}

		if loadedCollection, err := ReadCollection(collection); err == nil {
			if containsGame(loadedCollection.Games, game) {
				collections = append(collections, loadedCollection)
			}
		}
	}

	return collections
}

func containsGame(games []shared.Item, targetGame shared.Item) bool {
	return slices.ContainsFunc(games, func(game shared.Item) bool {
		return game.DisplayName == targetGame.DisplayName
	})
}

func GenerateCollectionMap() map[string][]models.Collection {
	collectionMap := make(map[string][]models.Collection)
	collectionList, _, _ := GenerateCollectionList("", false)
	for _, collection := range collectionList {
		for _, game := range collection.Games {
			collectionMap[game.DisplayName] = append(collectionMap[game.DisplayName], collection)
		}
	}
	return collectionMap
}

func GenerateCollectionList(searchFilter string, onScreen bool) (collections []models.Collection, exitCode int, e error) {
	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(GetCollectionDirectory(), false)
	if err != nil {
		if onScreen {
			ShowTimedMessage("Unable to Load Collections!", time.Second*2)
		}
		return nil, 404, nil
	}

	if fb.Items == nil || len(fb.Items) == 0 {
		return nil, 404, nil
	}

	itemList := fb.Items

	if searchFilter != "" {
		itemList = FilterList(itemList, searchFilter)
	}

	slices.SortFunc(itemList, func(a, b shared.Item) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	var collectionList []models.Collection
	for _, item := range itemList {
		col := models.Collection{DisplayName: item.DisplayName, CollectionFile: item.Path}
		col, err = ReadCollection(col)

		if err != nil {
			if onScreen {
				ShowTimedMessage("Unable to Load Collections!", time.Second*2)
			}
			return nil, -1, err
		}

		collectionList = append(collectionList, col)
	}

	return collectionList, 0, nil
}
