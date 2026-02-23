package cmd

import (
	"fmt"

	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all branches in the stack (force-with-lease)",
	RunE:  runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	st, _, _, err := mustCurrentStack(repo, store)
	if err != nil {
		return err
	}

	syncHint(repo, st)

	var tasks []ui.Task
	for _, br := range st.Branches {
		br := br
		tasks = append(tasks, ui.Task{
			Label: fmt.Sprintf("Pushing %s", ui.BranchName(br.Name)),
			Run: func() error {
				return repo.PushSilent(br.Name, true)
			},
		})
	}

	if err := ui.RunTasks(tasks); err != nil {
		return err
	}

	ui.Success("all branches pushed")
	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
