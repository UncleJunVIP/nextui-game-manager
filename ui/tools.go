package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
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
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Metadata, 0, nil
	}

	return nil, 2, nil
}
