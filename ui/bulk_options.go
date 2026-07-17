package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type BulkOptionsScreen struct {
	Games                []shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitBulkOptionsScreen(games []shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string) BulkOptionsScreen {
	return BulkOptionsScreen{
		Games:                games,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (b BulkOptionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.BulkActions
}

func (b BulkOptionsScreen) Draw() (action interface{}, exitCode int, e error) {
	actions := models.BulkActionKeys

	var actionEntries []gabagool.MenuItem
	for _, action := range actions {
		actionEntries = append(actionEntries, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(fmt.Sprintf("Manage %d Games", len(b.Games)), actionEntries)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.UseSmallTitle = true
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
