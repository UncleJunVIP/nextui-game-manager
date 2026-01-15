package ui

import (
	"cmp"
	"fmt"
	"maps"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"slices"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"qlova.tech/sum"
)

type PlayHistoryListScreen struct {
	PlayHistoryFilterList []models.PlayHistorySearchFilter
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
		title = fmt.Sprintf("%s: %.1f Total Hours Played", currentFilter.DisplayName, float64(totalPlay)/3600.0)
	}

	if consolePlayMap == nil || len(consolePlayMap) == 0 {
		return nil, 404, nil
	}

	var menuItems []gabagool.MenuItem
	consoles := slices.SortedStableFunc(maps.Keys(consolePlayMap), func(a, b string) int {
		return cmp.Compare(consolePlayMap[b], consolePlayMap[a])
	})
	for _, console := range consoles {
		consoleItem := gabagool.MenuItem{
			Text:     fmt.Sprintf("%.1fH : %s", min(9999, float64(consolePlayMap[console])/3600.0), console),
			Selected: false,
			Focused:  false,
			Metadata: console,
		}
		menuItems = append(menuItems, consoleItem)
	}

	options := gabagool.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.ActionButton = constants.VirtualButtonX
	options.SmallTitle = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Filter"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 && selection.Action == gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		console := selection.Items[selection.Selected[0]].Metadata.(string)
		return console, 0, nil
	}

	return nil, 2, nil
}
