package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type AddToCollectionScreen struct {
	Games                []shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitAddToCollectionScreen(games []shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) AddToCollectionScreen {
	return AddToCollectionScreen{
		Games:                games,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (a AddToCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.AddToCollection
}

func (a AddToCollectionScreen) Draw() (collection interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	fb := filebrowser.NewFileBrowser(logger)
	err := fb.CWD(utils.GetCollectionDirectory(), false)
	if err != nil {
		utils.ShowTimedMessage("Unable to Load Collections!", time.Second*2)
		return nil, -1, nil
	}

	if len(fb.Items) == 0 {
		if !utils.ConfirmAction("No Collections Found. \n Want to create your first?") {
			return nil, 2, nil
		}

		return nil, 404, nil
	}

	var collections []models.Collection
	collectionsMap := make(map[string]models.Collection)

	for _, item := range fb.Items {
		collection := models.Collection{
			DisplayName:    item.DisplayName,
			CollectionFile: item.Path,
		}

		var err error
		collection, err = utils.ReadCollection(collection)
		if err != nil {
			logger.Error("Error reading collection", zap.Error(err))
			continue
		}

		membershipCount := 0

		for _, game := range a.Games {
			if utils.GameExistsInCollection(collection.Games, game) {
				membershipCount++
			}
		}

		if len(a.Games) > membershipCount {
			collections = append(collections, collection)
			collectionsMap[item.DisplayName] = collection
		}
	}

	if len(a.Games) == 1 && len(collections) == 0 {
		res, err := gaba.ConfirmationMessage(fmt.Sprintf("Every collection contains %s. \n Want to create a new one?", a.Games[0].DisplayName),
			[]gaba.FooterHelpItem{{
				HelpText:   "No Thanks",
				ButtonName: "B",
			}, {
				HelpText:   "Yes",
				ButtonName: "A",
			}}, gaba.MessageOptions{})

		if err != nil || res.IsNone() {
			return nil, 2, nil
		}

		return nil, 404, nil
	}

	var itemList []shared.Item

	for _, collection := range collections {
		itemList = append(itemList, shared.Item{
			DisplayName: collection.DisplayName,
			Path:        collection.CollectionFile,
		})
	}

	slices.SortFunc(itemList, func(a, b shared.Item) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	var menuItems []gaba.MenuItem
	for _, item := range itemList {
		col := models.Collection{DisplayName: item.DisplayName, CollectionFile: item.Path}
		col, err = utils.ReadCollection(col)

		if err != nil {
			utils.ShowTimedMessage("Unable to Load Collections!", time.Second*2)
			return nil, -1, err
		}

		collection := gaba.MenuItem{
			Text:     item.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: col,
		}
		menuItems = append(menuItems, collection)
	}

	title := fmt.Sprintf("Add %s To Collection", a.Games[0].DisplayName)
	if len(a.Games) > 1 {
		title = "Add Games To Collection"
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EnableAction = true
	options.EnableMultiSelect = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Create Collection"},
		{ButtonName: "A", HelpText: "Add"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		state.ClearCollectionMap()
		selectedCol := selection.Unwrap().SelectedItem.Metadata.(models.Collection)

		_, err := utils.AddCollectionGames(selectedCol, a.Games)

		if err != nil {
			gameText := a.Games[0].DisplayName

			if len(a.Games) > 1 {
				gameText = fmt.Sprintf("%d Games", len(a.Games))
			}

			utils.ShowTimedMessage(fmt.Sprintf("Unable to Add %s To Collection!", gameText), time.Second*2)
			return nil, 0, err
		}

		successMessage := fmt.Sprintf("Added %s To Collection %s!", a.Games[0].DisplayName, selectedCol.DisplayName)

		if len(a.Games) > 1 {
			successMessage = fmt.Sprintf("Added %d Games To Collection %s!", len(a.Games), selectedCol.DisplayName)
		}

		utils.ShowTimedMessage(successMessage, time.Second*2)

		return nil, 0, nil
	} else if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	}

	return nil, 2, nil
}
