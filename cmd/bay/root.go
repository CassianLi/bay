package main

import (
	"fmt"
	"os"

	"github.com/CassianLi/bay/internal/config"
	"github.com/CassianLi/bay/internal/store"
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	cfg     config.Config
	st      *store.Store
)

func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func annotated(cmd *cobra.Command, key string) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Annotations[key] == "1" {
			return true
		}
	}
	return false
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bay",
		Short:         "Local project ledger and lifecycle moves",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cfgPath == "" {
				p, err := config.DefaultConfigPath()
				if err != nil {
					return err
				}
				cfgPath = p
			}
			if annotated(cmd, "bay.skipStore") && annotated(cmd, "bay.skipConfigSave") {
				return nil
			}

			var err error
			cfg, err = config.LoadOrDefault(cfgPath)
			if err != nil {
				return err
			}
			if !annotated(cmd, "bay.skipConfigSave") {
				if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
					if err := cfg.Save(cfgPath); err != nil {
						return fmt.Errorf("save default config: %w", err)
					}
				}
			}
			if annotated(cmd, "bay.skipStore") {
				return nil
			}
			dbPath, err := config.DefaultDBPath()
			if err != nil {
				return err
			}
			st, err = store.Open(dbPath)
			if err != nil {
				return err
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if st == nil {
				return nil
			}
			return st.Close()
		},
	}
	cmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config.yaml (default: ~/.config/bay/config.yaml)")
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newSetupCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newScanCmd())
	cmd.AddCommand(newLsCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDoctorCmd())
	return cmd
}
