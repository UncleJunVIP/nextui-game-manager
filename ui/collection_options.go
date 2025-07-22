package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type CollectionOptionsScreen struct {
	Collection   models.Collection
	SearchFilter string
}

func InitCollectionOptions(collection models.Collection, searchFilter string) CollectionOptionsScreen {
	return CollectionOptionsScreen{
		Collection:   collection,
		SearchFilter: searchFilter,
	}
}

func (c CollectionOptionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionOptions
}

func (c CollectionOptionsScreen) Draw() (screenReturn interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var actions []gabagool.MenuItem
	for _, action := range models.CollectionActionKeys {
		actions = append(actions, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(fmt.Sprintf("%s Options", c.Collection.DisplayName), actions)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)

	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		state.ClearCollectionMap()
		action := models.ActionMap[selection.Unwrap().SelectedItem.Metadata.(string)]

		switch action {
		case models.Actions.CollectionRename:
			newName, err := gabagool.Keyboard(c.Collection.DisplayName)
			if err != nil {
				return nil, -1, err
			}

			if newName.IsSome() {
				updatedCol, err := utils.RenameCollection(c.Collection, newName.Unwrap())
				if err != nil {
					logger.Error("failed to rename collection", zap.Error(err))
					return nil, -1, err
				}

				return updatedCol, 4, nil
			}

		case models.Actions.CollectionDelete:
			res, _ := gabagool.ConfirmationMessage(fmt.Sprintf("Are you sure you want to delete the collection\n%s?", c.Collection.DisplayName), []gabagool.FooterHelpItem{
				{ButtonName: "B", HelpText: "Cancel"},
				{ButtonName: "X", HelpText: "Delete"},
			}, gabagool.MessageOptions{
				ImagePath:     "",
				ConfirmButton: gabagool.ButtonX,
			})

			if res.IsSome() && !res.Unwrap().Cancelled {
				utils.DeleteCollection(c.Collection)
				return nil, 0, nil
			}

		case models.Actions.CollectionAlphabetize:
			confirm := utils.ConfirmAction(fmt.Sprintf("Alphabetize %s?\nThis cannot be undone!", c.Collection.DisplayName))

			if confirm {
				utils.AlphabetizeCollection(c.Collection)
			}
		}

		return c.Collection, 2, nil
	}

	return c.Collection, 2, nil
}
