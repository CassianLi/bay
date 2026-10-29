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
			tw := newTable(cmd.OutOrStdout())
			fmt.Fprintf(tw, "added\t%d\n", res.Added)
			fmt.Fprintf(tw, "updated\t%d\n", res.Updated)
			fmt.Fprintf(tw, "missing\t%d\n", res.Missing)
			return tw.Flush()
		},
	}
}
