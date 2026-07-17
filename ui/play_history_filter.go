package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type PlayHistoryFilterScreen struct {
	Console               string
	SearchFilter          string
	GameAggregate         models.PlayHistoryAggregate
	Game                  shared.Item
	RomDirectory          shared.RomDirectory
	PreviousRomDirectory  shared.RomDirectory
	PlayHistoryOrigin     bool
	PlayHistoryFilterList []models.PlayHistorySearchFilter
	MenuDepth             int
}

func InitPlayHistoryFilterScreen(console string, searchFilter string, gameAggregate models.PlayHistoryAggregate, game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, playHistoryOrigin bool, filterList []models.PlayHistorySearchFilter, menuDepth int) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		Console:               console,
		SearchFilter:          searchFilter,
		GameAggregate:         gameAggregate,
		Game:                  game,
		RomDirectory:          romDirectory,
		PreviousRomDirectory:  previousRomDirectory,
		PlayHistoryOrigin:     playHistoryOrigin,
		PlayHistoryFilterList: filterList,
		MenuDepth:             menuDepth,
	}
}

func InitPlayHistoryFilterScreenFromGameList(console string, filterList []models.PlayHistorySearchFilter) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		Console:               console,
		PlayHistoryFilterList: filterList,
		MenuDepth:             1,
	}
}

func InitPlayHistoryFilterScreenFromHistoryList(filterList []models.PlayHistorySearchFilter) PlayHistoryFilterScreen {
	return PlayHistoryFilterScreen{
		PlayHistoryFilterList: filterList,
		MenuDepth:             1,
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
			title = title + " " + phfs.Console[startIndex:endIndex+1]
		}
	}

	filterList := []models.PlayHistorySearchFilter{}
	if currentFilter.FilterType < 2 {
		filterList = utils.GenFiltersList(romIds, currentFilter.SqlFilter, currentFilter.FilterType)
	}

	var menuItems []gabagool.MenuItem

	for _, filter := range filterList {
		filterItem := gabagool.MenuItem{
			Text:     fmt.Sprintf("%s : %s", filter.DisplayName, utils.ConvertSecondsToHumanReadable(filter.PlayTime)),
			Selected: false,
			Focused:  false,
			Metadata: filter,
		}
		menuItems = append(menuItems, filterItem)
	}

	options := gabagool.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.ActionButton = constants.VirtualButtonX
	//options.UseSmallTitle = true
	options.EmptyMessage = "Max Filter Depth\nX to save filter"
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "X", HelpText: "Save Filter"},
		{ButtonName: "A", HelpText: "Select"},
	}

	if len(phfs.PlayHistoryFilterList) > 0 {
		options.FooterHelpItems = append([]gabagool.FooterHelpItem{{ButtonName: "B", HelpText: "Back"}}, options.FooterHelpItems...)
	}

	selection, err := gabagool.List(options)
	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 && selection.Action == gabagool.ListActionTriggered {
		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		newFilter := selection.Items[selection.Selected[0]].Metadata.(models.PlayHistorySearchFilter)
		return newFilter, 0, nil
	}

	return nil, 2, nil
}
