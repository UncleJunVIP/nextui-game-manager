package common

import "regexp"

const SDCardRoot = "/mnt/SDCARD"
const RomDirectory = "/mnt/SDCARD/Roms"
const CollectionDirectory = "/mnt/SDCARD/Collections"

var OrderedFolderRegex = regexp.MustCompile(`\d+\)\s`)
var TagRegex = regexp.MustCompile(`\((.*?)\)`)
