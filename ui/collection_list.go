package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"qlova.tech/sum"
)

type CollectionListScreen struct {
	SearchFilter string
}

func InitCollectionList(searchFilter string) CollectionListScreen {
	return CollectionListScreen{
		SearchFilter: searchFilter,
	}
}

func (c CollectionListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionsList
}

func (c CollectionListScreen) Draw() (collection interface{}, exitCode int, e error) {
	title := "Collections"

	collectionList, code, err := utils.GenerateCollectionList(c.SearchFilter, true)
	if code != 0 {
		return nil, code, err
	}

	if c.SearchFilter != "" {
		title = "[Search: \"" + c.SearchFilter + "\"]"
	}

	var menuItems []gabagool.MenuItem
	for _, collection := range collectionList {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     collection.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: collection,
		})
	}

	if len(collectionList) == 0 {
		title = "No Collections Found"
	}

	options := gabagool.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.ActionButton = constants.VirtualButtonX
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return selection.Items[selection.Selected[0]].Metadata.(models.Collection), 0, nil
	}

	return nil, 2, nil
}
