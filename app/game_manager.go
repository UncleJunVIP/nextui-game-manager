package main

import (
	"fmt"
	_ "github.com/UncleJunVIP/certifiable"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"log"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/ui"
	"nextui-game-manager/utils"
	"os"
	"qlova.tech/sum"
	"time"
)

const (
	defaultLogLevel      = "ERROR"
	defaultDirPerm       = 0755
	shortMessageDelay    = 1250 * time.Millisecond
	standardMessageDelay = 2 * time.Second
	longMessageDelay     = 3 * time.Second
)

const (
	ExitCodeSuccess       = 0   // Success, proceed with result
	ExitCodeError         = 1   // Generic error
	ExitCodeCancel        = 2   // User cancelled/back
	ExitCodeAction        = 4   // Action button pressed (X, Settings, etc.)
	ExitCodeEmpty         = 404 // No results/empty state
	ExitCodeInternalError = -1  // Internal error
)

func init() {
	gaba.InitSDL(gaba.GabagoolOptions{
		WindowTitle:    "Game Manager",
		ShowBackground: true,
	})

	common.SetLogLevel(defaultLogLevel)
	common.InitIncludes()

	config, err := loadConfig()
	if err != nil {
		log.Fatal("Unable to initialize configuration", zap.Error(err))
	}

	common.SetLogLevel(config.LogLevel)
	state.SetConfig(config)

	logger := common.GetLoggerInstance()
	logger.Debug("Configuration loaded", zap.Object("config", config))

	collectionDir := utils.GetCollectionDirectory()
	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(collectionDir, defaultDirPerm); mkdirErr != nil {
			gaba.ConfirmationMessage("Unable to create Collections directory!", []gaba.FooterHelpItem{
				{ButtonName: "B", HelpText: "Quit"},
			}, gaba.MessageOptions{})
			log.Fatal("Unable to create collection directory", zap.Error(mkdirErr))
		}
	}
}

func loadConfig() (*models.Config, error) {
	config, err := state.LoadConfig()
	if err != nil {
		config = &models.Config{
			ArtDownloadType: shared.ArtDownloadTypeFromString["BOX_ART"],
			HideEmpty:       false,
			LogLevel:        defaultLogLevel,
		}
		if saveErr := utils.SaveConfig(config); saveErr != nil {
			return nil, fmt.Errorf("failed to save default config: %w", saveErr)
		}
	}
	return config, nil
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()
	logger.Info("Starting Game Manager")

	runApplicationLoop()
}

func cleanup() {
	gaba.CloseSDL()
	common.CloseLogger()
}

func runApplicationLoop() {
	var screen models.Screen
	screen = ui.InitMainMenu()
	state.AddNewMenuPosition()

	for {
		result, code, _ := screen.Draw() // TODO: Implement proper error handling
		screen = handleScreenTransition(screen, result, code)
	}
}

func handleScreenTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	switch currentScreen.Name() {
	case models.ScreenNames.MainMenu:
		return handleMainMenuTransition(result, code)
	case models.ScreenNames.Settings:
		state.ReturnToMain()
		return ui.InitMainMenu()
	case models.ScreenNames.CollectionsList:
		return handleCollectionsListTransition(result, code)
	case models.ScreenNames.CollectionManagement:
		return handleCollectionManagementTransition(currentScreen, result, code)
	case models.ScreenNames.CollectionOptions:
		return handleCollectionOptionsTransition(currentScreen, result, code)
	case models.ScreenNames.Tools:
		return handleToolsTransition(result, code)
	case models.ScreenNames.GlobalActions:
		return handleGlobalActionsTransition(code)
	case models.ScreenNames.GamesList:
		return handleGamesListTransition(currentScreen, result, code)
	case models.ScreenNames.SearchBox:
		return handleSearchBoxTransition(currentScreen, result, code)
	case models.ScreenNames.Actions:
		return handleActionsTransition(currentScreen, result, code)
	case models.ScreenNames.BulkActions:
		return handleBulkActionsTransition(currentScreen, result, code)
	case models.ScreenNames.AddToCollection:
		return handleAddToCollectionTransition(currentScreen, code)
	case models.ScreenNames.CollectionCreate:
		return handleCollectionCreateTransition(currentScreen)
	case models.ScreenNames.DownloadArt:
		return handleDownloadArtTransition(currentScreen)
	case models.ScreenNames.AddToArchive:
		return handleAddToArchiveTransition(currentScreen, result, code)
	case models.ScreenNames.ArchiveCreate:
		return handleArchiveCreateTransition(currentScreen, result, code)
	case models.ScreenNames.ArchiveList:
		return handleArchiveListTransition(result, code)
	case models.ScreenNames.ArchiveGamesList:
		return handleArchiveGamesListTransition(currentScreen, result, code)
	case models.ScreenNames.ArchiveManagement:
		return handleArchiveManagementTransition(currentScreen, result, code)
	case models.ScreenNames.ArchiveOptions:
		return handleArchiveOptionsTransition(currentScreen, result, code)
	case models.ScreenNames.PlayHistoryList:
		return handlePlayHistoryListTransition(currentScreen, result, code)
	case models.ScreenNames.PlayHistoryGameList:
		return handlePlayHistoryGameListTransition(currentScreen, result, code)
	case models.ScreenNames.PlayHistoryGameDetails:
		return handlePlayHistoryGameDetailsTransition(currentScreen, result, code)
	case models.ScreenNames.PlayHistoryGameHistory:
		return handlePlayHistoryGameHistoryTransition(currentScreen, result, code)
	case models.ScreenNames.PlayHistoryFilter:
		return handlePlayHistoryFilterTransition(currentScreen, result, code)
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handlePlayHistoryFilterTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	phfs := currentScreen.(ui.PlayHistoryFilterScreen)
	switch code {
	case ExitCodeSuccess:
		newFilter := result.(models.PlayHistorySearchFilter)
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryFilterScreen(phfs.Console, phfs.SearchFilter, phfs.GameAggregate, 
			phfs.Game, phfs.RomDirectory, phfs.PreviousRomDirectory, phfs.PlayHistoryOrigin, 
			append(phfs.PlayHistoryFilterList, newFilter), phfs.MenuDepth+1)
	case ExitCodeCancel:
		if len(phfs.PlayHistoryFilterList) > 0 {
			if phfs.MenuDepth > 1 {
				state.RemoveMenuPositions(1)
				return ui.InitPlayHistoryFilterScreen(phfs.Console, phfs.SearchFilter, phfs.GameAggregate, 
					phfs.Game, phfs.RomDirectory, phfs.PreviousRomDirectory, phfs.PlayHistoryOrigin, 
					phfs.PlayHistoryFilterList[:len(phfs.PlayHistoryFilterList)-1], phfs.MenuDepth-1)
			}
			state.UpdateCurrentMenuPosition(0, 0)
			return ui.InitPlayHistoryFilterScreen(phfs.Console, phfs.SearchFilter, phfs.GameAggregate, 
				phfs.Game, phfs.RomDirectory, phfs.PreviousRomDirectory, phfs.PlayHistoryOrigin, 
				phfs.PlayHistoryFilterList[:len(phfs.PlayHistoryFilterList)-1], phfs.MenuDepth)
		}
		return ui.InitPlayHistoryFilterScreen(phfs.Console, phfs.SearchFilter, phfs.GameAggregate, 
			phfs.Game, phfs.RomDirectory, phfs.PreviousRomDirectory, phfs.PlayHistoryOrigin, 
			phfs.PlayHistoryFilterList, phfs.MenuDepth)
	}
	state.RemoveMenuPositions(phfs.MenuDepth)
	state.UpdateCurrentMenuPosition(0, 0)
	if phfs.Console == "" {
		return ui.InitPlayHistoryListScreen(phfs.PlayHistoryFilterList)
	}
	if phfs.GameAggregate.Name == "" {
		return ui.InitPlayHistoryGamesListScreen(phfs.Console, phfs.PlayHistoryFilterList)
	}
	return ui.InitPlayHistoryGameHistoryScreen(phfs.Console, phfs.SearchFilter, phfs.GameAggregate, 
		phfs.Game, phfs.RomDirectory, phfs.PreviousRomDirectory, phfs.PlayHistoryOrigin, phfs.PlayHistoryFilterList)
}

func handlePlayHistoryListTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	phls := currentScreen.(ui.PlayHistoryListScreen)
	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryGamesListScreen(result.(string), phls.PlayHistoryFilterList)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryFilterScreenFromHistoryList(phls.PlayHistoryFilterList)
	default:
		state.RemoveMenuPositions(1)
		return ui.InitToolsScreen()
	}
}

func handlePlayHistoryGameListTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	ptgls := currentScreen.(ui.PlayHistoryGamesListScreen)

	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryGameDetailsScreenFromPlayHistory(ptgls.Console, result.(models.PlayHistoryAggregate), ptgls.PlayHistoryFilterList)
	case ExitCodeCancel:
		state.RemoveMenuPositions(1)
		return ui.InitPlayHistoryListScreen(ptgls.PlayHistoryFilterList)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryFilterScreenFromGameList(ptgls.Console, ptgls.PlayHistoryFilterList)
	case ExitCodeEmpty:
		utils.ShowTimedMessage(fmt.Sprintf("%s history is empty!", ptgls.Console), longMessageDelay)
		state.RemoveMenuPositions(1)
		return ui.InitPlayHistoryListScreen(ptgls.PlayHistoryFilterList)
	default:
		state.RemoveMenuPositions(1)
		return ui.InitPlayHistoryListScreen(ptgls.PlayHistoryFilterList)
	}
}

func handlePlayHistoryGameHistoryTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	ptghs := currentScreen.(ui.PlayHistoryGameHistoryScreen)

	switch code {
	case ExitCodeSuccess:
		return ui.InitPlayHistoryGameHistoryScreen(ptghs.Console, ptghs.SearchFilter, ptghs.GameAggregate, 
			ptghs.Game, ptghs.RomDirectory, ptghs.PreviousRomDirectory, ptghs.PlayHistoryOrigin, ptghs.PlayHistoryFilterList)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryFilterScreen(ptghs.Console, ptghs.SearchFilter, ptghs.GameAggregate, 
			ptghs.Game, ptghs.RomDirectory, ptghs.PreviousRomDirectory, ptghs.PlayHistoryOrigin, ptghs.PlayHistoryFilterList, 1)
	default:
		return ui.InitPlayHistoryGameDetailsScreenFromSelf(ptghs.Console, ptghs.SearchFilter, ptghs.GameAggregate, 
			ptghs.Game, ptghs.RomDirectory, ptghs.PreviousRomDirectory, ptghs.PlayHistoryOrigin, ptghs.PlayHistoryFilterList)
	}
}

func handlePlayHistoryGameDetailsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	ptgds := currentScreen.(ui.PlayHistoryGameDetailsScreen)
	switch code {
	case ExitCodeSuccess:
		return ui.InitPlayHistoryGameHistoryScreen(ptgds.Console, ptgds.SearchFilter, ptgds.GameAggregate, 
		ptgds.Game, ptgds.RomDirectory, ptgds.PreviousRomDirectory, ptgds.PlayHistoryOrigin, ptgds.PlayHistoryFilterList)
	default:
		state.RemoveMenuPositions(1)
		if ptgds.PlayHistoryOrigin {
			return ui.InitPlayHistoryGamesListScreen(ptgds.Console, ptgds.PlayHistoryFilterList)
		}
		return ui.InitActionsScreen(ptgds.Game, ptgds.RomDirectory, ptgds.PreviousRomDirectory, ptgds.SearchFilter)
	}
}

