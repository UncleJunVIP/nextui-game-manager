package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"time"
)

type PlayHistoryGameHistoryScreen struct {
	Console         		string
	SearchFilter			string
	GameAggregate			models.PlayHistoryAggregate
	Game                 	shared.Item
	RomDirectory         	shared.RomDirectory
	PreviousRomDirectory 	shared.RomDirectory
	PlayHistoryOrigin		bool
}

func InitPlayHistoryGameHistoryScreen(console string, searchFilter string, gameAggregate models.PlayHistoryAggregate, game shared.Item, 
	romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, playHistoryOrigin bool) PlayHistoryGameHistoryScreen {
	return PlayHistoryGameHistoryScreen{
		Console:              	console,
		SearchFilter:         	searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayHistoryOrigin: 		playHistoryOrigin,
	}
}

func (ptgls PlayHistoryGameHistoryScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryGameHistory
}

func (ptghs PlayHistoryGameHistoryScreen) Draw() (item interface{}, exitCode int, e error) {
	title := ptghs.GameAggregate.Name + " Play History"

	playHistory := utils.GenerateSingleGameGranularRecords(ptghs.GameAggregate.Id)

	var menuItems []gaba.MenuItem
	for _, playRecord := range playHistory {
		duration := utils.ConvertSecondsToHumanReadableAbbreviated(playRecord.PlayTime)
		startTime := time.Unix(int64(playRecord.StartTime), 0).Format(time.UnixDate)
		playItem := gaba.MenuItem{
			Text:     fmt.Sprintf("%s ~ %s", startTime, duration),
			Selected: false,
			Focused:  false,
			Metadata: playRecord,
		}
		menuItems = append(menuItems, playItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Play Records Found"
	//options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		//{ButtonName: "X", HelpText: "Actions"},
		//{ButtonName: "A", HelpText: "Update"},
	}

	// selection, err := gaba.List(options)
	_, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	// if selection.IsSome() && selection.Unwrap().ActionTriggered {
	// 	state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
	// 	query, err := gaba.Keyboard("")

	// 	if err != nil {
	// 		return nil, 1, err
	// 	}

	// 	if query.IsSome() {
	// 		return query.Unwrap(), 4, nil
	// 	}

	// 	return nil, 4, nil
	// } else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
	// 	state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
	// 	game := selection.Unwrap().SelectedItem.Metadata.(string)
	// 	return game, 0, nil
	// }

	return nil, 2, nil
}
