package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"qlova.tech/sum"
)

type PlayHistoryGamesListScreen struct {
	Console               string
	PlayHistoryFilterList []models.PlayHistorySearchFilter
}

func InitPlayHistoryGamesListScreen(console string, filterList []models.PlayHistorySearchFilter) PlayHistoryGamesListScreen {
	return PlayHistoryGamesListScreen{
		Console:               console,
		PlayHistoryFilterList: filterList,
	}
}

func (ptgls PlayHistoryGamesListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryGameList
}

func (ptgls PlayHistoryGamesListScreen) Draw() (item interface{}, exitCode int, e error) {
	appState := state.GetAppState()

	var gamePlayMap map[string][]models.PlayHistoryAggregate
	var consoleMap map[string]int
	var title string
	if len(ptgls.PlayHistoryFilterList) == 0 {
		gamePlayMap, consoleMap, _ = state.GetPlayMaps()
		title = fmt.Sprintf("%.1fH : %s", float64(consoleMap[ptgls.Console])/3600.0, ptgls.Console)
	} else {
		currentFilter := ptgls.PlayHistoryFilterList[len(ptgls.PlayHistoryFilterList)-1]
		gamePlayMap, consoleMap, _ = utils.GenerateCurrentGameStats(currentFilter.SqlFilter)
		title = fmt.Sprintf("%s: %s", currentFilter.DisplayName, ptgls.Console)
	}

	gamesList := gamePlayMap[ptgls.Console]

	var menuItems []gabagool.MenuItem
	collectionMap := state.GetCollectionMap()

	for _, gamePlayAggregate := range gamesList {
		playHours := min(999, float64(gamePlayAggregate.PlayTimeTotal)/3600.0)
		romHomeStatus := utils.FindRomHomeFromAggregate(gamePlayAggregate, appState.Config.PlayHistoryShowArchives)
		collections := collectionMap[gamePlayAggregate.Name]
		collectionString := ""
		if appState.Config.PlayHistoryShowCollections {
			for _, collection := range collections {
				collectionString = collectionString + string(collection.DisplayName[0])
			}
			if collectionString != "" {
				collectionString = "[" + collectionString + "] "
			}
		}
		gameItem := gabagool.MenuItem{
			Text:     fmt.Sprintf("%.1fH %s%s: %s", playHours, romHomeStatus, collectionString, gamePlayAggregate.Name),
			Selected: false,
			Focused:  false,
			Metadata: gamePlayAggregate,
		}
		menuItems = append(menuItems, gameItem)
	}

	options := gabagool.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Play Records Found"
	options.ActionButton = constants.VirtualButtonX
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Filter"},
		{ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Details"},
	}

	options.HelpButton = constants.VirtualButtonMenu
	options.HelpTitle = "Tag Details"
	options.HelpText = []string{
		"(+) => Rom location matches play history",
		"(-) => Missing Rom, 'Orphaned' history",
		"(A) => Archived Rom, first letter of archive",
		"[ABC] => Collections containing Rom",
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
		game := selection.Items[selection.Selected[0]].Metadata.(models.PlayHistoryAggregate)
		return game, 0, nil
	}

	return nil, 2, nil
}
