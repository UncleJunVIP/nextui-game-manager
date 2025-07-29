package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"strconv"
	"time"
)

type PlayHistoryGameDetailsScreen struct {
	Console         		string
	SearchFilter			string
	GameAggregate			models.PlayHistoryAggregate
	Game                 	shared.Item
	RomDirectory         	shared.RomDirectory
	PreviousRomDirectory 	shared.RomDirectory
	PlayHistoryOrigin		bool
	PlayHistoryFilterList	[]models.PlayHistorySearchFilter
}

func InitPlayHistoryGameDetailsScreenFromPlayHistory(console string, gameAggregate models.PlayHistoryAggregate, filterList []models.PlayHistorySearchFilter) PlayHistoryGameDetailsScreen {
	return PlayHistoryGameDetailsScreen{
		Console:      		console,
		GameAggregate: 		gameAggregate,
		PlayHistoryOrigin: 	true,
		PlayHistoryFilterList:  filterList,
	}
}

func InitPlayHistoryGameDetailsScreenFromActions(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) PlayHistoryGameDetailsScreen {
	gamePlayMap, _, _ := state.GetPlayMaps()
	gameAggregate, console := utils.CollectGameAggregateFromGame(game, gamePlayMap)
	return PlayHistoryGameDetailsScreen{
		Console:				console,
		SearchFilter: 			searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayHistoryOrigin: 		false,
	}
}

func InitPlayHistoryGameDetailsScreenFromSelf(console string, searchFilter string, gameAggregate models.PlayHistoryAggregate, game shared.Item, 
	romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, playHistoryOrigin bool, filterList []models.PlayHistorySearchFilter) PlayHistoryGameDetailsScreen {
	return PlayHistoryGameDetailsScreen{
		Console:				console,
		SearchFilter: 			searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayHistoryOrigin: 		playHistoryOrigin,
		PlayHistoryFilterList:  filterList,
	}
}

func (ptgds PlayHistoryGameDetailsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayHistoryGameDetails
}

func (ptgds PlayHistoryGameDetailsScreen) Draw() (selection interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	gamePlayMap, consolePlayMap, totalPlay := state.GetPlayMaps()
	title := ptgds.GameAggregate.Name

	gameAggregate := utils.CollectGameAggregateFromGameName(ptgds.GameAggregate.Name, ptgds.Console, gamePlayMap)

	var sections []gaba.Section

	sections = append(sections, gaba.NewInfoSection(
		title,
		[]gaba.MetadataItem{
			{Label: "Console", 			Value: ptgds.Console},
			{Label: "First Played", 	Value: gameAggregate.FirstPlayedTime.Format(time.UnixDate)},
			{Label: "Last Played", 		Value: gameAggregate.LastPlayedTime.Format(time.UnixDate)},
			{Label: "Play Sessions", 	Value: strconv.Itoa(gameAggregate.PlayCountTotal)},
			{Label: "Total Play Time", 	Value: utils.ConvertSecondsToHumanReadable(gameAggregate.PlayTimeTotal)},
			{Label: "Average Session", 	Value: utils.ConvertSecondsToHumanReadable(gameAggregate.PlayTimeTotal/gameAggregate.PlayCountTotal)},
			{Label: "Pct of Total", 	Value: fmt.Sprintf("%.2f%%", (float64(gameAggregate.PlayTimeTotal)/float64(totalPlay))*100)},
			{Label: "Pct of Console", 	Value: fmt.Sprintf("%.2f%%", (float64(gameAggregate.PlayTimeTotal)/float64(consolePlayMap[ptgds.Console]))*100)},
		},
	))

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "History"},
	}

	sel, err := gaba.DetailScreen("Play Stats", options, footerItems)
	if err != nil {
		logger.Error("Unable to display Play History screen", zap.Error(err))
		return nil, -1, err
	}

	if sel.IsNone() {
		return nil, 2, nil
	}

	return nil, 0, nil
}
