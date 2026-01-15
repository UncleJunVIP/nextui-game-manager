package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"qlova.tech/sum"
)

type ToolsScreen struct {
}

func InitToolsScreen() ToolsScreen {
	return ToolsScreen{}
}

func (ts ToolsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Tools
}

func (ts ToolsScreen) Draw() (value interface{}, exitCode int, e error) {
	var menuItems []gabagool.MenuItem

	menuItems = append(menuItems, gabagool.MenuItem{
		Text:     "Global Actions",
		Selected: false,
		Focused:  false,
		Metadata: "Global Actions",
	})

	menuItems = append(menuItems, gabagool.MenuItem{
		Text:     "Play History",
		Selected: false,
		Focused:  false,
		Metadata: "Play History",
	})

	options := gabagool.DefaultListOptions("Tools", menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

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
		return selection.Items[selection.Selected[0]].Metadata, 0, nil
	}

	return nil, 2, nil
}
