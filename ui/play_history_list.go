package ui

import (
	"cmp"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"maps"
	"slices"
)

type PlayHistoryListScreen struct {
	PlayHistoryFilterList	[]models.PlayHistorySearchFilter
}

func InitPlayHistoryListScreen(filterList []models.PlayHistorySearchFilter) PlayHistoryListScreen {
	return PlayHistoryListScreen{
		PlayHistoryFilterList: filterList,
	}
}

func (ptls PlayHistoryListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryList
}

// Lists available play History consoles
func (ptls PlayHistoryListScreen) Draw() (item interface{}, exitCode int, e error) {
	var consolePlayMap map[string]int
	var totalPlay int 
	var title string
	if len(ptls.PlayHistoryFilterList) == 0 {
		_, consolePlayMap, totalPlay = state.GetPlayMaps()
		title = fmt.Sprintf("%.1f Total Hours Played", float64(totalPlay)/3600.0)
	} else {
		currentFilter := ptls.PlayHistoryFilterList[len(ptls.PlayHistoryFilterList)-1]
		_, consolePlayMap, totalPlay = utils.GenerateCurrentGameStats(currentFilter.SqlFilter)
		title = fmt.Sprintf("%s : %.1f Total Hours Played", currentFilter.DisplayName, float64(totalPlay)/3600.0)
	}

	if consolePlayMap == nil || len(consolePlayMap) == 0 {
		return nil, 404, nil
	}

	var menuItems []gaba.MenuItem
	consoles := slices.SortedStableFunc(maps.Keys(consolePlayMap), func(a, b string) int {
		return cmp.Compare(consolePlayMap[b], consolePlayMap[a])
	})
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
		{ButtonName: "X", HelpText: "Filter"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		console := selection.Unwrap().SelectedItem.Metadata.(string)
		return console, 0, nil
	}

	return nil, 2, nil
}
