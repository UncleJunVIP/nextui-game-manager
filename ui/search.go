package ui

import (
	"nextui-game-manager/models"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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
	query, err := gabagool.Keyboard("")
	if err != nil {
		return nil, -1, err
	}

	if query.IsSome() {
		return query.Unwrap(), 0, nil
	}

	return nil, 2, nil
}
