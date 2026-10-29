package move

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CassianLi/bay/internal/model"
	"github.com/CassianLi/bay/internal/paths"
)

type Plan struct {
	From, To string
	NeedMove bool
}

func PlanMove(root, name, pmDir, currentPath string, to model.Status) (Plan, error) {
	dest := paths.ExpectedPath(root, name, pmDir, to)
	from := filepath.Clean(currentPath)
	toPath := filepath.Clean(dest)
	return Plan{
		From:     from,
		To:       toPath,
		NeedMove: from != toPath,
	}, nil
}

func Apply(plan Plan) error {
	if !plan.NeedMove {
		return nil
	}
	if _, err := os.Stat(plan.To); err == nil {
		return fmt.Errorf("destination already exists: %s", plan.To)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plan.To), 0o755); err != nil {
		return err
	}
	// ponytail: same-filesystem rename only; cross-device needs copy+remove
	return os.Rename(plan.From, plan.To)
}
