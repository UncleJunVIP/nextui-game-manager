package ui

import (
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"path/filepath"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"go.uber.org/zap"
	"nextui-game-manager/common"
	"nextui-game-manager/filebrowser"
	"nextui-game-manager/shared"
	"qlova.tech/sum"
)

type GameList struct {
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitGamesList(romDirectory shared.RomDirectory, searchFilter string) GameList {
	return InitGamesListWithPreviousDirectory(romDirectory, shared.RomDirectory{}, searchFilter)
}

func InitGamesListWithPreviousDirectory(romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, searchFilter string) GameList {
	return GameList{
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (gl GameList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GamesList
}

func (gl GameList) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := gl.RomDirectory.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(gl.RomDirectory.Path, false)
	if err != nil {
		logger.Info("Unable to fetch ROM directory! Continuing without them",
			zap.String("rom_directory", gl.RomDirectory.Path),
			zap.Error(err))
		return shared.Item{}, 1, err
	}

	var roms shared.Items
	roms = fb.Items

	if gl.SearchFilter != "" {
		title = "[Search: \"" + gl.SearchFilter + "\"]"
		roms = utils.FilterList(roms, gl.SearchFilter)
	}

	var directoryEntries []gabagool.MenuItem
	var itemEntries []gabagool.MenuItem

	for _, item := range roms {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if item.IsMultiDiscDirectory || item.IsSelfContainedDirectory || !item.IsDirectory {
			imageFilename := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename)) + ".png"

			itemEntries = append(itemEntries, gabagool.MenuItem{
				Text:          itemName,
				Selected:      false,
				Focused:       false,
				Metadata:      item,
				ImageFilename: filepath.Join(gl.RomDirectory.Path, ".media", imageFilename),
			})
		} else {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, gabagool.MenuItem{
				Text:               itemName,
				Selected:           false,
				Focused:            false,
				Metadata:           item,
				NotMultiSelectable: true,
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
	}

	appState := state.GetAppState()

	if appState.Config.ShowArt {
		options.ShowImages = true
	}

	options.HelpButton = constants.VirtualButtonMenu
	options.HelpTitle = "ROMs List Controls"
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
		return nil, 4, nil
	} else if len(selection.Selected) > 0 && selection.Action != gabagool.ListActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Selected[0], selection.VisiblePosition)
		var selectedItems shared.Items
		for _, idx := range selection.Selected {
			selectedItems = append(selectedItems, selection.Items[idx].Metadata.(shared.Item))
		}
		return selectedItems, 0, nil
	}

	return nil, 2, nil
}
