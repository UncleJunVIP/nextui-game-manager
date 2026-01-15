package common

import (
	"os"
)

func DeleteFile(path string) bool {
	logger := GetLoggerInstance()

	err := os.Remove(path)
	if err != nil {
		logger.Error("Issue removing file",
			"path", path,
			"error", err)
		return false
	} else {
		logger.Debug("Removed file", "path", path)
		return true
	}
}
