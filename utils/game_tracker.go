package utils

import (
	"database/sql"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"path/filepath"
	"maps"
	"slices"
	"strings"
	"strconv"
	"time"
)

func HasGameTrackerData(romFilename string, romDirectory shared.RomDirectory) bool {
	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	gameTrackerRomPath := buildGameTrackerPath(romDirectory.Path, romFilename)

	var romID string
	err = db.QueryRow("SELECT id FROM rom WHERE file_path = ?", gameTrackerRomPath).Scan(&romID)
	return err == nil && romID != ""
}

func buildGameTrackerPath(romDirectoryPath, filename string) string {
	fullPath := filepath.Join(romDirectoryPath, filename)
	return strings.ReplaceAll(fullPath, common.RomDirectory+"/", "")
}

func MigrateGameTrackerData(filename, oldPath, newPath string) bool {
	logger := common.GetLoggerInstance()

	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	logger.Debug("Migrating game tracker data",
		zap.String("filename", filename),
		zap.String("oldPath", oldPath),
		zap.String("newPath", newPath))

	return executeGameTrackerMigration(db, filename, oldPath, newPath, logger)
}

func updateGameTrackerForRename(oldFilename, newFilename string, romDirectory shared.RomDirectory, logger *zap.Logger) {
	oldPath := buildGameTrackerPath(romDirectory.Path, oldFilename)
	newPath := buildGameTrackerPath(romDirectory.Path, newFilename+filepath.Ext(oldFilename))

	logger.Debug("Updating game tracker for rename", zap.String("old", oldPath), zap.String("new", newPath))
	MigrateGameTrackerData(newFilename, oldPath, newPath)
}

func findRomID(tx *sql.Tx, romPath string) (string, error) {
	var romID string
	err := tx.QueryRow("SELECT id FROM rom WHERE file_path = ?", romPath).Scan(&romID)
	return romID, err
}

func updateRomData(tx *sql.Tx, filename, newPath, romID string) error {
	_, err := tx.Exec("UPDATE rom SET name = ?, file_path = ? WHERE id = ?", filename, newPath, romID)
	return err
}

func executeGameTrackerMigration(db *sql.DB, filename, oldPath, newPath string, logger *zap.Logger) bool {
	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}
	defer tx.Rollback()

	romID, err := findRomID(tx, oldPath)
	if err != nil || romID == "" {
		logger.Error("Failed to find ROM ID", zap.String("path", oldPath), zap.Error(err))
		return false
	}

	if err := updateRomData(tx, filename, newPath, romID); err != nil {
		logger.Error("Failed to update ROM data", zap.Error(err))
		return false
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker ROM data updated successfully")
	return true
}

func ClearGameTracker(romName string, romDirectory shared.RomDirectory) bool {
	logger := common.GetLoggerInstance()

	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	romPath := buildGameTrackerPath(romDirectory.Path, romName)

	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}
	defer tx.Rollback()

	romID, err := findRomID(tx, romPath)
	if err != nil || romID == "" {
		logger.Warn("No ROM found to clear", zap.String("path", romPath))
		return false
	}

	if err := deleteGameTrackerData(tx, romID); err != nil {
		logger.Error("Failed to delete game tracker data", zap.Error(err))
		return false
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker data cleared successfully")
	return true
}

func deleteGameTrackerData(tx *sql.Tx, romID string) error {
	if _, err := tx.Exec("DELETE FROM play_activity WHERE rom_id = ?", romID); err != nil {
		return err
	}
	_, err := tx.Exec("DELETE FROM rom WHERE id = ?", romID)
	return err
}

func openGameTrackerDB() (*sql.DB, error) {
	logger := common.GetLoggerInstance()

	db, err := sql.Open("sqlite3", GetGameTrackerDBPath())
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return nil, err
	}
	return db, nil
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		logger := common.GetLoggerInstance()
		logger.Error("Failed to close database", zap.Error(err))
	}
}

func FindRomHomeFromAggregate(gameAggregate models.PlayHistoryAggregate, showArchives bool) string {
	gamePath := gameAggregate.Path
	if DoesFileExists(gamePath) {
		if showArchives {
			return "(+) "
		}
		return ""
	}

	archiveList, err := GetArchiveFileListBasic()
	if err == nil {
		for _, archiveName := range archiveList {
			gameSubPath := strings.ReplaceAll(gamePath, GetRomDirectory(), "")
			archivePath := filepath.Join(GetRomDirectory(), archiveName, gameSubPath)
			if DoesFileExists(archivePath) {
				if showArchives {
					return "(" + string(CleanArchiveName(archiveName)[0]) + ") "
				}
				return ""
			}
		}
	}

	return "(-) "
}

func CollectGameAggregateFromGame(gameItem shared.Item, gamePlayMap map[string][]models.PlayHistoryAggregate) (models.PlayHistoryAggregate, string) {
	console := extractItemConsoleName(gameItem)
	PlayHistoryList := gamePlayMap[console]

	for _, gameAggregate := range PlayHistoryList {
		if gameAggregate.Name == gameItem.DisplayName {
			return gameAggregate, console
		}
	}

	return models.PlayHistoryAggregate{
		Name: gameItem.DisplayName,
		PlayTimeTotal: 0,
		PlayCountTotal: 0,
	}, console
}

