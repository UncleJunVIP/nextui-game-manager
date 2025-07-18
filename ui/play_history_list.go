package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"qlova.tech/sum"
	"maps"
	"slices"
)

type PlayHistoryListScreen struct {}

func InitPlayHistoryListScreen() PlayHistoryListScreen {
	return PlayHistoryListScreen{}
}

func (ptls PlayHistoryListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryList
}

// Lists available play History consoles
func (ptls PlayHistoryListScreen) Draw() (item interface{}, exitCode int, e error) {
	_, consolePlayMap, totalPlay := state.GetPlayMaps()

	title := fmt.Sprintf("%.1f Total Hours Played", float64(totalPlay)/3600.0)

	if consolePlayMap == nil || len(consolePlayMap) == 0 {
		return nil, 404, nil
	}

	var menuItems []gaba.MenuItem
	consoles := slices.Sorted(maps.Keys(consolePlayMap))
	for _, console := range consoles {
		consoleItem := gaba.MenuItem{
			Text:     fmt.Sprintf("%.1fH : %s", min(9999, float64(consolePlayMap[console])/3600.0), console),
			Selected: false,
			Focused:  false,
			Metadata: console,
		}
		menuItems = append(menuItems, consoleItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		console := selection.Unwrap().SelectedItem.Metadata.(string)
		return console, 0, nil
	}

	return nil, 2, nil
}
