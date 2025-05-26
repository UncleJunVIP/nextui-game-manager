package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/veandco/go-sdl2/sdl"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type CollectionManagement struct {
	Collection   models.Collection
	SearchFilter string
}

func InitCollectionManagement(collection models.Collection) CollectionManagement {
	return CollectionManagement{
		Collection: collection,
	}
}

func (c CollectionManagement) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionManagement
}

func (c CollectionManagement) Draw() (value interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var err error
	c.Collection, err = utils.ReadCollection(c.Collection)
	if err != nil {
		logger.Error("failed to read collection", zap.Error(err))
		return shared.Item{}, 1, err
	}

	var menuItems []gaba.MenuItem

	for _, g := range c.Collection.Games {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     g.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: g,
		})
	}

	options := gaba.DefaultListOptions(c.Collection.DisplayName, menuItems)
	options.EnableAction = true
	options.EnableHelp = true
	options.HelpTitle = "Collection Management Controls"

	options.EnableMultiSelect = true
	options.MultiSelectKey = sdl.K_SPACE
	options.MultiSelectButton = gaba.ButtonSelect

	options.HelpText = []string{
		"• X: Open Options",
	}

	if len(menuItems) > 1 {
		options.EnableReordering = true
		options.ReorderKey = sdl.K_y
		options.ReorderButton = gaba.ButtonY
		options.HelpText = append(options.HelpText, "• Y: Toggle Reordering Mode")
		options.HelpText = append(options.HelpText, "• ↕: Move Selection")
	}

	options.HelpText = append(options.HelpText, "• Select: Toggle Multi-Select Removal")

	options.HelpText = append(options.HelpText, "• A: Remove ROM / Add to Remove Selection")
	options.HelpText = append(options.HelpText, "• Start: Remove Selected")
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Options"},
		{ButtonName: "Menu", HelpText: "Controls"},
	}

	selection, err := gaba.List(options)

	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {

	}

	rawItems := selection.Unwrap().Items

	var games shared.Items
	for _, item := range rawItems {
		games = append(games, item.Metadata.(shared.Item))
	}

	c.Collection.Games = games

	utils.SaveCollection(c.Collection)

	return c.Collection, 2, nil
}
