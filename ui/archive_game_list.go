package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
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

	var roms shared.Items

	for _, entry := range fb.Items {
		roms = append(roms, shared.Item{
			DisplayName:          entry.DisplayName,
			Filename:             entry.Filename,
			IsDirectory:          !entry.IsMultiDiscDirectory && entry.IsDirectory,
			IsMultiDiscDirectory: entry.IsMultiDiscDirectory,
			Path:                 entry.Path,
		})
	}

	if agl.SearchFilter != "" {
		title = "[Search: \"" + agl.SearchFilter + "\"]"
		roms = utils.FilterList(roms, agl.SearchFilter)
	}

	var directoryEntries []gaba.MenuItem
	var itemEntries []gaba.MenuItem

	for _, item := range roms {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if item.IsDirectory {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, gaba.MenuItem{
				Text:               itemName,
				Selected:           false,
				Focused:            false,
				Metadata:           item,
				NotMultiSelectable: true,
			})
		} else {
			itemEntries = append(itemEntries, gaba.MenuItem{
				Text:     itemName,
				Selected: false,
				Focused:  false,
				Metadata: item,
			})
		}
	}

	allEntries := append(directoryEntries, itemEntries...)

	options := gaba.DefaultListOptions(title, allEntries)
	options.SmallTitle = true
	options.EmptyMessage = "No ROMs Found"
	options.EnableAction = true
	options.EnableMultiSelect = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Restore"},
	}

	options.EnableHelp = true
	options.HelpTitle = "Archive ROMs List Controls"
	options.HelpText = []string{
		"• X: Open Options",
		"• Select: Toggle Multi-Select",
		"• Start: Confirm Multi-Selection",
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		query, err := gaba.Keyboard("")

		if err != nil {
			return nil, 1, err
		}

		if query.IsSome() {
			return query.Unwrap(), 4, nil
		}

		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		rawSelection := selection.Unwrap().SelectedItems

		firstItem := rawSelection[0].Metadata.(shared.Item)

		confirmMessage := fmt.Sprintf("Restore %s from archive %s?", firstItem.DisplayName, agl.Archive.DisplayName)
		successMessage := fmt.Sprintf("Restored %s from archive %s!", firstItem.DisplayName, agl.Archive.DisplayName)
		if len(rawSelection) > 1 {
			confirmMessage = fmt.Sprintf("Restore %d from archive %s?", len(rawSelection), agl.Archive.DisplayName)
			successMessage = fmt.Sprintf("Restored %d games from archive %s!", len(rawSelection), agl.Archive.DisplayName)
		} else {
			if firstItem.IsDirectory {
				newRomDirectory := shared.RomDirectory{
					DisplayName: firstItem.DisplayName,
					Tag:         firstItem.Tag,
					Path:        firstItem.Path,
				}
				return newRomDirectory, 0, nil
			}
		}

		result, err := gaba.ConfirmationMessage(confirmMessage, []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "I Changed My Mind"},
			{ButtonName: "A", HelpText: "Yes"},
		}, gaba.MessageOptions{})

		if err != nil || !result.IsSome() {
			return agl.SearchFilter, 4, err
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
