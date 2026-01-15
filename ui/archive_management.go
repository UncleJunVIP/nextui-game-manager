package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/filebrowser"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type ArchiveManagementScreen struct {
	Archive shared.RomDirectory
}

func InitArchiveManagementScreen(archive shared.RomDirectory) ArchiveManagementScreen {
	return ArchiveManagementScreen{
		Archive: archive,
	}
}

func (am ArchiveManagementScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveManagement
}

// Displays console folders in the selected archive folder and allows for archive deletion if all folders are empty
func (am ArchiveManagementScreen) Draw() (value interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := am.Archive.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(am.Archive.Path, false)
	if err != nil {
		logger.Info("Unable to fetch console directory! Continuing without them",
			zap.String("rom_directory", am.Archive.Path),
			zap.Error(err))
		return shared.Item{}, 1, err
	}

	var consoles []gabagool.MenuItem

	for _, item := range fb.Items {
		if !item.IsSelfContainedDirectory && !item.IsMultiDiscDirectory && item.IsDirectory {
			romDirectory := shared.RomDirectory{
				DisplayName: item.DisplayName,
				Tag:         item.Tag,
				Path:        item.Path,
			}
			menuItem := gabagool.MenuItem{
				Text:     romDirectory.DisplayName,
				Selected: false,
				Focused:  false,
				Metadata: romDirectory,
			}
			consoles = append(consoles, menuItem)
		}
	}

	options := gabagool.DefaultListOptions(title, consoles)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.ActionButton = constants.VirtualButtonX
	options.HelpButton = constants.VirtualButtonMenu
	options.HelpTitle = "Archive Management Controls"
	options.EmptyMessage = "This archive is empty."

	options.HelpText = []string{
		"• X: Open Options",
	}

	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Options"},
		{ButtonName: "Menu", HelpText: "Controls"},
	}

	selection, err := gabagool.List(options)

	if err != nil {
		if err == gabagool.ErrCancelled {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(selection.Selected) > 0 && selection.Action == gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		return selection.Items[selection.Selected[0]].Metadata.(shared.RomDirectory), 0, nil
	}

	return nil, 2, nil
}
