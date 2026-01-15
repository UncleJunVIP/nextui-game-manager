package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type PlayHistoryActionsScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitPlayHistoryActionsScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string) PlayHistoryActionsScreen {
	return PlayHistoryActionsScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (ptas PlayHistoryActionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryActions
}

func (ptas PlayHistoryActionsScreen) Draw() (action interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	existingArtFilename, err := utils.FindExistingArt(ptas.Game.Filename, ptas.RomDirectory)
	if err != nil {
		logger.Error("failed to find existing arts", zap.Error(err))
	}

	actions := models.ActionKeys

	if existingArtFilename == "" {
		actions = utils.InsertIntoSlice(actions, 1, "Download Art")
	} else {
		actions = utils.InsertIntoSlice(actions, 1, "Delete Art")
	}

	var actionEntries []gabagool.MenuItem
	for _, action := range actions {
		actionEntries = append(actionEntries, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(ptas.Game.DisplayName, actionEntries)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return selection.Items[selection.Selected[0]].Text, 0, nil
	}

	return nil, 2, nil
}
