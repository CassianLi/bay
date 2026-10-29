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
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "name:\t%s\n", p.Name)
			fmt.Fprintf(out, "root:\t%s\n", p.Root)
			fmt.Fprintf(out, "path:\t%s\n", p.Path)
			fmt.Fprintf(out, "status:\t%s\n", p.Status)
			fmt.Fprintf(out, "purpose_known:\t%v\n", p.PurposeKnown)
			fmt.Fprintf(out, "summary:\t%s\n", p.Summary)
			fmt.Fprintf(out, "user_set_status:\t%v\n", p.UserSetStatus)
			fmt.Fprintf(out, "missing:\t%v\n", p.Missing)
			fmt.Fprintf(out, "last_seen_at:\t%s\n", p.LastSeenAt)
			fmt.Fprintf(out, "last_moved_at:\t%s\n", p.LastMovedAt)
			fmt.Fprintf(out, "created_at:\t%s\n", p.CreatedAt)
			fmt.Fprintf(out, "updated_at:\t%s\n", p.UpdatedAt)
			fmt.Fprintf(out, "signals:\t%s\n", p.Signals)
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", "", "disambiguate by root path")
	return cmd
}