func handleMainMenuTransition(result interface{}, code int) models.Screen {
	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		romDir := result.(shared.RomDirectory)
		if romDir.DisplayName == "Collections" {
			return ui.InitCollectionList("")
		}
		if romDir.DisplayName == "Archives" {
			return ui.InitArchiveListScreen()
		}
		return ui.InitGamesList(romDir, "")
	case ExitCodeError, ExitCodeCancel:
		os.Exit(0)
		return nil
	case ExitCodeAction:
		return ui.InitSettingsScreen()
	case ui.ToolsExitCode:
		state.AddNewMenuPosition()
		return ui.InitToolsScreen()
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleCollectionsListTransition(result interface{}, code int) models.Screen {
	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		collection := result.(models.Collection)
		return ui.InitCollectionManagement(collection)
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleArchiveManagementTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	ams := currentScreen.(ui.ArchiveManagementScreen)

	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		return ui.InitArchiveGamesListScreen(ams.Archive, result.(shared.RomDirectory), "")
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitArchiveOptionsScreen(ams.Archive)
	default:
		state.RemoveMenuPositions(1)
		return ui.InitArchiveListScreen()
	}
}

func handleArchiveGamesListTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	agl := currentScreen.(ui.ArchiveGamesListScreen)

	switch code {
	case ExitCodeSuccess:
		newRomDirectory := result.(shared.RomDirectory)

		if newRomDirectory.Path != "" {
			state.AddNewMenuPosition()
			return ui.InitArchiveGamesListScreenWithPreviousDirectory(agl.Archive, newRomDirectory, agl.RomDirectory, "")
		}

		state.RemoveMenuPositions(1)
		return ui.InitArchiveGamesListScreen(agl.Archive, agl.RomDirectory, "")
	case ExitCodeCancel:
		if agl.PreviousRomDirectory.Path != "" {
			state.RemoveMenuPositions(1)
			return ui.InitArchiveGamesListScreen(agl.Archive, agl.PreviousRomDirectory, "")
		}

		if agl.SearchFilter != "" {
			state.UpdateCurrentMenuPosition(0, 0)
			return ui.InitArchiveGamesListScreen(agl.Archive, agl.RomDirectory, "")
		}

		state.ReturnToArchiveManagement()
		return ui.InitArchiveManagementScreen(agl.Archive)
	case ExitCodeAction, ExitCodeError:
		searchFilter := result.(string)

		if searchFilter != "" {
			state.UpdateCurrentMenuPosition(0, 0)
			return ui.InitArchiveGamesListScreen(agl.Archive, agl.RomDirectory, searchFilter)
		}

		return ui.InitArchiveGamesListScreen(agl.Archive, agl.RomDirectory, "")
	case ExitCodeEmpty:
		if agl.SearchFilter != "" {
			utils.ShowTimedMessage(fmt.Sprintf("No results found for %s!", agl.SearchFilter), shortMessageDelay)
			return ui.InitArchiveGamesListScreen(agl.Archive, agl.RomDirectory, "")
		}

		utils.ShowTimedMessage(fmt.Sprintf("%s is empty!", agl.RomDirectory.DisplayName), longMessageDelay)
		state.ReturnToArchiveManagement()
		return ui.InitArchiveManagementScreen(agl.Archive)
	default:
		state.ReturnToArchiveManagement()
		return ui.InitArchiveManagementScreen(agl.Archive)
	}
}

func handleArchiveOptionsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	aos := currentScreen.(ui.ArchiveOptionsScreen)

	switch code {
	case ExitCodeError, ExitCodeAction:
		if result != nil {
			// TODO: Update position around renamed archive
			state.UpdateCurrentMenuPosition(0, 0)
			return ui.InitArchiveOptionsScreen(result.(shared.RomDirectory))
		}
		return ui.InitArchiveOptionsScreen(aos.Archive)
	case ExitCodeSuccess:
		state.RemoveMenuPositions(2)
		return ui.InitArchiveListScreen()
	default:
		state.ReturnToArchiveManagement()
		return ui.InitArchiveManagementScreen(aos.Archive)
	}
}

