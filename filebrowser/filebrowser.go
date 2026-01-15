package filebrowser

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nextui-game-manager/common"
	"nextui-game-manager/shared"
)

type FileBrowser struct {
	logger           *slog.Logger
	WorkingDirectory string
	Items            shared.Items
	HumanReadableLS  map[string]shared.Item
}

func NewFileBrowser(logger *slog.Logger) *FileBrowser {
	return &FileBrowser{
		logger: logger,
	}
}

func (c *FileBrowser) CWD(newDirectory string, hideEmpty bool) error {
	return c.CWDDepth(newDirectory, hideEmpty, 1)
}

func (c *FileBrowser) CWDDepth(newDirectory string, hideEmpty bool, maxDepth int) error {
	c.WorkingDirectory = newDirectory
	updatedHumanReadable := make(map[string]shared.Item)

	allItems, err := FindAllItemsWithDepth(c.WorkingDirectory, maxDepth)
	if err != nil {
		return fmt.Errorf("unable to list directory: %w", err)
	}

	var items []shared.Item
	for _, item := range allItems {
		if !item.IsDirectory || (item.IsDirectory && (item.DirectoryFileCount > 0 || !hideEmpty)) {
			items = append(items, item)
			updatedHumanReadable[item.DisplayName] = item
		}
	}

	c.Items = items
	c.HumanReadableLS = updatedHumanReadable

	return nil
}

func FindAllItemsWithDepth(rootPath string, maxDepth int) ([]shared.Item, error) {
	var items []shared.Item

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		if maxDepth >= 0 {
			depth := strings.Count(relPath, string(os.PathSeparator)) + 1

			if depth > maxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		displayName, tag := common.ItemNameCleaner(info.Name(), false)

		if info.IsDir() && tag == "" && filepath.Dir(path) == common.GetRomDirectory() {
			tag = displayName
		}

		item := shared.Item{
			DisplayName:  displayName,
			Filename:     info.Name(),
			Path:         path,
			IsDirectory:  info.IsDir(),
			LastModified: info.ModTime().Format(time.RFC3339),
			Tag:          tag,
		}

		if info.IsDir() {
			item.FileSize = "-"
			contents, err := ListFilesInFolder(item.Path, false)

			if err != nil {
				return err
			}

			item.DirectoryFileCount = len(contents)

			item.IsSelfContainedDirectory = isSelfContainedDirectory(item.Filename, contents)

			for _, f := range contents {
				if strings.Contains(f.DisplayName, "(Disc") ||
					strings.Contains(f.DisplayName, "(Disk") {
					item.IsMultiDiscDirectory = true
					break
				}
			}

		}

		items = append(items, item)
		return nil
	})

	return items, err
}

func isSelfContainedDirectory(directoryName string, contents []shared.Item) bool {
	// Rule 1: If directory contains a directory, it's false
	for _, item := range contents {
		if item.IsDirectory {
			return false
		}
	}

	// Rule 2: One file with same name as directory, .m3u extension, only .m3u
	m3uCount := 0
	hasMatchingM3u := false
	for _, item := range contents {
		if strings.HasSuffix(strings.ToLower(item.Filename), ".m3u") {
			m3uCount++
			itemNameWithoutExt := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
			if itemNameWithoutExt == directoryName {
				hasMatchingM3u = true
			}
		}
	}
	if m3uCount == 1 && hasMatchingM3u {
		return true
	}

	// Rule 3: One file with same name as directory, .cue extension, only .cue
	cueCount := 0
	hasMatchingCue := false
	for _, item := range contents {
		if strings.HasSuffix(strings.ToLower(item.Filename), ".cue") {
			cueCount++
			itemNameWithoutExt := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
			if itemNameWithoutExt == directoryName {
				hasMatchingCue = true
			}
		}
	}
	if cueCount == 1 && hasMatchingCue {
		return true
	}

	// Rule 4: Only one file in directory and it has same name as directory
	if len(contents) == 1 {
		itemNameWithoutExt := strings.TrimSuffix(contents[0].Filename, filepath.Ext(contents[0].Filename))
		if itemNameWithoutExt == directoryName {
			return true
		}
	}

	return false
}

func ListFilesInFolder(folderPath string, recursive bool) ([]shared.Item, error) {
	depth := 1
	if recursive {
		depth = -1
	}

	items, err := FindAllItemsWithDepth(folderPath, depth)
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
	}

	return items, nil
}