func convertIntListToStringList(intList []int) []string {
	var stringList []string
	for _, i := range intList {
		stringList = append(stringList, strconv.Itoa(i))
	}
	return stringList
}

const (
	NoFilter		= 0   // No filter
	YearMonth       = 1   // YearMonth Filter
)

func GenFiltersList(romIds []int, searchFilterString string, existingFilterType int) []models.PlayHistorySearchFilter {
	logger := common.GetLoggerInstance()
	db, err := openGameTrackerDB()
	if err != nil {
		return nil
	}
	defer closeDB(db)

	// If any roms are selected by parents, filter for those roms
	romWhereStr := ""
	if len(romIds) > 0 {
		romWhereStr = " rom_id in ('" + strings.Join(convertIntListToStringList(romIds), "','") + "') "
	}

	// rom filters or search filters are present, include WHERE clause
	whereStr := ""
	if romWhereStr != "" || searchFilterString != "" {
		whereStr = " WHERE "
	}

	// rom filters and search filters are present, include AND clause
	andStr := ""
	if romWhereStr != "" && searchFilterString != "" {
		andStr = " AND "
	}

	// Generate date string modifier constructs
	newFilterFormat := ""
	switch existingFilterType {
	case YearMonth:
		newFilterFormat = "STRFTIME('%Y.%m.%d', DATETIME(play_activity.created_at, 'unixepoch'))"
	default:
		newFilterFormat = "STRFTIME('%Y.%m', DATETIME(play_activity.created_at, 'unixepoch'))"
	}

	rows, err := db.Query("SELECT " + newFilterFormat + " as new_filter, " + 
						  "sum(play_activity.play_time) as play_time " +
        				  "FROM play_activity " +
						  whereStr + romWhereStr + andStr + searchFilterString +
        				  "GROUP BY " + newFilterFormat)
	defer rows.Close()

	var filterList []models.PlayHistorySearchFilter
	for rows.Next() {
		var newFilter 	string
		var playTime 	int
		if err := rows.Scan(&newFilter, &playTime); err != nil {
			logger.Error("Failed to load game tracker data", zap.Error(err))
		}

		singleFilter := models.PlayHistorySearchFilter{
			DisplayName:	newFilter,
			SqlFilter: 		newFilterFormat + " = '" + newFilter + "'",
			FilterType:		existingFilterType + 1,
			PlayTime:		playTime,
		}

		filterList = append(filterList, singleFilter)
	}

	return filterList
}

func GenerateSingleGameGranularRecords(romIds []int, searchFilter string) []models.PlayHistoryGranular {
	if len(romIds) == 0 {
		return nil
	}
	
	logger := common.GetLoggerInstance()
	db, err := openGameTrackerDB()
	if err != nil {
		return nil
	}
	defer closeDB(db)

	romIdString := strings.Join(convertIntListToStringList(romIds), "','")

	whereMod := ""
	if searchFilter != "" {
		whereMod = " and " + searchFilter + " "
	}

	rows, err := db.Query("SELECT play_time, created_at, updated_at " +
        				  "FROM play_activity " +
						  "WHERE rom_id in ('"+romIdString+"') " +
						  whereMod +
        				  "ORDER BY created_at")
	defer rows.Close()

	var granularList []models.PlayHistoryGranular
	for rows.Next() {
		var playTime 	int
		var createTime 	int
		var updateTime 	int
		if err := rows.Scan(&playTime, &createTime, &updateTime); err != nil {
			logger.Error("Failed to load game tracker data", zap.Error(err))
		}

		playTrack := models.PlayHistoryGranular{
			PlayTime:	playTime,
			StartTime: 	createTime,
			UpdateTime:	updateTime,
		}
		granularList = append(granularList, playTrack)
	}

	return granularList
}

