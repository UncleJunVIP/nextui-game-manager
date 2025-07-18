package models

import "qlova.tech/sum"

type ScreenName struct {
	MainMenu,
	Settings,
	Tools,

	GamesList,
	SearchBox,
	Actions,
	BulkActions,
	AddToCollection,
	Confirm,
	DownloadArt,

	AddToArchive,
	ArchiveCreate,
	ArchiveList,
	ArchiveManagement,
	ArchiveOptions,
	ArchiveGamesList,

	CollectionsList,
	CollectionOptions,
	CollectionManagement,
	CollectionCreate,

	PlayHistoryActions,
	PlayHistoryGameDetails,
	PlayHistoryGameHistory,
	PlayHistoryGameList,
	PlayHistoryList,

	GlobalActions sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value interface{}, exitCode int, e error)
}
