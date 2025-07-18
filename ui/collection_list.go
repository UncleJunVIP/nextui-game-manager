package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
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

	var menuItems []gaba.MenuItem
	for _, collection := range collectionList {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     collection.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: collection,
		})
	}

	if len(collectionList) == 0 {
		title = "No Collections Found"
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Metadata.(models.Collection), 0, nil
	}

	return nil, 2, nil
}