func handleArchiveListTransition(result interface{}, code int) models.Screen {
	switch code {
	case ExitCodeSuccess:
		state.AddNewMenuPosition()
		return ui.InitArchiveManagementScreen(result.(shared.RomDirectory))
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleCollectionManagementTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	cm := currentScreen.(ui.CollectionManagement)

	switch code {
	case ExitCodeSuccess:
		collection := result.(models.Collection)
		//TODO: Update position around deleted game
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitCollectionManagement(collection)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitCollectionOptions(cm.Collection, cm.SearchFilter)
	case ExitCodeInternalError, ExitCodeCancel:
		state.RemoveMenuPositions(1)
		return ui.InitCollectionList(cm.SearchFilter)
	default:
		state.RemoveMenuPositions(1)
		return ui.InitCollectionList(cm.SearchFilter)
	}
}

func handleCollectionOptionsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	co := currentScreen.(ui.CollectionOptionsScreen)

	switch code {
	case ExitCodeSuccess:
		state.RemoveMenuPositions(2)
		return ui.InitCollectionList(co.SearchFilter)
	case ExitCodeCancel:
		state.ReturnToCollectionManagement()
		return ui.InitCollectionManagement(co.Collection)
	case ExitCodeAction:
		updatedCollection := result.(models.Collection)
		//TODO: Update position around renamed collection
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitCollectionOptions(updatedCollection, co.SearchFilter)
	default:
		state.RemoveMenuPositions(2)
		return ui.InitCollectionList(co.SearchFilter)
	}
}

