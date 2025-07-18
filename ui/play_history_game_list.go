package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type PlayHistoryGamesListScreen struct {
	Console         string
	SearchFilter	string
}

func InitPlayHistoryGamesListScreen(console string, searchFilter string) PlayHistoryGamesListScreen {
	return PlayHistoryGamesListScreen{
		Console:              console,
		SearchFilter:         searchFilter,
	}
}

func (ptgls PlayHistoryGamesListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryGameList
}

func (ptgls PlayHistoryGamesListScreen) Draw() (item interface{}, exitCode int, e error) {
	appState := state.GetAppState()
	gamePlayMap, consoleMap, _ := state.GetPlayMaps()

	title := fmt.Sprintf("%.1fH : %s", float64(consoleMap[ptgls.Console])/3600.0, ptgls.Console)

	gamesList := gamePlayMap[ptgls.Console]

	if ptgls.SearchFilter != "" {
		title = "[Search: \"" + ptgls.SearchFilter + "\"]"
		gamesList = utils.FilterPlayList(gamesList, ptgls.SearchFilter)
	}

	var menuItems []gaba.MenuItem
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
		gameItem := gaba.MenuItem{
			Text:     fmt.Sprintf("%.1fH %s%s: %s", playHours, romHomeStatus, collectionString, gamePlayAggregate.Name),
			Selected: false,
			Focused:  false,
			Metadata: gamePlayAggregate,
		}
		menuItems = append(menuItems, gameItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Play Records Found"
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Details"},
	}

	options.EnableHelp = true
	options.HelpTitle = "Tag Details"
	options.HelpText = []string{
		"(+) => Rom location matches play history",
		"(-) => Missing Rom, 'Orphaned' history",
		"(A) => Archived Rom, first letter of archive",
		"[ABC] => Collections containing Rom",
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		query, err := gaba.Keyboard("")

		if err != nil {
			return nil, 1, err
		}

		if query.IsSome() {
			return query.Unwrap(), 4, nil
		}

		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		game := selection.Unwrap().SelectedItem.Metadata.(models.PlayHistoryAggregate)
		return game, 0, nil
	}

	return nil, 2, nil
}
