package main

import (
	"fmt"

	"github.com/jinrunli/bay/internal/scan"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan roots and update the project ledger",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := scan.Run(cfg, st)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added=%d updated=%d missing=%d\n", res.Added, res.Updated, res.Missing)
			return nil
		},
	}
}
