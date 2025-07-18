package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type Config struct {
	ArtDownloadType 			sum.Int[shared.ArtDownloadType] `yaml:"art_download_type"`
	FuzzySearchThreshold float64                         `yaml:"fuzzy_search_threshold"`
	HideEmpty       			bool                            `yaml:"hide_empty"`
	ShowArt         			bool                            `yaml:"show_art"`
	LogLevel        			string                   		`yaml:"log_level"`
	PlayHistoryShowCollections	bool                            `yaml:"play_history_show_collections"`
	PlayHistoryShowArchives     bool                          	`yaml:"play_history_show_archives"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("log_level", c.LogLevel)

	return nil
}
