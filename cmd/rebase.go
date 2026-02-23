package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kerbaras/stacked/pkg/git"
	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/kerbaras/stacked/pkg/ui"
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
	var tasks []ui.Task
	for i := startIndex; i < len(st.Branches); i++ {
		i, br := i, st.Branches[i]
		parent := st.Parent(br.Name)
		if parent == "" {
			continue
		}

		tasks = append(tasks, ui.Task{
			Label: fmt.Sprintf("Rebasing %s onto %s", ui.BranchName(br.Name), ui.BranchName(parent)),
			Run: func() error {
				if err := repo.RebaseSilent(parent, br.Name); err != nil {
					state := RebaseState{
						StackName:      stackName,
						BranchIndex:    i,
						OriginalBranch: originalBranch,
					}
					if saveErr := saveRebaseState(repo.GitDir(), state); saveErr != nil {
						ui.Warnf("failed to save rebase state: %v", saveErr)
					}
					return fmt.Errorf("conflict on %s; resolve conflicts, then run `stacked continue`", br.Name)
				}
				return nil
			},
		})
	}

	if len(tasks) > 0 {
		if err := ui.RunTasks(tasks); err != nil {
			return err
		}
	}

	// Restore original branch
	if err := repo.CheckoutSilent(originalBranch, false); err != nil {
		ui.Warnf("could not restore original branch %s: %v", originalBranch, err)
	}

	ui.Success("rebase complete")
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
