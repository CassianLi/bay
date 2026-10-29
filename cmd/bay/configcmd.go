package main

import (
	"fmt"
	"os"

	"github.com/CassianLi/bay/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or manage bay configuration",
		Annotations: map[string]string{
			"bay.skipStore":      "1",
			"bay.skipConfigSave": "1",
		},
	}
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigPathCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the active config file path and contents",
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
			fmt.Fprintf(cmd.OutOrStdout(), "path: %s\n", path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Fprintln(cmd.OutOrStdout(), "(file missing — defaults would be used; run `bay setup`)")
				data, err := yaml.Marshal(config.Default())
				if err != nil {
					return err
				}
				fmt.Fprint(cmd.OutOrStdout(), string(data))
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path only",
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
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}
