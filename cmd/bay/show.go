package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show one registered project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := st.GetByName(args[0], root)
			if err != nil {
				return err
			}
			tw := newTable(cmd.OutOrStdout())
			rows := [][2]string{
				{"name", p.Name},
				{"path", p.Path},
				{"status", string(p.Status)},
				{"purpose", formatPurpose(p.PurposeKnown)},
				{"summary", p.Summary},
				{"note", p.Note},
				{"user_set_status", fmt.Sprintf("%v", p.UserSetStatus)},
				{"missing", fmt.Sprintf("%v", p.Missing)},
				{"last_seen_at", formatTime(p.LastSeenAt)},
				{"last_moved_at", formatTime(p.LastMovedAt)},
				{"created_at", formatTime(p.CreatedAt)},
				{"updated_at", formatTime(p.UpdatedAt)},
				{"signals", p.Signals},
			}
			for _, row := range rows {
				fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
			}
			return tw.Flush()
		},
	}
	cmd.Flags().StringVar(&root, "root", "", "disambiguate by root path")
	return cmd
}
