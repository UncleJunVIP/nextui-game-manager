package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type SettingsScreen struct {
}

func InitSettingsScreen() SettingsScreen {
	return SettingsScreen{}
}

func (s SettingsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Settings
}

func (s SettingsScreen) Draw() (settings interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	appState := state.GetAppState()

	items := []gabagool.ItemWithOptions{
		{
			Item: gabagool.MenuItem{
				Text: "Art Type",
			},
			Options: []gabagool.Option{
				{DisplayName: "Box Art", Value: "BOX_ART"},
				{DisplayName: "Title Screen", Value: "TITLE_SCREEN"},
				{DisplayName: "Logos", Value: "LOGOS"},
				{DisplayName: "Screenshots", Value: "SCREENSHOTS"},
			},
			SelectedOption: func() int {
				switch appState.Config.ArtDownloadType {
				case shared.ArtDownloadTypes.BOX_ART:
					return 0
				case shared.ArtDownloadTypes.TITLE_SCREEN:
					return 1
				case shared.ArtDownloadTypes.LOGOS:
					return 2
				case shared.ArtDownloadTypes.SCREENSHOTS:
					return 3
				default:
					return 0
				}
			}(),
		},
		{
			Item: gabagool.MenuItem{Text: "Art Fuzzy Search Threshold"},
			Options: []gabagool.Option{
				{DisplayName: "Lackadaisical (50%)", Value: .50},
				{DisplayName: "Loose (65%)", Value: .65},
				{DisplayName: "Eased (75%)", Value: .75},
				{DisplayName: "Default (80%)", Value: .80},
				{DisplayName: "Strict (85%)", Value: .85},
			},
			SelectedOption: func() int {
				switch appState.Config.FuzzySearchThreshold {
				case .50:
					return 0
				case .65:
					return 1
				case .75:
					return 2
				case .80:
					return 3
				case .85:
					return 4
				default:
					return 3
				}
			}(),
		},
		{
			Item: gabagool.MenuItem{Text: "Hide Empty Platforms"},
			Options: []gabagool.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
			SelectedOption: func() int {
				switch appState.Config.HideEmpty {
				case true:
					return 0
				case false:
					return 1
				default:
					return 1
				}
			}(),
		},
		{
			Item: gabagool.MenuItem{Text: "Show Art"},
			Options: []gabagool.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
			SelectedOption: func() int {
				switch appState.Config.ShowArt {
				case true:
					return 0
				case false:
					return 1
				default:
					return 1
				}
			}(),
		},
		{
			Item: gabagool.MenuItem{
				Text: "Log Level",
			},
			Options: []gabagool.Option{
				{DisplayName: "Debug", Value: "DEBUG"},
				{DisplayName: "Error", Value: "ERROR"},
			},
			SelectedOption: func() int {
				switch appState.Config.LogLevel {
				case "DEBUG":
					return 0
				case "ERROR":
					return 1
				}
				return 0
			}(),
		},
		{
			Item: gabagool.MenuItem{Text: "Play History Show Archive Tag"},
			Options: []gabagool.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
			SelectedOption: func() int {
				switch appState.Config.PlayHistoryShowArchives {
				case true:
					return 0
				case false:
					return 1
				default:
					return 1
				}
			}(),
		},
		{
			Item: gabagool.MenuItem{Text: "Play History Show Collection Tags"},
			Options: []gabagool.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
			SelectedOption: func() int {
				switch appState.Config.PlayHistoryShowCollections {
				case true:
					return 0
				case false:
					return 1
				default:
					return 1
				}
			}(),
		},
	}

	footerHelpItems := []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "←→", HelpText: "Cycle"},
		{ButtonName: "Start", HelpText: "Save"},
	}

	result, err := gabagool.OptionsList(
		"Game Manager Settings",
		items,
		footerHelpItems,
	)

	if err != nil {
		fmt.Println("Error showing options list:", err)
		return
	}

	if result.IsSome() {
		newSettingOptions := result.Unwrap().Items

		for _, option := range newSettingOptions {
			if option.Item.Text == "Art Type" {
				artTypeValue := option.Options[option.SelectedOption].Value.(string)
				switch artTypeValue {
				case "BOX_ART":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.BOX_ART
				case "TITLE_SCREEN":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.TITLE_SCREEN
				case "LOGOS":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.LOGOS
				case "SCREENSHOTS":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.SCREENSHOTS
				}
			} else if option.Item.Text == "Art Fuzzy Search Threshold" {
				appState.Config.FuzzySearchThreshold = option.Options[option.SelectedOption].Value.(float64)
			} else if option.Item.Text == "Hide Empty Platforms" {
				appState.Config.HideEmpty = option.Options[option.SelectedOption].Value.(bool)
			} else if option.Item.Text == "Show Art" {
				appState.Config.ShowArt = option.Options[option.SelectedOption].Value.(bool)
			} else if option.Item.Text == "Log Level" {
				logLevelValue := option.Options[option.SelectedOption].Value.(string)
				appState.Config.LogLevel = logLevelValue
			} else if option.Item.Text == "Play History Show Archive Tag" {
				appState.Config.PlayHistoryShowArchives = option.Options[option.SelectedOption].Value.(bool)
			} else if option.Item.Text == "Play History Show Collection Tags" {
				appState.Config.PlayHistoryShowCollections = option.Options[option.SelectedOption].Value.(bool)
			}
		}

		err := utils.SaveConfig(appState.Config)
		if err != nil {
			logger.Error("Error saving config", zap.Error(err))
			return nil, 0, err
		}

		state.UpdateAppState(appState)

		return result, 0, nil
	}

	return nil, 2, nil
}
