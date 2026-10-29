package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinrunli/bay/internal/move"
	"github.com/jinrunli/bay/internal/model"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var (
		statusStr string
		summary   string
		root      string
		yes       bool
		dryRun    bool
	)
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Update status and/or summary; move on disk when status changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if statusStr == "" && !cmd.Flags().Changed("summary") {
				return fmt.Errorf("require --status and/or --summary")
			}

			p, err := st.GetByName(args[0], root)
			if err != nil {
				return err
			}

			changedStatus := false
			if statusStr != "" {
				newStatus, err := model.ParseStatus(statusStr)
				if err != nil {
					return err
				}
				p.Status = newStatus
				p.UserSetStatus = true
				changedStatus = true
			}
			if cmd.Flags().Changed("summary") {
				p.Summary = summary
				p.PurposeKnown = true
			}

			out := cmd.OutOrStdout()
			pmDir := cfg.PMDir
			if pmDir == "" {
				pmDir = "_pm"
			}

			if changedStatus {
				plan, err := move.PlanMove(p.Root, p.Name, pmDir, p.Path, p.Status)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s → %s\n", plan.From, plan.To)

				if dryRun {
					return nil
				}

				if plan.NeedMove && !yes {
					fmt.Fprint(out, "proceed? [y/N] ")
					line, err := bufio.NewReader(os.Stdin).ReadString('\n')
					if err != nil {
						return err
					}
					ans := strings.TrimSpace(strings.ToLower(line))
					if ans != "y" && ans != "yes" {
						return fmt.Errorf("cancelled")
					}
				}

				if err := move.Apply(plan); err != nil {
					return err
				}
				if plan.NeedMove {
					p.Path = plan.To
					p.LastMovedAt = time.Now().UTC()
				}
			} else if dryRun {
				return nil
			}

			return st.UpdateMeta(p)
		},
	}
	cmd.Flags().StringVar(&statusStr, "status", "", "new status")
	cmd.Flags().StringVar(&summary, "summary", "", "project purpose summary")
	cmd.Flags().StringVar(&root, "root", "", "disambiguate by root path")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print plan only; do not move or update")
	return cmd
}
