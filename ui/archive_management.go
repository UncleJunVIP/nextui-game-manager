package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
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

	var consoles []gaba.MenuItem

	for _, item := range fb.Items {
		if !item.IsSelfContainedDirectory && !item.IsMultiDiscDirectory && item.IsDirectory {
			romDirectory := shared.RomDirectory{
				DisplayName: item.DisplayName,
				Tag:         item.Tag,
				Path:        item.Path,
			}
			menuItem := gaba.MenuItem{
				Text:     romDirectory.DisplayName,
				Selected: false,
				Focused:  false,
				Metadata: romDirectory,
			}
			consoles = append(consoles, menuItem)
		}
	}

	options := gaba.DefaultListOptions(title, consoles)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.EnableHelp = true
	options.HelpTitle = "Archive Management Controls"
	options.EmptyMessage = "This archive is empty."

	options.HelpText = []string{
		"• X: Open Options",
	}

	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Options"},
		{ButtonName: "Menu", HelpText: "Controls"},
	}

	selection, err := gaba.List(options)

	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Metadata.(shared.RomDirectory), 0, nil
	}

	return nil, 2, nil
}
