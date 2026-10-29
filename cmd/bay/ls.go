package main

import (
	"fmt"

	"github.com/jinrunli/bay/internal/model"
	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	var (
		statusStr string
		unknown   bool
		root      string
	)
	cmd := &cobra.Command{
		Use:   "ls [keyword]",
		Short: "List registered projects",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := model.ListFilter{
				UnknownOnly: unknown,
				Root:        root,
			}
			if statusStr != "" {
				stStatus, err := model.ParseStatus(statusStr)
				if err != nil {
					return err
				}
				f.Status = &stStatus
			}
			if len(args) == 1 {
				f.Query = args[0]
			}
			projects, err := st.List(f)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, p := range projects {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", p.Name, p.Status, p.Root, p.Path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&statusStr, "status", "", "filter by status")
	cmd.Flags().BoolVar(&unknown, "unknown", false, "only projects with unknown purpose")
	cmd.Flags().StringVar(&root, "root", "", "filter by root path")
	return cmd
}
