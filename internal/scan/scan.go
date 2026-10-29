package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jinrunli/bay/internal/config"
	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/paths"
	"github.com/jinrunli/bay/internal/store"
)

type Result struct {
	Added, Updated, Missing int
}

func Run(cfg config.Config, st *store.Store) (Result, error) {
	var res Result
	pmDir := cfg.PMDir
	if pmDir == "" {
		pmDir = "_pm"
	}

	for _, rootCfg := range cfg.Roots {
		rootAbs, err := filepath.Abs(rootCfg)
		if err != nil {
			return res, fmt.Errorf("abs root %q: %w", rootCfg, err)
		}
		fi, err := os.Stat(rootAbs)
		if err != nil {
			return res, fmt.Errorf("stat root %q: %w", rootAbs, err)
		}
		if !fi.IsDir() {
			return res, fmt.Errorf("root not a directory: %s", rootAbs)
		}

		seen, err := discoverRoot(rootAbs, pmDir, cfg.Ignore, st, &res)
		if err != nil {
			return res, err
		}
		n, err := st.MarkMissingExcept(rootAbs, seen)
		if err != nil {
			return res, fmt.Errorf("mark missing %q: %w", rootAbs, err)
		}
		res.Missing += n
	}
	return res, nil
}

func discoverRoot(root, pmDir string, ignore []string, st *store.Store, res *Result) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read root %q: %w", root, err)
	}

	var seen []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		abs := filepath.Join(root, name)

		switch {
		case name == pmDir:
			for _, zone := range []string{"incubator", "archive", "trash"} {
				zoneDir := filepath.Join(abs, zone)
				names, err := upsertChildren(root, zoneDir, pmDir, ignore, st, res)
				if err != nil {
					return nil, err
				}
				seen = append(seen, names...)
			}
		case name == "_incubator" || name == "_archive":
			names, err := upsertChildren(root, abs, pmDir, ignore, st, res)
			if err != nil {
				return nil, err
			}
			seen = append(seen, names...)
		case paths.ShouldSkipDirName(name, pmDir, ignore):
			continue
		default:
			if err := upsertProject(root, abs, pmDir, st, res); err != nil {
				return nil, err
			}
			seen = append(seen, filepath.Base(abs))
		}
	}
	return seen, nil
}

func upsertChildren(root, parent, pmDir string, ignore []string, st *store.Store, res *Result) ([]string, error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %q: %w", parent, err)
	}
	var seen []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if paths.ShouldSkipDirName(name, pmDir, ignore) {
			continue
		}
		abs := filepath.Join(parent, name)
		if err := upsertProject(root, abs, pmDir, st, res); err != nil {
			return nil, err
		}
		seen = append(seen, name)
	}
	return seen, nil
}

func upsertProject(root, absPath, pmDir string, st *store.Store, res *Result) error {
	name := filepath.Base(absPath)
	sig, err := CollectSignals(absPath)
	if err != nil {
		return fmt.Errorf("signals %q: %w", absPath, err)
	}
	sigJSON, err := json.Marshal(sig)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = st.GetByName(name, root)
	isNew := err == store.ErrNotFound
	if err != nil && err != store.ErrNotFound {
		return err
	}

	p := model.Project{
		Name:          name,
		Root:          root,
		Path:          absPath,
		Status:        paths.InferStatus(root, absPath, pmDir),
		PurposeKnown:  false,
		Summary:       "",
		Signals:       string(sigJSON),
		UserSetStatus: false,
		LastSeenAt:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := st.UpsertScan(p); err != nil {
		return fmt.Errorf("upsert %q: %w", name, err)
	}
	if isNew {
		res.Added++
	} else {
		res.Updated++
	}
	return nil
}
