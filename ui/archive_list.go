package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type ArchiveListScreen struct{}

func InitArchiveListScreen() ArchiveListScreen {
	return ArchiveListScreen{}
}

func (als ArchiveListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveList
}

// Lists available archive folders
func (als ArchiveListScreen) Draw() (item interface{}, exitCode int, e error) {
	title := "Archives"

	archiveFolders, err := utils.GetArchiveFileListBasic()
	if err != nil {
		utils.ShowTimedMessage("Unable to Load Archives!", time.Second*2)
		return nil, 404, nil
	}

	if archiveFolders == nil || len(archiveFolders) == 0 {
		return nil, 404, nil
	}

	var menuItems []gabagool.MenuItem
	for _, archiveFolder := range archiveFolders {
		archive := gabagool.MenuItem{
			Text:     archiveFolder,
			Selected: false,
			Focused:  false,
			Metadata: archiveFolder,
		}
		menuItems = append(menuItems, archive)
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
		selectedItem := selection.Items[selection.Selected[0]]
		archive := selectedItem.Metadata.(string)
		archiveDirectory := shared.RomDirectory{
			DisplayName: archive,
			Path:        utils.GetArchiveRoot(archive),
		}
		return archiveDirectory, 0, nil
	}

	return nil, 2, nil
}
