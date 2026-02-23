package cmd

import (
	"fmt"

	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue a paused rebase",
	RunE:  runContinue,
}

var abortCmd = &cobra.Command{
	Use:   "abort",
	Short: "Abort a paused rebase",
	RunE:  runAbort,
}

func runContinue(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	state, err := loadRebaseState(repo.GitDir())
	if err != nil {
		return fmt.Errorf("no rebase in progress (missing state file)")
	}

	// Continue the current rebase
	ui.Step("continuing rebase...")
	if err := repo.RebaseContinue(); err != nil {
		return fmt.Errorf("rebase --continue failed; resolve remaining conflicts, then run `stacked continue` again")
	}

	// Remove state file — this branch is done
	removeRebaseState(repo.GitDir())

	// Resume cascade from the next branch
	st := store.State.Stacks[state.StackName]
	if st == nil {
		return fmt.Errorf("stack %q not found", state.StackName)
	}

	nextIndex := state.BranchIndex + 1
	if nextIndex < len(st.Branches) {
		return cascadeRebase(repo, store, st, state.StackName, state.OriginalBranch, nextIndex)
	}

	// No more branches — restore original branch
	if err := repo.CheckoutSilent(state.OriginalBranch, false); err != nil {
		ui.Warnf("could not restore original branch %s: %v", state.OriginalBranch, err)
	}

	ui.Success("rebase complete")
	return store.Save()
}

func runAbort(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	state, err := loadRebaseState(repo.GitDir())
	if err != nil {
		return fmt.Errorf("no rebase in progress (missing state file)")
	}

	if err := repo.RebaseAbort(); err != nil {
		return fmt.Errorf("rebase --abort failed: %w", err)
	}

	removeRebaseState(repo.GitDir())

	// Restore original branch
	if err := repo.CheckoutSilent(state.OriginalBranch, false); err != nil {
		ui.Warnf("could not restore original branch %s: %v", state.OriginalBranch, err)
	}

	ui.Success("rebase aborted")
	return nil
}

func init() {
	rootCmd.AddCommand(continueCmd)
	rootCmd.AddCommand(abortCmd)
}
