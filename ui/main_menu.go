package ui

import (
	"log/slog"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

const (
	collectionsDisplayName = "Collections"
	collectionsTag         = "Collections"
	archivesDisplayName    = "Archives"
	archivesTag            = "Archives"
	settingsExitCode       = 4
	selectExitCode         = 0
	ToolsExitCode          = 5
	quitExitCode           = 2
	errorExitCode          = -1
)

type MainMenu struct{}

func InitMainMenu() MainMenu {
	return MainMenu{}
}

func (m MainMenu) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.MainMenu
}

func (m MainMenu) Draw() (interface{}, int, error) {
	logger := common.GetLoggerInstance()

	menuItems, err := buildMenuItems(logger)
	if err != nil {
		return nil, errorExitCode, err
	}

	return handleMenuSelection(menuItems)
}

func buildMenuItems(logger *slog.Logger) ([]gaba.MenuItem, error) {
	var menuItems []gaba.MenuItem

	if collectionsItem := buildCollectionsMenuItem(logger); collectionsItem != nil {
		menuItems = append(menuItems, *collectionsItem)
	}

	if archivesItem := buildArchivesMenuItem(); archivesItem != nil {
		menuItems = append(menuItems, *archivesItem)
	}

	romItems, err := buildRomDirectoryMenuItems(logger)
	if err != nil {
		return nil, err
	}

	menuItems = append(menuItems, romItems...)

	menuItems = append(menuItems, gaba.MenuItem{
		Text:     "Tools",
		Selected: false,
		Focused:  false,
		Metadata: "Tools",
	})

	return menuItems, nil
}

func buildCollectionsMenuItem(logger *slog.Logger) *gaba.MenuItem {
	fb := filebrowser.NewFileBrowser(logger)

	if err := fb.CWD(utils.GetCollectionDirectory(), false); err != nil {
		logger.Info("Unable to fetch collection directories, skipping", "error", err)
		return nil
	}

	if len(fb.Items) == 0 {
		return nil
	}

	collections := createCollectionsRomDirectory()
	return &gaba.MenuItem{
		Text:     collectionsDisplayName,
		Selected: false,
		Focused:  false,
		Metadata: collections,
	}
}

func createCollectionsRomDirectory() shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: collectionsDisplayName,
		Tag:         collectionsTag,
		Path:        common.CollectionDirectory,
	}
}

func buildArchivesMenuItem() *gaba.MenuItem {
	archiveFolders, err := utils.GetArchiveFileListBasic()

	if err != nil {
		return nil
	}

	if archiveFolders == nil {
		return nil
	}

	archives := createArchivesRomDirectory()
	return &gaba.MenuItem{
		Text:     archivesDisplayName,
		Selected: false,
		Focused:  false,
		Metadata: archives,
	}
}

func createArchivesRomDirectory() shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: archivesDisplayName,
		Tag:         archivesTag,
		Path:        utils.GetRomDirectory(),
	}
}

func buildRomDirectoryMenuItems(logger *slog.Logger) ([]gaba.MenuItem, error) {
	fb := filebrowser.NewFileBrowser(logger)

	if err := fb.CWD(utils.GetRomDirectory(), state.GetAppState().Config.HideEmpty); err != nil {
		showRomDirectoryError()
		common.LogStandardFatal("Error fetching ROM directories", err)
		return nil, err
	}

	var menuItems []gaba.MenuItem
	for _, item := range fb.Items {
		if item.IsDirectory {

			if item.Tag == "(PORTS)" {
				continue
			}

			romDirectory := utils.CreateRomDirectoryFromItem(item)
			menuItem := createMenuItemFromRomDirectory(romDirectory)
			menuItems = append(menuItems, menuItem)
		}
	}

	return menuItems, nil
}

func createMenuItemFromRomDirectory(romDirectory shared.RomDirectory) gaba.MenuItem {
	return gaba.MenuItem{
		Text:     romDirectory.DisplayName,
		Selected: false,
		Focused:  false,
		Metadata: romDirectory,
	}
}

func showRomDirectoryError() {
	gaba.ConfirmationMessage("Unable to fetch ROM directories!", []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
	}, gaba.MessageOptions{})
}

func handleMenuSelection(menuItems []gaba.MenuItem) (interface{}, int, error) {
	options := createListOptions(menuItems)
	selection, err := gaba.List(options)

	if err != nil {
		return nil, errorExitCode, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return nil, settingsExitCode, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		if selection.Unwrap().SelectedItem.Metadata == "Tools" {
			return nil, ToolsExitCode, nil
		}

		return selection.Unwrap().SelectedItem.Metadata.(shared.RomDirectory), selectExitCode, nil
	}

	return nil, quitExitCode, nil
}

func createListOptions(menuItems []gaba.MenuItem) gaba.ListOptions {
	options := gaba.DefaultListOptions("Game Manager", menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "X", HelpText: "Settings"},
		{ButtonName: "A", HelpText: "Select"},
	}
	return options
}
