package utils

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"path/filepath"
	"strings"
)

const CHEAT_HOST_ROOT = "https://cheats.unclejun.vip"

var inMemoryCheatMap = make(map[string]map[string]shared.Item)

func init() {
	go fetchCheats()
}

func fetchCheats() {
	if common.IsConnectedToInternet() {
		systemMapping := common.LoadSystemMapping()
		client := common.NewHttpTableClient(CHEAT_HOST_ROOT, shared.HostTypes.APACHE, shared.TableColumns{}, nil, nil)

		for short, full := range systemMapping {
			res, err := client.ListDirectory(full)

			if err != nil {
				continue
			}

			if res != nil && len(res) > 0 {
				inMemoryCheatMap[short] = make(map[string]shared.Item)

				for idx, cheat := range res {
					cleaned, _ := common.ItemNameCleaner(cheat.Filename, true)
					res[idx].DisplayName = cleaned
					inMemoryCheatMap[short][cleaned] = res[idx]
				}
			}
		}
	}
}

func CheatFileAvailable(game shared.Item, directory shared.RomDirectory) bool {
	if _, ok := inMemoryCheatMap[directory.Tag]; ok {
		cleaned, _ := common.ItemNameCleaner(game.Filename, true)
		if _, hasCheat := inMemoryCheatMap[directory.Tag][cleaned]; hasCheat {
			return true
		}
	}

	return false
}

func FindExistingCheat(game shared.Item, directory shared.RomDirectory) (string, bool) {
	cheatFile := strings.ReplaceAll(game.Filename, filepath.Ext(game.Filename), ".cht")
	cheatFullPath := filepath.Join(GetCheatDirectory(), directory.Tag, cheatFile)

	if DoesFileExists(cheatFullPath) {
		return cheatFullPath, true
	}

	return "", false
}

func DownloadCheatFile() {

}

func DeleteCheatFile(game shared.Item, directory shared.RomDirectory) {

}
