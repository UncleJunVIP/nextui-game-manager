package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type ArchiveOptionsScreen struct {
	Archive shared.RomDirectory
}

func InitArchiveOptionsScreen(archive shared.RomDirectory) ArchiveOptionsScreen {
	return ArchiveOptionsScreen{
		Archive: archive,
	}
}

func (aos ArchiveOptionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveOptions
}

func (aos ArchiveOptionsScreen) Draw() (screenReturn interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var actions []gabagool.MenuItem
	for _, action := range models.ArchiveActionKeys {
		actions = append(actions, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(fmt.Sprintf("%s Options", aos.Archive.DisplayName), actions)

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
		action := models.ActionMap[selectedItem.Metadata.(string)]

		switch action {
		case models.Actions.ArchiveRename:
			oldArchive := utils.CleanArchiveName(aos.Archive.DisplayName)
			res, err := gabagool.Keyboard(oldArchive, "")

			if err != nil {
				if err == gabagool.ErrCancelled {
					return nil, 4, nil
				}
				return nil, 1, err
			}

			if res != nil && res.Text != "" {
				newArchive := res.Text
				if newArchive != oldArchive {
					newArchive = utils.PrepArchiveName(newArchive)
					newArchivePath := utils.GetArchiveRoot(newArchive)

					err := utils.MoveFile(aos.Archive.Path, newArchivePath)

					if err != nil {
						logger.Error("Failed to rename archive", zap.Error(err))
						utils.ShowTimedMessage("Failed to rename archive", time.Second*2)
						return nil, 1, err
					}

					archiveDirectory := shared.RomDirectory{
						DisplayName: newArchive,
						Path:        newArchivePath,
					}

					return archiveDirectory, 4, nil
				}
			}

			return nil, 4, nil

		case models.Actions.ArchiveDelete:
			res, err := gabagool.ConfirmationMessage(fmt.Sprintf("Are you sure you want to delete the archive\n%s?", aos.Archive.DisplayName), []gabagool.FooterHelpItem{
				{ButtonName: "B", HelpText: "Cancel"},
				{ButtonName: "X", HelpText: "Delete"},
			}, gabagool.MessageOptions{
				ImagePath:     "",
				ConfirmButton: constants.VirtualButtonX,
			})

			if err == nil && res != nil && res.Confirmed {
				res, err := utils.DeleteArchive(aos.Archive)

				if err != nil {
					logger.Error("Failed to delete archive", zap.Error(err))
					utils.ShowTimedMessage("Failed to delete archive", time.Second*2)
					return nil, 1, err
				}

				if res != "" {
					utils.ShowTimedMessage(fmt.Sprintf("Cannot delete while file exists in archive\n%s", res), time.Second*2)
					return aos.Archive, 2, nil
				}

				return nil, 0, nil
			}

		}

		return aos.Archive, 2, nil
	}

	return aos.Archive, 2, nil
}
