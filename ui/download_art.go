package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"nextui-game-manager/common"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type DownloadArtScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
	DownloadType         sum.Int[shared.ArtDownloadType]
}

func InitDownloadArtScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string, downloadType sum.Int[shared.ArtDownloadType]) DownloadArtScreen {
	return DownloadArtScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
		DownloadType:         downloadType,
	}
}

func (da DownloadArtScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadArt
}

func (da DownloadArtScreen) Draw() (value interface{}, exitCode int, e error) {

	artPath, _ := gabagool.ProcessMessage(fmt.Sprintf("Finding art for %s...", da.Game.DisplayName), gabagool.ProcessMessageOptions{}, func() (string, error) {
		artPath := utils.FindArt(da.RomDirectory, da.Game, da.DownloadType, state.GetAppState().Config.FuzzySearchThreshold)
		return artPath, nil
	})

	if artPath == "" {
		utils.ShowTimedMessage("No art found!", time.Second*3)
		return shared.Item{}, 404, nil
	}

	result, err := gabagool.ConfirmationMessage("Found This Art!",
		[]gabagool.FooterHelpItem{
			{ButtonName: "B", HelpText: "I'll Find My Own"},
			{ButtonName: "A", HelpText: "Use It!"},
		},
		gabagool.MessageOptions{
			ImagePath: artPath,
		})

	if err != nil || result == nil || !result.Confirmed {
		common.DeleteFile(artPath)
	}

	return nil, 0, nil

}
