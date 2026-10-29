package main

import (
	"fmt"
	"os"
	"time"

	"github.com/CassianLi/bay/internal/doctor"
	"github.com/CassianLi/bay/internal/model"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var (
		mdPath string
		root   string
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Report unknown purpose, stale, trash, and missing projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := st.List(model.ListFilter{Root: root})
			if err != nil {
				return err
			}
			stale := cfg.StaleAfterDays
			if stale <= 0 {
				stale = 90
			}
			rep := doctor.Build(projects, stale, time.Now())
			fmt.Fprint(cmd.OutOrStdout(), rep.FormatText())
			if mdPath != "" {
				if err := os.WriteFile(mdPath, []byte(rep.FormatMarkdown()), 0o644); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&mdPath, "md", "", "write Markdown report to file")
	cmd.Flags().StringVar(&root, "root", "", "filter by root path")
	return cmd
}
