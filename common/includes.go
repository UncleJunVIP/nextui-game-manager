package common

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed resources/data/systems-mapping.json
var systemMapping []byte

func InitIncludes() {
	logger := GetLoggerInstance()
	cwd, _ := os.Getwd()

	dataPath := filepath.Join(cwd, "data")

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		err := os.MkdirAll(dataPath, 0755)
		if err != nil {
			logger.Error("Failed to create data directory", "error", err)
			os.Exit(1)
		}
	}

	saveFile(systemMapping, filepath.Join(dataPath, "systems-mapping.json"))
}

func saveFile(data []byte, path string) {
	logger := GetLoggerInstance()

	file, err := os.Create(path)
	if err != nil {
		logger.Error("Failed to open / create / truncate", "file_path", path, "error", err)
		os.Exit(1)
	}

	_, err = file.Write(data)
	if err != nil {
		logger.Error("Failed to write", "file_path", path, "error", err)
		os.Exit(1)
	}

}
