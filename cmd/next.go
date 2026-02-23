package cmd

import (
	"fmt"
	"time"

	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next <title>",
	Short: "Add the next branch to the current stack",
	Args:  cobra.ExactArgs(1),
	RunE:  runNext,
}

func runNext(cmd *cobra.Command, args []string) error {
	title := args[0]

	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	st, stackName, _, err := mustCurrentStack(repo, store)
	if err != nil {
		return err
	}

	syncHint(repo, st)

	// Checkout tip before creating the next branch
	tip := st.Tip()
	if tip != nil {
		if err := repo.Checkout(tip.Name, false); err != nil {
			return fmt.Errorf("checkout tip %s: %w", tip.Name, err)
		}
	}

	index := len(st.Branches) + 1
	branchName := stack.BranchName(stackName, index, title)

	if err := repo.Checkout(branchName, true); err != nil {
		return fmt.Errorf("create branch %s: %w", branchName, err)
	}

	st.Branches = append(st.Branches, stack.Branch{
		Name:      branchName,
		Title:     title,
		CreatedAt: time.Now(),
	})

	if err := store.Save(); err != nil {
		return err
	}

	fmt.Printf("created branch %s (#%d in stack)\n", branchName, index)
	return nil
}

func init() {
	rootCmd.AddCommand(nextCmd)
}
