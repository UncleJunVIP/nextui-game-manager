package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/shared"
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

	var menuItems []gabagool.MenuItem

	for _, g := range c.Collection.Games {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     g.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: g,
		})
	}

	options := gabagool.DefaultListOptions(c.Collection.DisplayName, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.ActionButton = constants.VirtualButtonX
	options.HelpButton = constants.VirtualButtonMenu
	options.HelpTitle = "Collection Management Controls"
	options.EmptyMessage = "This collection is empty.\nAdd some games you silly goose!"

	options.MultiSelectButton = constants.VirtualButtonSelect

	options.HelpText = []string{
		"• X: Open Options",
	}

	if len(menuItems) > 1 {
		options.ReorderButton = constants.VirtualButtonY
		options.HelpText = append(options.HelpText, "• Y: Toggle Reordering Mode")
		options.HelpText = append(options.HelpText, "• ↕: Move Selection")
	}

	options.HelpText = append(options.HelpText, "• Select: Toggle Multi-Select Removal")

	options.HelpText = append(options.HelpText, "• A: Remove ROM / Add to Remove Selection")
	options.HelpText = append(options.HelpText, "• Start: Remove Selected")
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Options"},
		{ButtonName: "Menu", HelpText: "Controls"},
	}

	selection, err := gabagool.List(options)

	if err != nil {
		if err == gabagool.ErrCancelled {
			// Save reordering before exit
			var games shared.Items
			for _, item := range selection.Items {
				games = append(games, item.Metadata.(shared.Item))
			}
			c.Collection.Games = games
			utils.SaveCollection(c.Collection)
			return c.Collection, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 && selection.Action == gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)

		var message string

		if len(selection.Selected) == 1 {
			message = fmt.Sprintf("Remove %s from %s?", selection.Items[selection.Selected[0]].Text, c.Collection.DisplayName)
		} else {
			message = fmt.Sprintf("Remove %d ROMs from %s?", len(selection.Selected), c.Collection.DisplayName)
		}

		if utils.ConfirmBulkAction(message) {
			var games shared.Items
			for _, item := range c.Collection.Games {
				isSelected := false
				for _, idx := range selection.Selected {
					if item.DisplayName == selection.Items[idx].Text {
						isSelected = true
						break
					}
				}
				if !isSelected {
					games = append(games, item)
				}
			}

			c.Collection.Games = games

			utils.SaveCollection(c.Collection)
		}

		return c.Collection, 0, nil
	} else {
		var games shared.Items
		for _, item := range selection.Items {
			games = append(games, item.Metadata.(shared.Item))
		}

		c.Collection.Games = games

		utils.SaveCollection(c.Collection)

		return c.Collection, 2, nil
	}
}
