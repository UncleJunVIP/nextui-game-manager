package ui

import (
	"fmt"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"path/filepath"
	"strings"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/filebrowser"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type ArchiveGamesListScreen struct {
	Archive              shared.RomDirectory
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitArchiveGamesListScreen(archive shared.RomDirectory, romDirectory shared.RomDirectory, searchFilter string) ArchiveGamesListScreen {
	return InitArchiveGamesListScreenWithPreviousDirectory(archive, romDirectory, shared.RomDirectory{}, searchFilter)
}

func InitArchiveGamesListScreenWithPreviousDirectory(archive shared.RomDirectory, romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, searchFilter string) ArchiveGamesListScreen {
	return ArchiveGamesListScreen{
		Archive:              archive,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (agl ArchiveGamesListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveGamesList
}

func (agl ArchiveGamesListScreen) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := agl.Archive.DisplayName + " : " + agl.RomDirectory.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(agl.RomDirectory.Path, false)
	if err != nil {
		logger.Info("Unable to fetch ROM directory! Continuing without them",
			zap.String("rom_directory", agl.RomDirectory.Path),
			zap.Error(err))
		return shared.Item{}, 1, err
	}

	roms := fb.Items

	if agl.SearchFilter != "" {
		title = "[Search: \"" + agl.SearchFilter + "\"]"
		roms = utils.FilterList(roms, agl.SearchFilter)
	}

	var directoryEntries []gabagool.MenuItem
	var itemEntries []gabagool.MenuItem

	for _, item := range roms {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if !item.IsSelfContainedDirectory && !item.IsMultiDiscDirectory && item.IsDirectory {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, gabagool.MenuItem{
				Text:               itemName,
				Selected:           false,
				Focused:            false,
				Metadata:           item,
				NotMultiSelectable: true,
			})
		} else {
			itemEntries = append(itemEntries, gabagool.MenuItem{
				Text:     itemName,
				Selected: false,
				Focused:  false,
				Metadata: item,
			})
		}
	}

	allEntries := append(directoryEntries, itemEntries...)

	options := gabagool.DefaultListOptions(title, allEntries)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.UseSmallTitle = true
	options.EmptyMessage = "No ROMs Found"
	options.ActionButton = constants.VirtualButtonX
	options.MultiSelectButton = constants.VirtualButtonSelect
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Restore"},
	}

	options.HelpButton = constants.VirtualButtonMenu
	options.HelpTitle = "Archive ROMs List Controls"
	options.HelpText = []string{
		"• X: Open Options",
		"• Select: Toggle Multi-Select",
		"• Start: Confirm Multi-Selection",
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
		query, err := gabagool.Keyboard("", "")

		if err != nil {
			if err == gabagool.ErrCancelled {
				return nil, 4, nil
			}
			return nil, 1, err
		}

		if query != nil && query.Text != "" {
			return query.Text, 4, nil
		}

		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		// Map selected indices to menu items
		var rawSelection []gabagool.MenuItem
		for _, idx := range selection.Selected {
			rawSelection = append(rawSelection, selection.Items[idx])
		}

		firstItem := rawSelection[0].Metadata.(shared.Item)

		confirmMessage := fmt.Sprintf("Restore %s from archive %s?", firstItem.DisplayName, agl.Archive.DisplayName)
		successMessage := fmt.Sprintf("Restored %s from archive %s!", firstItem.DisplayName, agl.Archive.DisplayName)
		if len(rawSelection) > 1 {
			confirmMessage = fmt.Sprintf("Restore %d from archive %s?", len(rawSelection), agl.Archive.DisplayName)
			successMessage = fmt.Sprintf("Restored %d games from archive %s!", len(rawSelection), agl.Archive.DisplayName)
		} else {
			if !firstItem.IsSelfContainedDirectory && !firstItem.IsMultiDiscDirectory && firstItem.IsDirectory {
				newRomDirectory := shared.RomDirectory{
					DisplayName: firstItem.DisplayName,
					Tag:         firstItem.Tag,
					Path:        firstItem.Path,
				}
				return newRomDirectory, 0, nil
			}
		}

		if !utils.ConfirmAction(confirmMessage) {
			return agl.SearchFilter, 4, nil
		}

		for _, selection := range rawSelection {
			item := selection.Metadata.(shared.Item)
			err := utils.RestoreRom(item, agl.RomDirectory, agl.Archive)
			if err != nil {
				utils.ShowTimedMessage(fmt.Sprintf("Unable to restore %s!", item.DisplayName), time.Second*2)
				return shared.RomDirectory{}, 0, err
			}
		}

		utils.ShowTimedMessage(successMessage, time.Second*2)

		return shared.RomDirectory{}, 0, nil
	}

	return nil, 2, nil
}