func handleToolsTransition(result interface{}, code int) models.Screen {
	switch code {
	case ExitCodeSuccess:
		selection := result.(string)
		state.AddNewMenuPosition()
		switch selection {
		case "Global Actions":
			return ui.InitGlobalActionsScreen()
		case "Play History":
			return ui.InitPlayHistoryListScreen(nil)
		}
		return ui.InitToolsScreen()
	case ExitCodeAction:
		state.ReturnToMain()
		return ui.InitMainMenu()
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleGlobalActionsTransition(code int) models.Screen {
	switch code {
	case ExitCodeCancel:
		state.RemoveMenuPositions(1)
		return ui.InitToolsScreen()
	default:
		return ui.InitGlobalActionsScreen()
	}
}

func handleGamesListTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	gl := currentScreen.(ui.GameList)

	switch code {
	case ExitCodeSuccess:
		return handleGameSelection(gl, result)
	case ExitCodeCancel:
		return handleGameListBack(gl)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitSearch(gl.RomDirectory)
	case ExitCodeEmpty:
		return handleEmptyGamesList(gl)
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleGameSelection(gl ui.GameList, result interface{}) models.Screen {
	selections := result.(shared.Items)

	if len(selections) == 0 {
		return ui.InitGamesListWithPreviousDirectory(gl.RomDirectory, gl.PreviousRomDirectory, gl.SearchFilter)
	}

	if len(selections) == 1 {
		return handleSingleGameSelection(gl, selections[0])
	}

	state.AddNewMenuPosition()
	return ui.InitBulkOptionsScreen(selections, gl.RomDirectory, gl.PreviousRomDirectory, gl.SearchFilter)
}

func handleSingleGameSelection(gl ui.GameList, selection shared.Item) models.Screen {
	state.AddNewMenuPosition()
	if selection.IsDirectory && !selection.IsMultiDiscDirectory && !selection.IsSelfContainedDirectory {
		newRomDirectory := shared.RomDirectory{
			DisplayName: selection.DisplayName,
			Tag:         gl.RomDirectory.Tag,
			Path:        selection.Path,
		}
		return ui.InitGamesListWithPreviousDirectory(newRomDirectory, gl.RomDirectory, "")
	}

	return ui.InitActionsScreen(selection, gl.RomDirectory, gl.PreviousRomDirectory, gl.SearchFilter)
}

func handleGameListBack(gl ui.GameList) models.Screen {
	if gl.PreviousRomDirectory.Path != "" {
		state.RemoveMenuPositions(1)
		return ui.InitGamesList(gl.PreviousRomDirectory, "")
	}

	if gl.SearchFilter != "" {
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitGamesList(gl.RomDirectory, "")
	}

	state.ReturnToMain()
	return ui.InitMainMenu()
}

func handleEmptyGamesList(gl ui.GameList) models.Screen {
	if gl.SearchFilter != "" {
		utils.ShowTimedMessage(fmt.Sprintf("No results found for %s!", gl.SearchFilter), shortMessageDelay)
		state.AddNewMenuPosition()
		return ui.InitSearch(gl.RomDirectory)
	}

	state.ReturnToMain()
	utils.ShowTimedMessage(fmt.Sprintf("%s is empty!", gl.RomDirectory.DisplayName), longMessageDelay)
	return ui.InitMainMenu()
}

func handleSearchBoxTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	search := currentScreen.(ui.Search)
	searchFilter := ""

	if code == ExitCodeSuccess {
		searchFilter = result.(string)
	}

	state.RemoveMenuPositions(1)
	return ui.InitGamesList(search.RomDirectory, searchFilter)
}

func handleActionsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	as := currentScreen.(ui.ActionsScreen)

	if code != ExitCodeSuccess {
		state.RemoveMenuPositions(1)
		return ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	action := models.ActionMap[result.(string)]
	return executeGameAction(as, action)
}

func executeGameAction(as ui.ActionsScreen, action sum.Int[models.Action]) models.Screen {
	switch action {
	case models.Actions.DownloadArt:
		state.AddNewMenuPosition()
		return ui.InitDownloadArtScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter, state.GetAppState().Config.ArtDownloadType)
	case models.Actions.DeleteArt:
		return handleDeleteArtAction(as)
	case models.Actions.RenameRom:
		return handleRenameRomAction(as)
	case models.Actions.CollectionAdd:
		state.AddNewMenuPosition()
		return ui.InitAddToCollectionScreen([]shared.Item{as.Game}, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	case models.Actions.ClearGameTracker:
		return handleClearGameTrackerAction(as)
	case models.Actions.ArchiveRom:
		state.AddNewMenuPosition()
		return ui.InitAddToArchiveScreen([]shared.Item{as.Game}, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	case models.Actions.DeleteRom:
		return handleDeleteRomAction(as)
	case models.Actions.Nuke:
		return handleNukeAction(as)
	case models.Actions.PlayHistoryOpen:
		state.AddNewMenuPosition()
		return ui.InitPlayHistoryGameDetailsScreenFromActions(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	default:
		state.RemoveMenuPositions(1)
		return ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}
}

func handleDeleteArtAction(as ui.ActionsScreen) models.Screen {
	logger := common.GetLoggerInstance()

	existingArtPath, err := utils.FindExistingArt(as.Game.Filename, as.RomDirectory)
	if err != nil {
		logger.Error("Failed to find existing art", zap.Error(err))
		utils.ShowTimedMessage("Unable to delete art!", longMessageDelay)
		return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	if confirmDeletion("Delete this beautiful art?", existingArtPath) {
		common.DeleteFile(existingArtPath)
	}

	return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
}

func handleRenameRomAction(as ui.ActionsScreen) models.Screen {
	newName, err := gaba.Keyboard(as.Game.DisplayName)
	if err != nil {
		utils.ShowTimedMessage("Unable to rename ROM!", longMessageDelay)
		return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	if !newName.IsSome() {
		return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	newFilename := newName.Unwrap()
	newPath, err := utils.RenameRom(as.Game, newFilename, as.RomDirectory)
	if err != nil {
		utils.ShowTimedMessage("Unable to rename ROM!", longMessageDelay)
		return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	as.Game.DisplayName = newFilename
	as.Game.Filename = newPath

	return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
}

func handleClearGameTrackerAction(as ui.ActionsScreen) models.Screen {
	message := fmt.Sprintf("Clear %s from Game Tracker?", as.Game.DisplayName)
	if !utils.ConfirmAction(message) {
		return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	success := utils.ClearGameTracker(as.Game.Filename, as.RomDirectory)
	if success {
		utils.ShowTimedMessage("Game Tracker data cleared!", longMessageDelay)
	} else {
		utils.ShowTimedMessage("Unable to clear Game Tracker data!", longMessageDelay)
	}

	return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
}

func handleDeleteRomAction(as ui.ActionsScreen) models.Screen {
	message := fmt.Sprintf("Delete %s?", as.Game.DisplayName)
	if utils.ConfirmAction(message) {
		utils.DeleteRom(as.Game, as.RomDirectory)
		//TODO: Update position around deleted rom
		state.RemoveMenuPositions(1)
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
}

func handleNukeAction(as ui.ActionsScreen) models.Screen {
	message := fmt.Sprintf("Nuke %s?", as.Game.DisplayName)
	if utils.ConfirmAction(message) {
		utils.Nuke(as.Game, as.RomDirectory)
		//TODO: Update position around deleted rom
		state.RemoveMenuPositions(1)
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
	}

	return ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
}

func handleBulkActionsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	ba := currentScreen.(ui.BulkOptionsScreen)

	if code != ExitCodeSuccess {
		state.RemoveMenuPositions(1)
		return ui.InitGamesListWithPreviousDirectory(ba.RomDirectory, ba.PreviousRomDirectory, ba.SearchFilter)
	}

	action := models.ActionMap[result.(string)]

	if action == models.Actions.CollectionAdd {
		state.AddNewMenuPosition()
		return ui.InitAddToCollectionScreen(ba.Games, ba.RomDirectory, ba.PreviousRomDirectory, ba.SearchFilter)
	}

	return executeBulkAction(ba, action)
}

func executeBulkAction(ba ui.BulkOptionsScreen, action sum.Int[models.Action]) models.Screen {
	switch action {
	case models.Actions.DownloadArt:
		handleBulkDownloadArt(ba)
	case models.Actions.DeleteArt:
		handleBulkDeleteArt(ba)
	case models.Actions.ArchiveRom:
		state.AddNewMenuPosition()
		return ui.InitAddToArchiveScreen(ba.Games, ba.RomDirectory, ba.PreviousRomDirectory, ba.SearchFilter)
	case models.Actions.DeleteRom:
		handleBulkDelete(ba)
	case models.Actions.Nuke:
		handleBulkNuke(ba)
	}
	state.RemoveMenuPositions(1)
	return ui.InitGamesListWithPreviousDirectory(ba.RomDirectory, ba.PreviousRomDirectory, ba.SearchFilter)
}

func handleBulkDownloadArt(ba ui.BulkOptionsScreen) {
	var artPaths []string

	gaba.ProcessMessage(
		fmt.Sprintf("Downloading art for %d games...", len(ba.Games)),
		gaba.ProcessMessageOptions{ShowThemeBackground: true},
		func() (interface{}, error) {
			for _, game := range ba.Games {
				if artPath := utils.FindArt(ba.RomDirectory, game, state.GetAppState().Config.ArtDownloadType, state.GetAppState().Config.FuzzySearchThreshold); artPath != "" {
					artPaths = append(artPaths, artPath)
				}
			}
			return nil, nil
		},
	)

	showArtDownloadResult(artPaths, len(ba.Games))
}

func showArtDownloadResult(artPaths []string, totalGames int) {
	if len(artPaths) == 0 {
		utils.ShowTimedMessage("No art found!", standardMessageDelay)
		return
	}

	if totalGames > 1 {
		message := fmt.Sprintf("Art found for %d/%d games!", len(artPaths), totalGames)
		utils.ShowTimedMessage(message, standardMessageDelay)
	}
}

func handleBulkDeleteArt(ba ui.BulkOptionsScreen) {
	if utils.ConfirmBulkAction("Delete art for the selected games?") {
		for _, game := range ba.Games {
			utils.DeleteArt(game.Filename, ba.RomDirectory)
		}
	}
}

func handleBulkDelete(ba ui.BulkOptionsScreen) {
	if utils.ConfirmBulkAction("Delete the selected games?") {
		for _, game := range ba.Games {
			utils.DeleteRom(game, ba.RomDirectory)
		}
	}
}

func handleBulkNuke(ba ui.BulkOptionsScreen) {
	if utils.ConfirmBulkAction("Nuke the selected games?") {
		for _, game := range ba.Games {
			utils.Nuke(game, ba.RomDirectory)
		}
	}
}

func handleAddToCollectionTransition(currentScreen models.Screen, code int) models.Screen {
	atc := currentScreen.(ui.AddToCollectionScreen)

	switch code {
	case ExitCodeSuccess:
		//TODO: Update position around collection removed from list due to game added
		state.UpdateCurrentMenuPosition(0, 0)
		return ui.InitAddToCollectionScreen(atc.Games, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
	case ExitCodeAction, ExitCodeEmpty:
		state.AddNewMenuPosition()
		return ui.InitCreateCollectionScreen(atc.Games, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
	case ExitCodeCancel:
		return navigateBackFromAddToCollection(atc)
	default:
		return navigateBackFromAddToCollection(atc)
	}
}

func navigateBackFromAddToCollection(atc ui.AddToCollectionScreen) models.Screen {
	state.RemoveMenuPositions(1)
	if len(atc.Games) == 1 {
		return ui.InitActionsScreen(atc.Games[0], atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
	}
	return ui.InitBulkOptionsScreen(atc.Games, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
}

func handleCollectionCreateTransition(currentScreen models.Screen) models.Screen {
	cc := currentScreen.(ui.CreateCollectionScreen)
	state.RemoveMenuPositions(1)
	return ui.InitAddToCollectionScreen(cc.Games, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)
}

func handleAddToArchiveTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	atas := currentScreen.(ui.AddToArchiveScreen)

	switch code {
	case ExitCodeSuccess:
		state.RemoveMenuPositions(1)
		return ui.InitGamesListWithPreviousDirectory(atas.RomDirectory, atas.PreviousRomDirectory, atas.SearchFilter)
	case ExitCodeEmpty:
		return ui.InitAddToArchiveScreen(atas.Games, atas.RomDirectory, atas.PreviousRomDirectory, atas.SearchFilter)
	case ExitCodeAction:
		state.AddNewMenuPosition()
		return ui.InitArchiveCreateScreen(atas.Games, atas.RomDirectory, atas.PreviousRomDirectory, atas.SearchFilter)
	default:
		state.RemoveMenuPositions(1)
		if len(atas.Games) > 1 {
			return ui.InitBulkOptionsScreen(atas.Games, atas.RomDirectory, atas.PreviousRomDirectory, atas.SearchFilter)
		}
		return ui.InitActionsScreen(atas.Games[0], atas.RomDirectory, atas.PreviousRomDirectory, atas.SearchFilter)
	}
}

func handleArchiveCreateTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	acs := currentScreen.(ui.ArchiveCreateScreen)
	state.RemoveMenuPositions(1)
	return ui.InitAddToArchiveScreen(acs.Games, acs.RomDirectory, acs.PreviousRomDirectory, acs.SearchFilter)
}

func handleDownloadArtTransition(currentScreen models.Screen) models.Screen {
	das := currentScreen.(ui.DownloadArtScreen)
	state.RemoveMenuPositions(1)
	return ui.InitActionsScreen(das.Game, das.RomDirectory, das.PreviousRomDirectory, das.SearchFilter)
}

func confirmDeletion(message, imagePath string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "I Changed My Mind"},
		{ButtonName: "A", HelpText: "Trash It!"},
	}, gaba.MessageOptions{
		ImagePath: imagePath,
	})

	return err == nil && result.IsSome()
}
