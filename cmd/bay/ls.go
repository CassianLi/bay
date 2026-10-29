package main

import (
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
			rows := make([][]string, 0, len(projects)+1)
			rows = append(rows, []string{"NAME", "STATUS", "SUMMARY", "PATH"})
			for _, p := range projects {
				rows = append(rows, []string{
					p.Name,
					string(p.Status),
					listSummary(p.Summary, p.Signals, 40),
					p.Path,
				})
			}
			return writeAligned(cmd.OutOrStdout(), rows, 2)
		},
	}
	cmd.Flags().StringVar(&statusStr, "status", "", "filter by status ("+model.StatusValuesHint()+")")
	cmd.Flags().BoolVar(&unknown, "unknown", false, "only projects with unknown purpose")
	cmd.Flags().StringVar(&root, "root", "", "filter by root path")
	return cmd
}
