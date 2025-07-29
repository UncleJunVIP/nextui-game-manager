package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"strings"
)

type PlayHistoryFilterScreen struct {
	Console         		string
	SearchFilter			string
	GameAggregate			models.PlayHistoryAggregate
	Game                 	shared.Item
	RomDirectory         	shared.RomDirectory
	PreviousRomDirectory 	shared.RomDirectory
	PlayHistoryOrigin		bool
	PlayHistoryFilterList	[]models.PlayHistorySearchFilter
	MenuDepth				int
}

func InitPlayHistoryFilterScreen(console string, searchFilter string, gameAggregate models.PlayHistoryAggregate, game shared.Item, romDirectory shared.RomDirectory, 
	previousRomDirectory shared.RomDirectory, playHistoryOrigin bool, filterList []models.PlayHistorySearchFilter, menuDepth int) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		Console:              	console,
		SearchFilter:         	searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayHistoryOrigin: 		playHistoryOrigin,
		PlayHistoryFilterList:	filterList,
		MenuDepth:				menuDepth,
	}
}

func InitPlayHistoryFilterScreenFromGameList(console string, filterList []models.PlayHistorySearchFilter) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		Console:              	console,
		PlayHistoryFilterList:	filterList,
		MenuDepth:				1,
	}
}

func InitPlayHistoryFilterScreenFromHistoryList(filterList []models.PlayHistorySearchFilter) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		PlayHistoryFilterList:	filterList,
		MenuDepth:				1,
	}
}

func (phfs PlayHistoryFilterScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryFilter
}

// Lists available play History consoles
func (phfs PlayHistoryFilterScreen) Draw() (item interface{}, exitCode int, e error) {
	title := "Filter"

	currentFilter := models.PlayHistorySearchFilter{}
	if len(phfs.PlayHistoryFilterList) != 0 {
		currentFilter = phfs.PlayHistoryFilterList[len(phfs.PlayHistoryFilterList)-1]
		title = title + ": " + currentFilter.DisplayName
	}

	romIds := []int{}
	if len(phfs.GameAggregate.Id) > 0 {
		romIds = phfs.GameAggregate.Id
		title = title + " (" + phfs.GameAggregate.Name + ")"
	} else if phfs.Console != "" {
		gamePlayMap, _, _ := state.GetPlayMaps()
		gamesList := gamePlayMap[phfs.Console]
		for _, game := range gamesList {
			romIds = append(romIds, game.Id...)
		}

		startIndex := strings.LastIndex(phfs.Console, "(")
		endIndex := strings.LastIndex(phfs.Console, ")")
		if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
			title = title + " (" + phfs.Console + ")"
		} else {
			title = title + phfs.Console[startIndex : endIndex+1]
		}
	}

	filterList := []models.PlayHistorySearchFilter{}
	if currentFilter.FilterType < 2 {
		filterList = utils.GenFiltersList(romIds, currentFilter.SqlFilter, currentFilter.FilterType)
	}

	var menuItems []gaba.MenuItem
	
	for _, filter := range filterList {
		filterItem := gaba.MenuItem{
			Text:     fmt.Sprintf("%s : %.1fH",filter.DisplayName, min(9999, float64(filter.PlayTime)/3600.0)),
			Selected: false,
			Focused:  false,
			Metadata: filter,
		}
		menuItems = append(menuItems, filterItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	//options.SmallTitle = true
	options.EmptyMessage = "Max Filter Depth\nX to save filter"
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "X", HelpText: "Save Filter"},
		{ButtonName: "A", HelpText: "Select"},
	}

	if len(phfs.PlayHistoryFilterList) > 0 {
		options.FooterHelpItems = append([]gaba.FooterHelpItem{{ButtonName: "B", HelpText: "Back"},}, options.FooterHelpItems...)
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		newFilter := selection.Unwrap().SelectedItem.Metadata.(models.PlayHistorySearchFilter)
		return newFilter, 0, nil
	}

	state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
	return nil, 2, nil
}
