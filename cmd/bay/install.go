package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinrunli/bay/internal/config"
	"github.com/jinrunli/bay/internal/install"
	"github.com/jinrunli/bay/internal/scan"
	"github.com/jinrunli/bay/internal/setup"
	"github.com/jinrunli/bay/internal/store"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		binDir     string
		skipSetup  bool
		forceSetup bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install bay binary and optionally run config wizard",
		Long: `Copy the current bay executable into a bin directory (default: ~/bin
or GOPATH/bin), print PATH hints if needed, then guide interactive
configuration when config is missing (or with --force-setup).`,
		Annotations: map[string]string{
			"bay.skipStore":      "1",
			"bay.skipConfigSave": "1",
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			in := cmd.InOrStdin()
			rd := bufio.NewReader(in)

			if binDir == "" {
				def, err := install.DefaultBinDir()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Install directory [%s]: ", def)
				line, err := rd.ReadString('\n')
				if err != nil && err != io.EOF {
					return err
				}
				line = strings.TrimSpace(line)
				if line == "" {
					binDir = def
				} else {
					binDir = line
				}
			}
			expanded, err := expandUserPath(binDir)
			if err != nil {
				return err
			}
			absBin, err := filepath.Abs(expanded)
			if err != nil {
				return err
			}
			dest := filepath.Join(absBin, "bay")
			fmt.Fprintf(out, "Installing to %s ...\n", dest)
			if err := install.CopyExecutable(dest); err != nil {
				return err
			}
			fmt.Fprintln(out, "Binary installed.")

			if !install.PathContainsDir(absBin) {
				fmt.Fprintln(out)
				fmt.Fprintf(out, "Warning: %s is not on PATH.\n", absBin)
				fmt.Fprintf(out, "Add it, for example:\n  %s\n", install.PathHint(absBin))
			} else {
				fmt.Fprintf(out, "PATH ok: %s\n", absBin)
			}

			if skipSetup {
				return nil
			}

			cfgFile := cfgPath
			if cfgFile == "" {
				cfgFile, err = config.DefaultConfigPath()
				if err != nil {
					return err
				}
			}
			_, statErr := os.Stat(cfgFile)
			needWizard := os.IsNotExist(statErr) || forceSetup
			if !needWizard {
				fmt.Fprintf(out, "\nConfig already exists: %s\n", cfgFile)
				fmt.Fprint(out, "Run config wizard now? [y/N]: ")
				line, _ := rd.ReadString('\n')
				ans := strings.TrimSpace(strings.ToLower(line))
				needWizard = ans == "y" || ans == "yes"
			}
			if !needWizard {
				fmt.Fprintln(out, "Skipping setup. Run `bay setup --force` later if needed.")
				return nil
			}

			fmt.Fprintln(out)
			base := config.Default()
			if statErr == nil {
				if loaded, err := config.LoadOrDefault(cfgFile); err == nil {
					base = loaded
				}
			}
			saved, err := setup.Wizard(in, out, base, cfgFile)
			if err != nil {
				return err
			}

			fmt.Fprint(out, "Run first scan now? [Y/n]: ")
			line, _ := rd.ReadString('\n')
			ans := strings.TrimSpace(strings.ToLower(line))
			if ans != "" && ans != "y" && ans != "yes" {
				fmt.Fprintln(out, "Done. Run `bay scan` when ready.")
				return nil
			}
			dbPath, err := config.DefaultDBPath()
			if err != nil {
				return err
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return err
			}
			defer s.Close()
			res, err := scan.Run(saved, s)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "scan: added=%d updated=%d missing=%d\n", res.Added, res.Updated, res.Missing)
			fmt.Fprintln(out, "Install complete.")
			return nil
		},
	}
	cmd.Flags().StringVar(&binDir, "bin-dir", "", "install directory (default: prompt, usually ~/bin or GOPATH/bin)")
	cmd.Flags().BoolVar(&skipSetup, "skip-setup", false, "only install binary; skip config wizard")
	cmd.Flags().BoolVar(&forceSetup, "force-setup", false, "always run config wizard after install")
	return cmd
}

func expandUserPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
