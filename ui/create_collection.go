package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"path/filepath"
	"strings"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type CreateCollectionScreen struct {
	Games                shared.Items
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitCreateCollectionScreen(games shared.Items, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) CreateCollectionScreen {
	return CreateCollectionScreen{
		Games:                games,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (c CreateCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionCreate
}

func (c CreateCollectionScreen) Draw() (collection interface{}, exitCode int, e error) {
	res, err := gaba.Keyboard("", "")

	if err != nil {
		if err == gaba.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if res != nil && res.Text != "" {
		newCollectionName := strings.Trim(res.Text, " ")

		if newCollectionName == "" {
			return nil, 2, nil
		}

		utils.AddCollectionGames(state.GetCollectionMap(), models.Collection{
			DisplayName:    newCollectionName,
			CollectionFile: filepath.Join(utils.GetCollectionDirectory(), newCollectionName+".txt"),
		}, c.Games)
		state.ClearCollectionMap()

		message := fmt.Sprintf("Created %s!", newCollectionName)

		if len(c.Games) > 1 {
			message = fmt.Sprintf("%s\nAlso added %d Games!", message, len(c.Games))
		} else {
			message = fmt.Sprintf("%s\nAlso added %s!", message, c.Games[0].DisplayName)
		}

		utils.ShowTimedMessage(message, time.Second*2)

		return nil, 0, nil
	}

	return nil, 2, nil
}
