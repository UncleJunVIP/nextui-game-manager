package models

import (
	"qlova.tech/sum"
)

type Action struct {
	RenameRom,
	DownloadArt,
	DeleteArt,
	DownloadCheats,
	DeleteCheats,
	ClearGameTracker,
	ClearSaveStates,
	ArchiveRom,
	ArchiveRename,
	ArchiveDelete,
	DeleteRom,
	Nuke,

	CollectionRename,
	CollectionAlphabetize,
	CollectionDelete,
	CollectionAdd,

	PlayHistoryOpen,
	PlayHistoryAdopt,

	GlobalDownloadArt,
	GlobalClearRecents sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Rename ROM":          Actions.RenameRom,
	"Download Art":        Actions.DownloadArt,
	"Delete Art":          Actions.DeleteArt,
	"Download Cheat File": Actions.DownloadCheats,
	"Delete Cheat File":   Actions.DeleteCheats,
	"Clear Game Tracker":  Actions.ClearGameTracker,
	"Archive ROM":         Actions.ArchiveRom,
	"Rename Archive":      Actions.ArchiveRename,
	"Delete Archive":      Actions.ArchiveDelete,
	"Delete ROM":          Actions.DeleteRom,
	"Nuclear Option":      Actions.Nuke,

	"Rename Collection": Actions.CollectionRename,
	"Alphabetize Games": Actions.CollectionAlphabetize,
	"Delete Collection": Actions.CollectionDelete,
	"Add to Collection": Actions.CollectionAdd,

	"View Play Details": Actions.PlayHistoryOpen,
}

var GlobalActionMap = map[string]sum.Int[Action]{
	"Download Missing Art":  Actions.GlobalDownloadArt,
	"Clear Recently Played": Actions.GlobalClearRecents,
}

var ActionKeys = []string{
	"Rename ROM",
	"Add to Collection",
	//"Clear Save States",
	"Archive ROM",
	"Delete ROM",
	//"Nuclear Option",
}

var GlobalActionKeys = []string{
	"Download Missing Art",
	"Clear Recently Played",
}

var BulkActionKeys = []string{
	"Add to Collection",
	"Download Art",
	"Delete Art",
	//"Clear Game Tracker",
	"Archive ROM",
	//"Delete ROM",
	//"Nuclear Option",
}

var CollectionActionKeys = []string{
	"Alphabetize Games",
	"Rename Collection",
	"Delete Collection",
}

var ArchiveActionKeys = []string{
	"Rename Archive",
	"Delete Archive",
}

var PlayHistoryActionKeys = []string{
	//"Rehome Orphaned History",
	//"Delete Existing History",
}

var ActionNames = map[sum.Int[Action]]string{}

func init() {
	for name, action := range ActionMap {
		ActionNames[action] = name
	}
}
