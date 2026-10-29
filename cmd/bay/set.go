package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinrunli/bay/internal/model"
	"github.com/jinrunli/bay/internal/move"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var (
		statusStr string
		summary   string
		note      string
		root      string
		yes       bool
		dryRun    bool
	)
	cmd := &cobra.Command{
		Use:   "set <name> [name...]",
		Short: "Update status, summary, and/or note; move on disk when status changes",
		Long: `Update one or more projects' status, purpose summary, and/or remark note.

Status values: ` + model.StatusValuesHint() + `

Changing status may move project directories under _pm/{incubator,archive,trash}.
With multiple names, plans are shown together and confirmed once (unless --yes).`,
		Example: `  bay set my-tool --status paused --note "先搁置，下周再看"
  bay set a b c --status archived --note "批量归档试运行" --yes
  bay set my-tool --summary "关税试算小工具"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			changedSummary := cmd.Flags().Changed("summary")
			changedNote := cmd.Flags().Changed("note")
			if statusStr == "" && !changedSummary && !changedNote {
				return fmt.Errorf("require --status, --summary, and/or --note (status: %s)", model.StatusValuesHint())
			}

			var newStatus model.Status
			if statusStr != "" {
				stStatus, err := model.ParseStatus(statusStr)
				if err != nil {
					return err
				}
				newStatus = stStatus
			}

			type item struct {
				project model.Project
				plan    move.Plan
				needPlan bool
			}
			items := make([]item, 0, len(args))
			for _, name := range args {
				p, err := st.GetByName(name, root)
				if err != nil {
					return fmt.Errorf("%s: %w", name, err)
				}
				if statusStr != "" {
					p.Status = newStatus
					p.UserSetStatus = true
				}
				if changedSummary {
					trimmed := strings.TrimSpace(summary)
					if trimmed != "" {
						p.Summary = trimmed
						p.PurposeKnown = true
					} else {
						p.Summary = ""
						p.PurposeKnown = false
					}
				}
				if changedNote {
					p.Note = strings.TrimSpace(note)
				}

				it := item{project: p}
				if statusStr != "" {
					pmDir := cfg.PMDir
					if pmDir == "" {
						pmDir = "_pm"
					}
					plan, err := move.PlanMove(p.Root, p.Name, pmDir, p.Path, p.Status)
					if err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}
					it.plan = plan
					it.needPlan = true
				}
				items = append(items, it)
			}

			out := cmd.OutOrStdout()
			needConfirm := false
			for _, it := range items {
				if it.needPlan {
					fmt.Fprintf(out, "%s: %s → %s\n", it.project.Name, it.plan.From, it.plan.To)
					if it.plan.NeedMove {
						needConfirm = true
					}
				} else {
					fmt.Fprintf(out, "%s: (no move)\n", it.project.Name)
				}
				if changedNote {
					fmt.Fprintf(out, "  note: %s\n", it.project.Note)
				}
			}

			if dryRun {
				return nil
			}

			if needConfirm && !yes {
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

			for _, it := range items {
				p := it.project
				if it.needPlan {
					if err := move.Apply(it.plan); err != nil {
						return fmt.Errorf("%s: %w", p.Name, err)
					}
					if it.plan.NeedMove {
						p.Path = it.plan.To
						p.LastMovedAt = time.Now().UTC()
					}
				}
				if err := st.UpdateMeta(p); err != nil {
					return fmt.Errorf("%s: %w", p.Name, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&statusStr, "status", "", "new status ("+model.StatusValuesHint()+")")
	cmd.Flags().StringVar(&summary, "summary", "", "project purpose summary")
	cmd.Flags().StringVar(&note, "note", "", "remark when setting (especially status changes)")
	cmd.Flags().StringVar(&root, "root", "", "disambiguate by root path")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print plan only; do not move or update")
	return cmd
}
