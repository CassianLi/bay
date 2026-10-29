package main

import (
	"fmt"
	"os"

	"github.com/jinrunli/bay/internal/config"
	"github.com/jinrunli/bay/internal/setup"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive configuration wizard",
		Long: `Interactively create or update ~/.config/bay/config.yaml
(roots, stale_after_days, pm_dir, ignore).`,
		Annotations: map[string]string{
			"bay.skipStore":      "1",
			"bay.skipConfigSave": "1",
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := cfgPath
			if path == "" {
				p, err := config.DefaultConfigPath()
				if err != nil {
					return err
				}
				path = p
			}
			if _, err := os.Stat(path); err == nil && !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Config already exists: %s\n", path)
				fmt.Fprintln(cmd.OutOrStdout(), "Re-run with --force to overwrite via the wizard.")
				return fmt.Errorf("config exists (use --force)")
			}

			base, err := config.LoadOrDefault(path)
			if err != nil {
				return err
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				base = config.Default()
			}
			_, err = setup.Wizard(cmd.InOrStdin(), cmd.OutOrStdout(), base, path)
			return err
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing config via wizard")
	return cmd
}
