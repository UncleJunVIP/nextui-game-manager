package ui

import (
	"nextui-game-manager/models"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type Search struct {
	RomDirectory shared.RomDirectory
}

func InitSearch(romDirectory shared.RomDirectory) Search {
	return Search{
		RomDirectory: romDirectory,
	}
}

func (s Search) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.SearchBox
}

func (s Search) Draw() (value interface{}, exitCode int, e error) {
	query, err := gabagool.Keyboard("", "")
	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if query != nil && query.Text != "" {
		return query.Text, 0, nil
	}

	return nil, 2, nil
}
