package cmd

import (
	"context"
	"fmt"
	"os"

	gh "github.com/kerbaras/stacked/pkg/github"
	"github.com/kerbaras/stacked/pkg/stack"
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
	fmt.Fprintln(os.Stderr, "fetching origin...")
	if err := repo.Fetch("origin"); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if err := repo.Checkout(st.Base, false); err != nil {
		return fmt.Errorf("checkout %s: %w", st.Base, err)
	}
	if err := repo.Pull("origin", st.Base); err != nil {
		return fmt.Errorf("pull %s: %w", st.Base, err)
	}

	// Step 2: Try to get GitHub client for PR operations
	owner, repoName, remoteErr := repo.RemoteOwnerRepo()
	var client gh.Client
	if remoteErr == nil {
		client, _ = gh.NewClientForRepo(owner, repoName, cfg.GitHubToken)
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
			fmt.Fprintf(os.Stderr, "warning: could not check PR #%d: %v\n", *br.PR, err)
			break
		}
		if !merged {
			break
		}

		fmt.Fprintf(os.Stderr, "pruning merged branch %s (PR #%d)\n", br.Name, *br.PR)
		if delErr := repo.DeleteBranch(br.Name); delErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not delete branch %s: %v\n", br.Name, delErr)
		}

		st.Branches = st.Branches[1:]
		prunedCount++
	}

	if len(st.Branches) == 0 {
		// All branches merged — remove the stack
		fmt.Fprintf(os.Stderr, "all branches merged; removing stack %q\n", stackName)
		delete(store.State.Stacks, stackName)
		store.State.CurrentStack = ""
		if err := repo.Checkout(st.Base, false); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not checkout %s: %v\n", st.Base, err)
		}
		return store.Save()
	}

	// Step 4: If we pruned branches, cascade rebase remaining onto base
	if prunedCount > 0 {
		fmt.Fprintln(os.Stderr, "rebasing remaining branches...")
		for _, br := range st.Branches {
			parent := st.Parent(br.Name)
			if parent == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "rebasing %s onto %s...\n", br.Name, parent)
			if err := repo.Rebase(parent, br.Name); err != nil {
				return fmt.Errorf("rebase %s: %w; resolve conflicts then run `stacked continue`", br.Name, err)
			}
		}

		// Retarget bottom PR to base
		if client != nil && len(st.Branches) > 0 && st.Branches[0].PR != nil {
			fmt.Fprintf(os.Stderr, "retargeting PR #%d to %s...\n", *st.Branches[0].PR, st.Base)
			if err := client.UpdatePRBase(ctx, owner, repoName, *st.Branches[0].PR, st.Base); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not retarget PR #%d: %v\n", *st.Branches[0].PR, err)
			}
		}

		// Force-push all remaining branches
		for _, br := range st.Branches {
			fmt.Fprintf(os.Stderr, "pushing %s...\n", br.Name)
			if err := repo.Push(br.Name, true); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not push %s: %v\n", br.Name, err)
			}
		}

		// Update diagrams in all PR bodies
		if client != nil {
			updateAllDiagrams(ctx, client, st, owner, repoName)
		}
	} else {
		// No pruning — just cascade rebase onto updated base
		fmt.Fprintln(os.Stderr, "rebasing stack onto updated base...")
		for _, br := range st.Branches {
			parent := st.Parent(br.Name)
			if parent == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "rebasing %s onto %s...\n", br.Name, parent)
			if err := repo.Rebase(parent, br.Name); err != nil {
				return fmt.Errorf("rebase %s: %w; resolve conflicts then run `stacked continue`", br.Name, err)
			}
		}
	}

	// Restore branch (or first stack branch if current was pruned)
	restoreBranch := currentBranch
	if !st.HasBranch(currentBranch) {
		restoreBranch = st.Branches[0].Name
	}
	if err := repo.Checkout(restoreBranch, false); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not checkout %s: %v\n", restoreBranch, err)
	}

	fmt.Fprintln(os.Stderr, "sync complete")
	return store.Save()
}

func updateAllDiagrams(ctx context.Context, client gh.Client, st *stack.Stack, owner, repo string) {
	for i, br := range st.Branches {
		if br.PR == nil {
			continue
		}

		diagram := buildDiagram(st.Branches, st.Base, i, owner, repo)

		pr, err := client.GetPR(ctx, owner, repo, *br.PR)
		if err != nil {
			continue
		}

		newBody := gh.UpdateBody(pr.Body, diagram)
		if newBody != pr.Body {
			fmt.Fprintf(os.Stderr, "updating PR #%d diagram...\n", *br.PR)
			_ = client.UpdatePR(ctx, owner, repo, *br.PR, gh.UpdatePRInput{Body: newBody})
		}
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
