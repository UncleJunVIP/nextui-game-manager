package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/veandco/go-sdl2/sdl"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type GlobalActionsScreen struct {
}

func InitGlobalActionsScreen() GlobalActionsScreen {
	return GlobalActionsScreen{}
}

func (gas GlobalActionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GlobalActions
}

func (gas GlobalActionsScreen) Draw() (value interface{}, exitCode int, e error) {
	globalActions := models.GlobalActionKeys

	var globalActionEntries []gabagool.MenuItem
	for _, action := range globalActions {
		globalActionEntries = append(globalActionEntries, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: models.GlobalActionMap[action],
		})
	}

	options := gabagool.DefaultListOptions("Global Actions", globalActionEntries)
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)

		if selection.Unwrap().SelectedItem.Metadata == models.Actions.GlobalDownloadArt {
			noArt, err := utils.FindRomsWithoutArt()
			if err != nil {
				utils.ShowTimedMessage("Failed to scan for missing art", time.Second*2)
				return nil, -1, err
			}

			missingArtCount := 0

			for _, missing := range noArt {
				missingArtCount += len(missing)
			}

			if missingArtCount == 0 {
				utils.ShowTimedMessage("All your games have art!\nGo play something!", time.Second*2)
				return nil, 0, nil
			}

			var missingArtPlatforms []gabagool.MenuItem

			for platform, games := range noArt {
				label := "Games"
				if len(games) == 1 {
					label = "Game"
				}

				missingArtPlatforms = append(missingArtPlatforms, gabagool.MenuItem{
					Text:     fmt.Sprintf("%s (%d %s)", platform.DisplayName, len(games), label),
					Selected: true,
					Focused:  false,
					Metadata: platform,
				})
			}

			slices.SortFunc(missingArtPlatforms, func(m1 gabagool.MenuItem, m2 gabagool.MenuItem) int {
				return strings.Compare(m1.Text, m2.Text)
			})

			platformSelectionOptions := gabagool.DefaultListOptions("Platforms Missing Art", missingArtPlatforms)

			platformSelectionOptions.EnableMultiSelect = true
			platformSelectionOptions.StartInMultiSelectMode = true
			platformSelectionOptions.MultiSelectButton = gabagool.ButtonUnassigned
			platformSelectionOptions.MultiSelectKey = sdl.K_0

			platformSelectionOptions.FooterHelpItems = []gabagool.FooterHelpItem{
				{ButtonName: "B", HelpText: "Back"},
				{ButtonName: "A", HelpText: "Select / Unselect"},
				{ButtonName: "Start", HelpText: "Confirm"},
			}

			platformSelection, err := gabagool.List(platformSelectionOptions)
			if err != nil {
				return nil, 0, err
			}

			if platformSelection.IsSome() && platformSelection.Unwrap().SelectedIndex == -1 {
				return nil, 0, nil
			}

			selectedPlatforms := platformSelection.Unwrap().SelectedItems

			if len(selectedPlatforms) == 0 {
				utils.ShowTimedMessage("Please select at least one platform!", time.Second*2)
				return nil, 0, nil
			}

			selectedPlatformsMap := make(map[shared.RomDirectory][]shared.Item)

			selectedMissingArtCount := 0

			var downloads []gabagool.Download

			for _, selection := range selectedPlatforms {
				platform := selection.Metadata.(shared.RomDirectory)
				selectedPlatformsMap[platform] = noArt[platform]
				selectedMissingArtCount += len(noArt[platform])
			}

			platformLabel := "Platform"
			if len(selectedPlatformsMap) > 1 {
				platformLabel = "Platforms"
			}

			gamesLabel := "Game"
			if selectedMissingArtCount > 1 {
				gamesLabel = "Games"
			}

			gabagool.ProcessMessage(fmt.Sprintf("Searching for art...\n%d %s | %d %s Total",
				len(selectedPlatformsMap), platformLabel, selectedMissingArtCount, gamesLabel), gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
				for romDir, games := range selectedPlatformsMap {
					downloads = append(downloads, utils.FindAllArt(romDir, games, state.GetAppState().Config.ArtDownloadType, state.GetAppState().Config.FuzzySearchThreshold)...)
				}
				return nil, nil
			})

			res, err := gabagool.DownloadManager(downloads, make(map[string]string))
			if err != nil {
				utils.ShowTimedMessage("Failed to download art!", time.Second*2)
				return nil, 0, nil
			}

			if len(res.CompletedDownloads) == 0 {
				utils.ShowTimedMessage("No art found!", time.Second*2)
				return
			} else {
				message := fmt.Sprintf("Art found for %d/%d games!", len(res.CompletedDownloads), selectedMissingArtCount)
				utils.ShowTimedMessage(message, time.Second*2)
			}
		} else if selection.Unwrap().SelectedItem.Metadata == models.Actions.GlobalClearRecents {
			confirmClear := utils.ConfirmAction("Are you sure you want to clear your recently played list?\n\nThis cannot be undone!")

			if confirmClear {
				deletedRes, _ := gabagool.ProcessMessage("Clearing Recently Played List.", gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
					time.Sleep(1500 * time.Millisecond)
					return common.DeleteFile(utils.RecentlyPlayedFile), nil
				})

				if deletedRes.Result.(bool) {
					utils.ShowTimedMessage("Recently Played List Cleared!", time.Millisecond*1500)
				} else {
					utils.ShowTimedMessage("Failed to confirmClear recently played list!", time.Millisecond*1500)
				}
			}
		}

		return nil, 0, nil
	}

	return nil, 2, nil
}
