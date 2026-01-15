package common

import (
	"encoding/json"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func IsConnectedToInternet() bool {
	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	return err == nil
}

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}
	return RomDirectory
}

func ItemNameCleaner(filename string, stripTag bool) (string, string) {
	cleaned := filepath.Clean(filename)

	tags := TagRegex.FindAllStringSubmatch(cleaned, -1)

	var foundTags []string
	foundTag := ""

	if len(tags) > 0 {
		for _, tagPair := range tags {
			foundTags = append(foundTags, tagPair[0])
		}

		foundTag = strings.Join(foundTags, " ")
	}

	if stripTag {
		for _, tag := range foundTags {
			cleaned = strings.ReplaceAll(cleaned, tag, "")
		}
	}

	orderedFolderRegex := OrderedFolderRegex.FindStringSubmatch(cleaned)

	if len(orderedFolderRegex) > 0 {
		cleaned = strings.ReplaceAll(cleaned, orderedFolderRegex[0], "")
	}

	cleaned = strings.ReplaceAll(cleaned, path.Ext(cleaned), "")

	cleaned = strings.TrimSpace(cleaned)

	foundTag = strings.ReplaceAll(foundTag, "(", "")
	foundTag = strings.ReplaceAll(foundTag, ")", "")

	return cleaned, foundTag
}

func LoadSystemMapping() map[string]string {
	cwd, _ := os.Getwd()

	jsonPath := filepath.Join(cwd, "data", "systems-mapping.json")
	file, err := os.Open(jsonPath)

	var systemMapping map[string]string

	if err == nil {
		defer file.Close()
		_ = json.NewDecoder(file).Decode(&systemMapping)

		return systemMapping
	}

	return nil
}
