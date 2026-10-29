package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CassianLi/bay/internal/config"
)

// Wizard runs an interactive config flow. base is the starting config (usually Default()).
func Wizard(in io.Reader, out io.Writer, base config.Config, cfgPath string) (config.Config, error) {
	rd := bufio.NewReader(in)
	cfg := base

	fmt.Fprintln(out, "bay setup — interactive configuration")
	fmt.Fprintf(out, "Config file: %s\n\n", cfgPath)

	rootsDefault := strings.Join(cfg.Roots, ", ")
	rootsLine, err := prompt(rd, out, "Project roots (comma-separated)", rootsDefault)
	if err != nil {
		return config.Config{}, err
	}
	roots := splitCSV(rootsLine)
	if len(roots) == 0 {
		return config.Config{}, fmt.Errorf("at least one root is required")
	}
	absRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		abs, err := expandPath(r)
		if err != nil {
			return config.Config{}, err
		}
		absRoots = append(absRoots, abs)
	}
	cfg.Roots = absRoots

	staleDefault := strconv.Itoa(cfg.StaleAfterDays)
	if cfg.StaleAfterDays <= 0 {
		staleDefault = "90"
	}
	staleLine, err := prompt(rd, out, "Stale after days (doctor)", staleDefault)
	if err != nil {
		return config.Config{}, err
	}
	stale, err := strconv.Atoi(strings.TrimSpace(staleLine))
	if err != nil || stale <= 0 {
		return config.Config{}, fmt.Errorf("invalid stale_after_days %q", staleLine)
	}
	cfg.StaleAfterDays = stale

	pmDefault := cfg.PMDir
	if pmDefault == "" {
		pmDefault = "_pm"
	}
	pmLine, err := prompt(rd, out, "Lifecycle dir name under each root", pmDefault)
	if err != nil {
		return config.Config{}, err
	}
	cfg.PMDir = strings.TrimSpace(pmLine)
	if cfg.PMDir == "" {
		cfg.PMDir = "_pm"
	}

	ignoreDefault := strings.Join(cfg.Ignore, ", ")
	ignoreLine, err := prompt(rd, out, "Extra ignore dir names (comma-separated, optional)", ignoreDefault)
	if err != nil {
		return config.Config{}, err
	}
	cfg.Ignore = splitCSV(ignoreLine)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Will save:")
	fmt.Fprintf(out, "  roots:            %s\n", strings.Join(cfg.Roots, ", "))
	fmt.Fprintf(out, "  stale_after_days: %d\n", cfg.StaleAfterDays)
	fmt.Fprintf(out, "  pm_dir:           %s\n", cfg.PMDir)
	fmt.Fprintf(out, "  ignore:           %v\n", cfg.Ignore)
	fmt.Fprintf(out, "  path:             %s\n", cfgPath)

	ok, err := promptYesNo(rd, out, "Write config?", true)
	if err != nil {
		return config.Config{}, err
	}
	if !ok {
		return config.Config{}, fmt.Errorf("cancelled")
	}
	if err := cfg.Save(cfgPath); err != nil {
		return config.Config{}, err
	}
	fmt.Fprintln(out, "Config saved.")
	return cfg, nil
}

func prompt(rd *bufio.Reader, out io.Writer, label, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(out, "%s: ", label)
	}
	line, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

func promptYesNo(rd *bufio.Reader, out io.Writer, label string, defYes bool) (bool, error) {
	hint := "Y/n"
	if !defYes {
		hint = "y/N"
	}
	fmt.Fprintf(out, "%s [%s]: ", label, hint)
	line, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	ans := strings.TrimSpace(strings.ToLower(line))
	if ans == "" {
		return defYes, nil
	}
	return ans == "y" || ans == "yes", nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func expandPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}
	return filepath.Abs(p)
}
