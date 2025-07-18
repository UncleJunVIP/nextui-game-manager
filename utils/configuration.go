package utils

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
)

func SaveConfig(config *models.Config) error {
	configFile := "config.yml"

	if !DoesFileExists(configFile) {
		if err := CreateEmptyConfigFile(configFile); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to setup viper: %w", err)
	}

	viper.Set("art_download_type", config.ArtDownloadType)
	viper.Set("fuzzy_search_threshold", config.FuzzySearchThreshold)
	viper.Set("hide_empty", config.HideEmpty)
	viper.Set("show_art", config.ShowArt)
	viper.Set("log_level", config.LogLevel)
	viper.Set("play_history_show_collections", config.PlayHistoryShowCollections)
	viper.Set("play_history_show_archives", config.PlayHistoryShowArchives)


	return viper.WriteConfigAs(configFile)
}

func CreateEmptyConfigFile(filename string) error {
	logger := common.GetLoggerInstance()
	logger.Info("Creating new config file", zap.String("file", filename))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return file.Close()
}
