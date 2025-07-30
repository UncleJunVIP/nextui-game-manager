package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type ActionsScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitActionsScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string) ActionsScreen {
	return ActionsScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (a ActionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Actions
}

func (a ActionsScreen) Draw() (action interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	existingArtFilename, err := utils.FindExistingArt(a.Game.Filename, a.RomDirectory)
	if err != nil {
		logger.Error("failed to find existing arts", zap.Error(err))
	}

	actions := models.ActionKeys

	if existingArtFilename == "" {
		actions = utils.InsertIntoSlice(actions, 1, "Download Art")
	} else {
		actions = utils.InsertIntoSlice(actions, 1, "Delete Art")
	}

	_, hasCheat := utils.FindExistingCheat(a.Game, a.RomDirectory)

	if hasCheat {
		actions = utils.InsertIntoSlice(actions, 2, "Delete Cheat File")
	} else if utils.CheatFileAvailable(a.Game, a.RomDirectory) {
		actions = utils.InsertIntoSlice(actions, 2, "Download Cheat File")
	}

	gamePlayMap, _, _ := state.GetPlayMaps()
	gameAggregate, _ := utils.CollectGameAggregateFromGame(a.Game, gamePlayMap)
	if gameAggregate.PlayCountTotal != 0 {
		actions = append(actions, "View Play Details")
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

	options := gabagool.DefaultListOptions(a.Game.DisplayName, actionEntries)

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
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Text, 0, nil
	}

	return nil, 2, nil
}
