package paths

import (
	"path/filepath"
	"strings"

	"github.com/CassianLi/bay/internal/model"
)

func ExpectedPath(root, name, pmDir string, st model.Status) string {
	switch st {
	case model.StatusIdea:
		return filepath.Join(root, pmDir, "incubator", name)
	case model.StatusActive, model.StatusPaused, model.StatusKept:
		return filepath.Join(root, name)
	case model.StatusArchived:
		return filepath.Join(root, pmDir, "archive", name)
	case model.StatusTrash:
		return filepath.Join(root, pmDir, "trash", name)
	default:
		return filepath.Join(root, name)
	}
}

func InferStatus(root, absPath, pmDir string) model.Status {
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return model.StatusPaused
	}
	rel = filepath.ToSlash(rel)
	parts := strings.Split(rel, "/")
	if len(parts) >= 2 {
		switch parts[0] {
		case pmDir:
			switch parts[1] {
			case "incubator":
				return model.StatusIdea
			case "archive":
				return model.StatusArchived
			case "trash":
				return model.StatusTrash
			}
		case "_incubator":
			return model.StatusIdea
		case "_archive":
			return model.StatusArchived
		}
	}
	return model.StatusPaused
}

func ShouldSkipDirName(name, pmDir string, ignore []string) bool {
	if name == pmDir {
		return true
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	for _, ig := range ignore {
		if name == ig {
			return true
		}
	}
	return false
}
