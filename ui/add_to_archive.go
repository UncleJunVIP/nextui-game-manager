package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"nextui-game-manager/shared"
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
	var archiveFolderEntries []gabagool.MenuItem
	for _, item := range archiveFolders {
		archiveFolderEntries = append(archiveFolderEntries, gabagool.MenuItem{
			Text:               item,
			Selected:           false,
			Focused:            false,
			Metadata:           item,
			NotMultiSelectable: true,
		})
	}

	options := gabagool.DefaultListOptions(title, archiveFolderEntries)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Archive Folders Found"
	options.ActionButton = constants.VirtualButtonX
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Create Archive"},
		{ButtonName: "A", HelpText: "Move"},
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
		archiveFolder := selection.Items[selection.Selected[0]].Text

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

	if len(selection.Selected) > 0 && selection.Action == gabagool.ListActionTriggered {
		return nil, 4, nil
	}

	return nil, 2, nil
}