func GenerateCurrentGameStats(searchFilter string) (map[string][]models.PlayHistoryAggregate, map[string]int, int) {
	logger := common.GetLoggerInstance()
	db, err := openGameTrackerDB()
	if err != nil {
		return nil, nil, 0
	}
	defer closeDB(db)

	whereMod := ""
	if searchFilter != "" {
		whereMod = " WHERE " + searchFilter + " "
	}

	rows, err := db.Query("SELECT rom.id, rom.name, rom.file_path, " +
						  "SUM(play_activity.play_time) AS play_time_total, " +
						  "COUNT(play_activity.ROWID) AS play_count_total, " +
						  "MIN(play_activity.created_at) AS first_played_at, " +
						  "MAX(play_activity.created_at) AS last_played_at " +
        				  "FROM rom " +
						  "LEFT JOIN play_activity " +
						  "ON rom.id = play_activity.rom_id " +
						  whereMod +
        				  "GROUP BY rom.id " +
        				  "HAVING play_time_total > 0 " +
        				  "ORDER BY play_time_total DESC")
	defer rows.Close()

	gamePlayMap := make(map[string][]models.PlayHistoryAggregate)
	consolePlayMap := make(map[string]int)
	totalPlay := 0
	multiMap := make(map[string]bool)
	for rows.Next() {
		var id 				int
		var name 			string
		var filePath 		string
		var playTimeTotal 	int
		var playCountTotal 	int
		var firstPlayedTime int
		var lastPlayedTime 	int
		if err := rows.Scan(&id, &name, &filePath, &playTimeTotal, &playCountTotal, &firstPlayedTime, &lastPlayedTime); err != nil {
			logger.Error("Failed to load game tracker data", zap.Error(err))
		}

		romName, romPath, multi := extractMultiDiscName(name, filePath)
		playTrack := models.PlayHistoryAggregate{
			Id:					[]int{id},
			Name: 				romName,
			Path:				romPath,
			PlayTimeTotal:    	playTimeTotal,
			PlayCountTotal:    	playCountTotal,
			FirstPlayedTime: 	time.Unix(int64(firstPlayedTime), 0),
			LastPlayedTime:    	time.Unix(int64(lastPlayedTime), 0),
		}
		console := extractPlayConsoleName(filePath)

		if multi {
			multiMap[console] = true
			gamePlayMap[console] = appendMultiDiscAggregate(gamePlayMap[console], playTrack)
		} else {
			gamePlayMap[console] = append(gamePlayMap[console], playTrack)
		}

		consolePlayMap[console] = consolePlayMap[console] + playTrack.PlayTimeTotal

		totalPlay = totalPlay + playTrack.PlayTimeTotal
	}

	gamePlayMap = sortPlayMap(gamePlayMap, multiMap)

	return gamePlayMap, consolePlayMap, totalPlay
}

func sortPlayMap(playMap map[string][]models.PlayHistoryAggregate, multiMap map[string]bool) map[string][]models.PlayHistoryAggregate {
	keys := slices.Sorted(maps.Keys(multiMap))
	for _, key := range keys {
		aggregateList := playMap[key]
		slices.SortFunc(aggregateList, func(a, b models.PlayHistoryAggregate) int {
			return  b.PlayTimeTotal - a.PlayTimeTotal
		})
		playMap[key] = aggregateList
	}
	return playMap
}

func appendMultiDiscAggregate(existingList []models.PlayHistoryAggregate, newAggregate models.PlayHistoryAggregate) []models.PlayHistoryAggregate {
	for index, existingAggregate := range existingList {
		if existingAggregate.Name == newAggregate.Name {
			existingList[index] = models.PlayHistoryAggregate{
				Id:					appendUniqueAggregateId(existingAggregate.Id, newAggregate.Id[0]),
				Name: 				existingAggregate.Name,
				PlayTimeTotal:    	existingAggregate.PlayTimeTotal+newAggregate.PlayTimeTotal,
				PlayCountTotal:    	existingAggregate.PlayCountTotal+newAggregate.PlayCountTotal,
				FirstPlayedTime: 	minTime(existingAggregate.FirstPlayedTime, newAggregate.FirstPlayedTime),
				LastPlayedTime:    	maxTime(existingAggregate.FirstPlayedTime, newAggregate.FirstPlayedTime),
			}
			return existingList
		}
	}
	return append(existingList, newAggregate)
}

func appendUniqueAggregateId(existingIds []int, newId int) []int {
	for _, id := range existingIds {
		if id == newId {
			return existingIds
		}
	}
	return append(existingIds, newId)
}

func minTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func extractMultiDiscName(romName string, filePath string) (string, string, bool) {
	if (strings.Contains(romName, "(Disc") || strings.Contains(romName, "(Disk")) {
		pathList := strings.Split(filePath, "/")
		if len(pathList) >= 2 {
			return pathList[len(pathList)-2], filepath.Join(GetRomDirectory(), filePath, ".."), true
		}
	}
	return romName, filepath.Join(GetRomDirectory(), filePath), false
}

func extractPlayConsoleName(romFilePath string) string {
	return strings.Split(romFilePath, "/")[0]
}

func extractItemConsoleName(gameItem shared.Item) string {
	pathSplit := strings.Split(gameItem.Path, "/")
	if len(pathSplit) < 5 {
		return ""
	}
	return pathSplit[4]
}

func ConvertSecondsToHumanReadable(gameTimeSeconds int) string {
	hours := gameTimeSeconds/3600
	minutes := (gameTimeSeconds/60)%60
	seconds := gameTimeSeconds%60
	return fmt.Sprintf("%d Hours, %d Minutes, %d Seconds", hours, minutes, seconds)
}

func ConvertSecondsToHumanReadableAbbreviated(gameTimeSeconds int) string {
	hours := gameTimeSeconds/3600
	minutes := (gameTimeSeconds/60)%60
	seconds := gameTimeSeconds%60
	return fmt.Sprintf("%dH %dM %dS", hours, minutes, seconds)
}

func FilterPlayList(itemList []models.PlayHistoryAggregate, keywords ...string) []models.PlayHistoryAggregate {
	if len(keywords) == 0 {
		return itemList
	}

	var filteredItems []models.PlayHistoryAggregate
	for _, item := range itemList {
		if matchesAnyKeyword(item.Name, keywords) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}
