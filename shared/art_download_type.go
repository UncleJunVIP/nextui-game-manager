package shared

import "qlova.tech/sum"

type ArtDownloadType struct {
	BOX_ART,
	TITLE_SCREEN,
	LOGOS,
	SCREENSHOTS sum.Int[ArtDownloadType]
}

var ArtDownloadTypes = sum.Int[ArtDownloadType]{}.Sum()

var ArtDownloadTypeMapping = map[sum.Int[ArtDownloadType]]string{
	ArtDownloadTypes.BOX_ART:      "Named_Boxarts",
	ArtDownloadTypes.TITLE_SCREEN: "Named_Titles",
	ArtDownloadTypes.LOGOS:        "Named_Logos",
	ArtDownloadTypes.SCREENSHOTS:  "Named_Snaps",
}

var ArtDownloadTypeFromString = map[string]sum.Int[ArtDownloadType]{
	"BOX_ART":      ArtDownloadTypes.BOX_ART,
	"TITLE_SCREEN": ArtDownloadTypes.TITLE_SCREEN,
	"LOGOS":        ArtDownloadTypes.LOGOS,
	"SCREENSHOTS":  ArtDownloadTypes.SCREENSHOTS,
}
