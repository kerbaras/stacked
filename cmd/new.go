package cmd

import (
	"fmt"
	"time"

	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new stack",
	Long:  "Create a new stack branching from the specified base branch (default: main).",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

var fromFlag string

func runNew(cmd *cobra.Command, args []string) error {
	title := args[0]

	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	stackName := stack.StackName(title)
	if _, exists := store.State.Stacks[stackName]; exists {
		return fmt.Errorf("stack %q already exists", stackName)
	}

	branchName := stack.BranchName(cfg.BranchTemplate, stackName, 1, title)

	// Ensure we're on the base branch before creating the new branch
	if err := repo.Checkout(fromFlag, false); err != nil {
		return fmt.Errorf("checkout %s: %w", fromFlag, err)
	}

	if err := repo.Checkout(branchName, true); err != nil {
		return fmt.Errorf("create branch %s: %w", branchName, err)
	}

	store.State.Stacks[stackName] = &stack.Stack{
		Base: fromFlag,
		Branches: []stack.Branch{
			{
				Name:      branchName,
				Title:     title,
				CreatedAt: time.Now(),
			},
		},
	}
	store.State.CurrentStack = stackName

	if err := store.Save(); err != nil {
		return err
	}

	ui.Successf("created stack %s on %s", ui.Bold.Render(stackName), ui.BranchName(branchName))
	return nil
}

func init() {
	newCmd.Flags().StringVar(&fromFlag, "from", "main", "base branch to stack on")
	rootCmd.AddCommand(newCmd)
}
