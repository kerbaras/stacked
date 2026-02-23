package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kerbaras/stacked/pkg/git"
	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/spf13/cobra"
)

// RebaseState tracks an in-progress cascade rebase for continue/abort.
type RebaseState struct {
	StackName      string `json:"stack_name"`
	BranchIndex    int    `json:"branch_index"`
	OriginalBranch string `json:"original_branch"`
}

const rebaseStateFile = "stacked-rebase-state.json"

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Cascade rebase all branches in the stack",
	RunE:  runRebase,
}

func runRebase(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	st, stackName, currentBranch, err := mustCurrentStack(repo, store)
	if err != nil {
		return err
	}

	return cascadeRebase(repo, store, st, stackName, currentBranch, 0)
}

func cascadeRebase(repo *git.Repo, store *stack.Store, st *stack.Stack, stackName, originalBranch string, startIndex int) error {
	for i := startIndex; i < len(st.Branches); i++ {
		br := st.Branches[i]
		parent := st.Parent(br.Name)
		if parent == "" {
			continue
		}

		fmt.Fprintf(os.Stderr, "rebasing %s onto %s...\n", br.Name, parent)
		if err := repo.Rebase(parent, br.Name); err != nil {
			state := RebaseState{
				StackName:      stackName,
				BranchIndex:    i,
				OriginalBranch: originalBranch,
			}
			if saveErr := saveRebaseState(repo.GitDir(), state); saveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save rebase state: %v\n", saveErr)
			}
			return fmt.Errorf("rebase conflict on %s; resolve conflicts, then run `stacked continue`", br.Name)
		}
	}

	// Restore original branch
	if err := repo.Checkout(originalBranch, false); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not restore original branch %s: %v\n", originalBranch, err)
	}

	fmt.Fprintln(os.Stderr, "rebase complete")
	return store.Save()
}

func saveRebaseState(gitDir string, state RebaseState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(gitDir, rebaseStateFile), data, 0o644)
}

func loadRebaseState(gitDir string) (*RebaseState, error) {
	data, err := os.ReadFile(filepath.Join(gitDir, rebaseStateFile))
	if err != nil {
		return nil, err
	}
	var state RebaseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func removeRebaseState(gitDir string) {
	os.Remove(filepath.Join(gitDir, rebaseStateFile))
}

func init() {
	rootCmd.AddCommand(rebaseCmd)
}
