package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"time"
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

	artPath, _ := gaba.ProcessMessage(fmt.Sprintf("Finding art for %s...", da.Game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		artPath := utils.FindArt(da.RomDirectory, da.Game, da.DownloadType, state.GetAppState().Config.FuzzySearchThreshold)
		return artPath, nil
	})

	if artPath.Result.(string) == "" {
		utils.ShowTimedMessage("No art found!", time.Second*3)
		return shared.Item{}, 404, nil
	}

	result, err := gaba.ConfirmationMessage("Found This Art!",
		[]gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "I'll Find My Own"},
			{ButtonName: "A", HelpText: "Use It!"},
		},
		gaba.MessageOptions{
			ImagePath: artPath.Result.(string),
		})

	if err != nil || result.IsNone() {
		common.DeleteFile(artPath.Result.(string))
	}

	return nil, 0, nil

}
