package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"slices"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
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

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.EnableHelp = true
	options.HelpTitle = "Collection Management Controls"
	options.EmptyMessage = "This collection is empty.\nAdd some games you silly goose!"

	options.EnableMultiSelect = true
	options.MultiSelectButton = gaba.InternalButtonSelect

	options.HelpText = []string{
		"• X: Open Options",
	}

	if len(menuItems) > 1 {
		options.EnableReordering = true
		options.ReorderButton = gaba.InternalButtonY
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

	selection, _ := gaba.List(options)

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		selected := selection.Unwrap()

		var message string

		if len(selected.SelectedItems) == 1 {
			message = fmt.Sprintf("Remove %s from %s?", selected.SelectedItem.Text, c.Collection.DisplayName)
		} else {
			message = fmt.Sprintf("Remove %d ROMs from %s?", len(selected.SelectedItems), c.Collection.DisplayName)
		}

		if utils.ConfirmBulkAction(message) {
			var games shared.Items
			for _, item := range c.Collection.Games {
				if !slices.ContainsFunc(selected.SelectedItems, func(i *gaba.MenuItem) bool {
					return item.DisplayName == i.Text
				}) {
					games = append(games, item)
				}
			}

			c.Collection.Games = games

			utils.SaveCollection(c.Collection)
		}

		return c.Collection, 0, nil
	} else {
		rawItems := selection.Unwrap().Items

		var games shared.Items
		for _, item := range rawItems {
			games = append(games, item.Metadata.(shared.Item))
		}

		c.Collection.Games = games

		utils.SaveCollection(c.Collection)

		return c.Collection, 2, nil
	}
}
