package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type PlayHistoryGameHistoryScreen struct {
	Console               string
	SearchFilter          string
	GameAggregate         models.PlayHistoryAggregate
	Game                  shared.Item
	RomDirectory          shared.RomDirectory
	PreviousRomDirectory  shared.RomDirectory
	PlayHistoryOrigin     bool
	PlayHistoryFilterList []models.PlayHistorySearchFilter
}

func InitPlayHistoryGameHistoryScreen(console string, searchFilter string, gameAggregate models.PlayHistoryAggregate, game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, playHistoryOrigin bool, filterList []models.PlayHistorySearchFilter) PlayHistoryGameHistoryScreen {
	return PlayHistoryGameHistoryScreen{
		Console:               console,
		SearchFilter:          searchFilter,
		GameAggregate:         gameAggregate,
		Game:                  game,
		RomDirectory:          romDirectory,
		PreviousRomDirectory:  previousRomDirectory,
		PlayHistoryOrigin:     playHistoryOrigin,
		PlayHistoryFilterList: filterList,
	}
}

func (ptghs PlayHistoryGameHistoryScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryGameHistory
}

func (ptghs PlayHistoryGameHistoryScreen) Draw() (item interface{}, exitCode int, e error) {
	var playHistory []models.PlayHistoryGranular
	var title string
	if len(ptghs.PlayHistoryFilterList) == 0 {
		playHistory = utils.GenerateSingleGameGranularRecords(ptghs.GameAggregate.Id, "")
		title = ptghs.GameAggregate.Name
	} else {
		currentFilter := ptghs.PlayHistoryFilterList[len(ptghs.PlayHistoryFilterList)-1]
		playHistory = utils.GenerateSingleGameGranularRecords(ptghs.GameAggregate.Id, currentFilter.SqlFilter)
		title = fmt.Sprintf("%s: %s", currentFilter.DisplayName, ptghs.GameAggregate.Name)
	}

	var menuItems []gabagool.MenuItem
	for _, playRecord := range playHistory {
		duration := utils.ConvertSecondsToHumanReadableAbbreviated(playRecord.PlayTime)
		startTime := time.Unix(int64(playRecord.StartTime), 0).Format(time.UnixDate)
		playItem := gabagool.MenuItem{
			Text:     fmt.Sprintf("%s ~ %s", startTime, duration),
			Selected: false,
			Focused:  false,
			Metadata: playRecord,
		}
		menuItems = append(menuItems, playItem)
	}

	options := gabagool.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.UseSmallTitle = true
	options.EmptyMessage = "No Play Records Found"
	options.ActionButton = constants.VirtualButtonX
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Filter"},
		//{ButtonName: "A", HelpText: "Update"},
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
		// game := selection.Items[selection.Selected[0]].Metadata.(string)
		// return game, 0, nil
		return nil, 0, nil
	}

	return nil, 2, nil
}
