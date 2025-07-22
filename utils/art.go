package utils

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	"go.uber.org/zap"
	"math"
	"net/url"
	"path/filepath"
	"qlova.tech/sum"
	"regexp"
	"slices"
	"strings"
)

func FindExistingArt(selectedFile string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	mediaDir := filepath.Join(romDirectory.Path, ".media")
	if err := EnsureDirectoryExists(mediaDir); err != nil {
		logger.Info("No media directory found", zap.String("directory", romDirectory.Path))
		return "", nil
	}

	artList, err := GetFileList(mediaDir)
	if err != nil {
		return "", fmt.Errorf("failed to list art files: %w", err)
	}

	targetName := removeFileExtension(selectedFile)
	for _, art := range artList {
		if removeFileExtension(art.Name()) == targetName {
			return filepath.Join(mediaDir, art.Name()), nil
		}
	}

	return "", nil
}

func FindAllArt(romDirectory shared.RomDirectory, games shared.Items, downloadType sum.Int[shared.ArtDownloadType], fuzzySearchThreshold float64) []gaba.Download {
	logger := common.GetLoggerInstance()

	artMap := make(map[shared.Item]string)

	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(cleanTag(romDirectory.Tag))

	artList, err := client.ListDirectory(section.HostSubdirectory)
	if err != nil {
		logger.Info("Unable to fetch art list", zap.Error(err))
		return nil
	}

	for _, game := range games {
		matchedArt := findMatchingArt(artList, game.Filename, fuzzySearchThreshold)
		if matchedArt.Filename != "" {
			artMap[game] = matchedArt.Filename
		}
	}

	downloads := buildArtDownloads(artMap, client.RootURL, section)

	return downloads
}

