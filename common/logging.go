package common

import (
	"log/slog"
	"os"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"go.uber.org/zap"
)

func LogStandardFatal(msg string, err error) {
	logger := GetLoggerInstance()
	logger.Error(msg, zap.Error(err))
	os.Exit(1)
}

func GetLoggerInstance() *slog.Logger {
	return gabagool.GetLogger()
}
