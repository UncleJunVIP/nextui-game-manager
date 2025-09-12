package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"strings"
	"time"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type ArchiveCreateScreen struct {
	Games                []shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitArchiveCreateScreen(gamesList []shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) ArchiveCreateScreen {
	return ArchiveCreateScreen{
		Games:                gamesList,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (acs ArchiveCreateScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveCreate
}

func (acs ArchiveCreateScreen) Draw() (value interface{}, exitCode int, e error) {
	res, err := gaba.Keyboard("")
	if err != nil {
		return nil, -1, err
	}

	if res.IsSome() {
		newArchiveName := res.Unwrap()

		if newArchiveName == "" || newArchiveName == "." || strings.Contains(newArchiveName, "/") {
			return nil, 2, nil
		}

		newArchiveName = utils.PrepArchiveName(newArchiveName)

		if newArchiveName == ".media" {
			utils.ShowTimedMessage(".media folder is reserved for themes.\nPlease choose a different name.", time.Second*2)
			return nil, 0, nil
		}

		dirErr := utils.EnsureDirectoryExists(utils.GetArchiveRoot(newArchiveName))

		message := fmt.Sprintf("Created %s!", newArchiveName)

		if dirErr != nil {
			message = fmt.Sprintf("Creation of %s failed.\nTry a different name.", newArchiveName)
		}

		utils.ShowTimedMessage(message, time.Second*2)

		return nil, 0, nil
	}

	return nil, 2, nil
}