func FindArt(romDirectory shared.RomDirectory, game shared.Item, downloadType sum.Int[shared.ArtDownloadType], fuzzySearchThreshold float64) string {
	logger := common.GetLoggerInstance()

	artDirectory := buildArtDirectory(game)
	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(cleanTag(romDirectory.Tag))

	artList, err := client.ListDirectory(section.HostSubdirectory)
	if err != nil {
		logger.Info("Unable to fetch art list", zap.Error(err))
		return ""
	}

	matchedArt := findMatchingArt(artList, game.Filename, fuzzySearchThreshold)
	if matchedArt.Filename == "" {
		return ""
	}

	lastSavedArtPath, err := client.DownloadArt(section.HostSubdirectory, artDirectory, matchedArt.Filename, game.Filename)
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

func FindRomsWithoutArt() (map[shared.RomDirectory][]shared.Item, error) {
	logger := common.GetLoggerInstance()
	romDirectories := make(map[shared.RomDirectory][]shared.Item)

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(GetRomDirectory(), false)
	if err != nil {
		logger.Error("Failed to get rom directories", zap.Error(err))
		return nil, fmt.Errorf("failed to get rom directories: %w", err)
	}

	for _, dir := range fb.Items {
		romDir := CreateRomDirectoryFromItem(dir)

		if romDir.Tag == "(PORTS)" {
			continue
		}

		romsWithoutArt, err := findRomsWithoutArtInDirectory(romDir)
		if err != nil {
			logger.Error("Failed to process rom directory", zap.String("directory", romDir.Path), zap.Error(err))
			continue
		}

		if len(romsWithoutArt) > 0 {
			romDirectories[romDir] = romsWithoutArt
		}
	}

	return romDirectories, nil
}

func findRomsWithoutArtInDirectory(romDir shared.RomDirectory) ([]shared.Item, error) {
	romFiles, err := getRomFilesRecursive(romDir.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get ROM files: %w", err)
	}

	var romsWithoutArt []shared.Item
	for _, romFile := range romFiles {
		romNameWithoutExt := removeFileExtension(romFile.Filename)

		artFilename := filepath.Join(filepath.Dir(romFile.Path), ".media", romNameWithoutExt+".png")

		if !DoesFileExists(artFilename) {
			romsWithoutArt = append(romsWithoutArt, romFile)
		}
	}

	return romsWithoutArt, nil
}

func findMatchingArt(artList []shared.Item, filename string, fuzzySearchThreshold float64) shared.Item {
	// toastd's trick for Libretro Thumbnail Naming
	cleanedName := strings.ReplaceAll(filename, "&", "_")

	targetName := removeFileExtension(cleanedName)

	slices.SortFunc(artList, func(a, b shared.Item) int {
		return strings.Compare(strings.ToLower(a.Filename), strings.ToLower(b.Filename))
	})

	// naive search first
	if idx, found := slices.BinarySearchFunc(artList, targetName, func(item shared.Item, tg string) int {
		stripExt := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		return strings.Compare(strings.ToLower(stripExt), strings.ToLower(tg))
	}); found {
		return artList[idx]
	} else if fuzzyMatch, meetsThreshold := fuzzyArtSearch(targetName, artList, fuzzySearchThreshold); meetsThreshold {
		return shared.Item{
			Filename: fuzzyMatch,
		}
	}

	return shared.Item{}
}

func fuzzyArtSearch(romFilename string, artList []shared.Item, threshold float64) (string, bool) {
	if threshold > .85 || threshold < .5 {
		threshold = .8 // Default
	}

	// Calculate similarity between two strings using Jaccard similarity
	similarity := func(s1, s2 string) float64 {
		// Tokenize
		tokens1 := make(map[string]bool)
		tokens2 := make(map[string]bool)

		for _, token := range strings.Fields(s1) {
			tokens1[token] = true
		}
		for _, token := range strings.Fields(s2) {
			tokens2[token] = true
		}

		// Calculate intersection and union
		intersection := 0
		for token := range tokens1 {
			if tokens2[token] {
				intersection++
			}
		}

		union := len(tokens1) + len(tokens2) - intersection
		if union == 0 {
			return 0
		}

		// Also consider character-level similarity for the main title
		title1 := strings.Split(s1, "(")[0]
		title2 := strings.Split(s2, "(")[0]

		// Simple character overlap ratio
		charSim := 0.0
		if len(title1) > 0 && len(title2) > 0 {
			matches := 0
			maxLen := len(title1)
			if len(title2) > maxLen {
				maxLen = len(title2)
			}

			// Count matching characters in order
			j := 0
			for i := 0; i < len(title1) && j < len(title2); i++ {
				if title1[i] == title2[j] {
					matches++
					j++
				} else {
					// Look ahead for match
					for k := j + 1; k < len(title2) && k-j < 3; k++ {
						if title1[i] == title2[k] {
							j = k + 1
							matches++
							break
						}
					}
				}
			}
			charSim = float64(matches) / float64(maxLen)
		}

		// Weighted combination of token and character similarity
		tokenSim := float64(intersection) / float64(union)
		return tokenSim*0.6 + charSim*0.4
	}

	bestMatch := ""
	bestScore := 0.0

	for _, art := range artList {
		pngNorm := removeFileExtension(art.Filename)
		score := similarity(romFilename, pngNorm)

		zipRegions := regexp.MustCompile(`\((.*?)\)`).FindAllStringSubmatch(romFilename, -1)
		pngRegions := regexp.MustCompile(`\((.*?)\)`).FindAllStringSubmatch(art.Filename, -1)

		regionMatch := false
		for _, zr := range zipRegions {
			for _, pr := range pngRegions {
				if strings.Contains(zr[1], "USA") && strings.Contains(pr[1], "USA") {
					regionMatch = true
					break
				}
				if strings.Contains(zr[1], "Europe") && strings.Contains(pr[1], "Europe") {
					regionMatch = true
					break
				}
			}
		}

		if regionMatch {
			score += 0.1
			if score > 1.0 {
				score = 1.0
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = art.Filename
		}
	}

	reachedThreshold := math.Round(bestScore*100) >= threshold*100

	return bestMatch, reachedThreshold
}

func buildArtDownloads(artMap map[shared.Item]string, rootUrl string, section shared.Section) []gaba.Download {
	var downloads []gaba.Download

	for game, artFilename := range artMap {
		remotePath := section.HostSubdirectory

		localPath := filepath.Join(buildArtDirectory(game), removeFileExtension(game.Filename)+".png")

		sourceURL, err := url.JoinPath(rootUrl, remotePath, artFilename)
		if err != nil {
			continue
		}

		downloads = append(downloads, gaba.Download{
			URL:         sourceURL,
			Location:    localPath,
			DisplayName: game.DisplayName,
		})
	}

	return downloads
}

func buildArtDirectory(game shared.Item) string {
	romDirectoryPath := filepath.Dir(game.Path)

	if IsDev() {
		adjustedPath := strings.ReplaceAll(romDirectoryPath, common.RomDirectory, GetRomDirectory())
		return filepath.Join(adjustedPath, ".media")
	}
	return filepath.Join(romDirectoryPath, ".media")
}

func renameArtFile(oldFilename, newFilename string, romDirectory shared.RomDirectory, logger *zap.Logger) {
	existingArtPath, err := FindExistingArt(oldFilename, romDirectory)
	if err != nil {
		logger.Error("Failed to find existing art", zap.Error(err))
		return
	}

	if existingArtPath == "" {
		return
	}

	if !DoesFileExists(existingArtPath) {
		logger.Info("Art file does not exist, skipping rename")
		return
	}

	ext := filepath.Ext(existingArtPath)
	newArtPath := filepath.Join(filepath.Dir(existingArtPath), newFilename+ext)

	if err := MoveFile(existingArtPath, newArtPath); err != nil {
		logger.Error("Failed to rename art file", zap.Error(err))
	}
}

func DeleteArt(filename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil {
		logger.Error("Failed to find existing art", zap.Error(err))
		return
	}

	if artPath == "" {
		logger.Info("No art found to delete")
		return
	}

	common.DeleteFile(artPath)
}
