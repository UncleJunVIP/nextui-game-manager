package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type AddToArchiveScreen struct {
	Games                []shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitAddToArchiveScreen(gamesList []shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) AddToArchiveScreen {
	return AddToArchiveScreen{
		Games:                gamesList,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (atas AddToArchiveScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.AddToArchive
}

// Adds selected rom(s) to an archive option. New archives can be created through the action button
func (atas AddToArchiveScreen) Draw() (item interface{}, exitCode int, e error) {
	bulk := len(atas.Games) > 1

	title := fmt.Sprintf("Move %s To Archive", atas.Games[0].DisplayName)
	if bulk {
		title = fmt.Sprintf("Move %d Games To Archive", len(atas.Games))
	}

	archiveFolders, err := utils.GetArchiveFileList()
	if err != nil {
		utils.ShowTimedMessage("Unable to Load Archives!", time.Second*2)
		return nil, -1, nil
	}
	var archiveFolderEntries []gaba.MenuItem
	for _, item := range archiveFolders {
		archiveFolderEntries = append(archiveFolderEntries, gaba.MenuItem{
			Text:               item,
			Selected:           false,
			Focused:            false,
			Metadata:           item,
			NotMultiSelectable: true,
		})
	}

	options := gaba.DefaultListOptions(title, archiveFolderEntries)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Archive Folders Found"
	options.EnableAction = true
	options.EnableMultiSelect = false
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Create Archive"},
		{ButtonName: "A", HelpText: "Move"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		archiveFolder := selection.Unwrap().SelectedItem.Text

		message := fmt.Sprintf("Archive %s into %s?", atas.Games[0].DisplayName, archiveFolder)
		if bulk {
			message = fmt.Sprintf("Archive %d games into %s?", len(atas.Games), archiveFolder)
		}

		if !utils.ConfirmAction(message) {
			return nil, 404, nil
		}

		for _, game := range atas.Games {
			if err := utils.ArchiveRom(game, atas.RomDirectory, archiveFolder); err != nil {
				utils.ShowTimedMessage(fmt.Sprintf("Unable to archive %s!", game.DisplayName), time.Second*3)
				return nil, 404, err
			}
		}

		successMessage := fmt.Sprintf("Added %s To Archive %s!", atas.Games[0].DisplayName, archiveFolder)
		if bulk {
			successMessage = fmt.Sprintf("Added %d Games To Archive %s!", len(atas.Games), archiveFolder)
		}

		utils.ShowTimedMessage(successMessage, time.Second*2)

		return nil, 0, nil
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	}

	return nil, 2, nil
}
