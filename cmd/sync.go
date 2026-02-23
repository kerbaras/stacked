package cmd

import (
	"context"
	"fmt"

	gh "github.com/kerbaras/stacked/pkg/github"
	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync stack after merges: prune merged branches, rebase, retarget PRs",
	RunE:  runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
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

	// Step 1: Fetch and update base
	fetchTasks := []ui.Task{
		{
			Label: fmt.Sprintf("Fetching %s", ui.Faint("origin")),
			Run: func() error {
				return repo.FetchSilent("origin")
			},
		},
		{
			Label: fmt.Sprintf("Pulling %s", ui.BranchName(st.Base)),
			Run: func() error {
				if err := repo.CheckoutSilent(st.Base, false); err != nil {
					return fmt.Errorf("checkout %s: %w", st.Base, err)
				}
				return repo.PullSilent("origin", st.Base)
			},
		},
	}
	if err := ui.RunTasks(fetchTasks); err != nil {
		return err
	}

	// Step 2: Try to get GitHub client for PR operations
	owner, repoName, remoteErr := repo.RemoteOwnerRepo()
	var client gh.Client
	if remoteErr == nil {
		client, _ = gh.NewClientForRepo(owner, repoName)
	}

	// Step 3: Prune merged branches from the bottom
	prunedCount := 0
	ctx := context.Background()
	for len(st.Branches) > 0 {
		br := st.Branches[0]
		if br.PR == nil || client == nil {
			break
		}

		merged, err := client.IsMerged(ctx, owner, repoName, *br.PR)
		if err != nil {
			ui.Warnf("could not check %s: %v", ui.PRRef(*br.PR), err)
			break
		}
		if !merged {
			break
		}

		ui.Infof("pruning %s %s", ui.BranchName(br.Name), ui.Faint(fmt.Sprintf("(PR #%d merged)", *br.PR)))
		if delErr := repo.DeleteBranchSilent(br.Name); delErr != nil {
			ui.Warnf("could not delete branch %s: %v", br.Name, delErr)
		}

		st.Branches = st.Branches[1:]
		prunedCount++
	}

	if len(st.Branches) == 0 {
		ui.Successf("all branches merged — removing stack %s", ui.Bold.Render(stackName))
		delete(store.State.Stacks, stackName)
		store.State.CurrentStack = ""
		_ = repo.CheckoutSilent(st.Base, false)
		return store.Save()
	}

	// Step 4: Rebase
	var rebaseTasks []ui.Task
	for _, br := range st.Branches {
		br := br
		parent := st.Parent(br.Name)
		if parent == "" {
			continue
		}
		rebaseTasks = append(rebaseTasks, ui.Task{
			Label: fmt.Sprintf("Rebasing %s onto %s", ui.BranchName(br.Name), ui.BranchName(parent)),
			Run: func() error {
				return repo.RebaseSilent(parent, br.Name)
			},
		})
	}
	if len(rebaseTasks) > 0 {
		if err := ui.RunTasks(rebaseTasks); err != nil {
			return fmt.Errorf("rebase failed: %w; resolve conflicts then run `stacked continue`", err)
		}
	}

	// Step 5: If we pruned, retarget + push + update diagrams
	if prunedCount > 0 {
		var postTasks []ui.Task

		// Retarget bottom PR
		if client != nil && len(st.Branches) > 0 && st.Branches[0].PR != nil {
			prNum := *st.Branches[0].PR
			base := st.Base
			postTasks = append(postTasks, ui.Task{
				Label: fmt.Sprintf("Retargeting %s → %s", ui.PRRef(prNum), ui.BranchName(base)),
				Run: func() error {
					return client.UpdatePRBase(ctx, owner, repoName, prNum, base)
				},
			})
		}

		// Push all remaining branches
		for _, br := range st.Branches {
			br := br
			postTasks = append(postTasks, ui.Task{
				Label: fmt.Sprintf("Pushing %s", ui.BranchName(br.Name)),
				Run: func() error {
					return repo.PushSilent(br.Name, true)
				},
			})
		}

		if len(postTasks) > 0 {
			if err := ui.RunTasks(postTasks); err != nil {
				ui.Warnf("post-sync step failed: %v", err)
			}
		}

		// Update diagrams
		if client != nil {
			updateAllDiagrams(ctx, client, st, owner, repoName)
		}
	}

	// Restore branch
	restoreBranch := currentBranch
	if !st.HasBranch(currentBranch) {
		restoreBranch = st.Branches[0].Name
	}
	_ = repo.CheckoutSilent(restoreBranch, false)

	ui.Success("sync complete")
	return store.Save()
}

func updateAllDiagrams(ctx context.Context, client gh.Client, st *stack.Stack, owner, repo string) {
	var tasks []ui.Task
	for i, br := range st.Branches {
		i, br := i, br
		if br.PR == nil {
			continue
		}
		tasks = append(tasks, ui.Task{
			Label: fmt.Sprintf("Updating %s diagram", ui.PRRef(*br.PR)),
			Run: func() error {
				diagram := buildDiagram(st.Branches, st.Base, i, owner, repo)
				pr, err := client.GetPR(ctx, owner, repo, *br.PR)
				if err != nil {
					return nil
				}
				newBody := gh.UpdateBody(pr.Body, diagram)
				if newBody != pr.Body {
					_ = client.UpdatePR(ctx, owner, repo, *br.PR, gh.UpdatePRInput{Body: newBody})
				}
				return nil
			},
		})
	}
	if len(tasks) > 0 {
		_ = ui.RunTasks(tasks)
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
