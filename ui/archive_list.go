package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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
		archive := selection.Unwrap().SelectedItem.Metadata.(string)
		archiveDirectory := shared.RomDirectory{
			DisplayName: archive,
			Path:        utils.GetArchiveRoot(archive),
		}
		return archiveDirectory, 0, nil
	}

	return nil, 2, nil
}
