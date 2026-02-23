package cmd

import (
	"context"
	"fmt"
	"os"

	gh "github.com/kerbaras/stacked/pkg/github"
	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Create or update GitHub PRs for all branches in the stack",
	RunE:  runReview,
}

func runReview(cmd *cobra.Command, args []string) error {
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

	owner, repoName, err := repo.RemoteOwnerRepo()
	if err != nil {
		return err
	}

	client, err := gh.NewClientForRepo(owner, repoName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// First pass: ensure all PRs exist
	for i := range st.Branches {
		br := &st.Branches[i]
		base := st.Parent(br.Name)

		prNum, url, err := gh.EnsurePR(ctx, client, owner, repoName, gh.CreatePRInput{
			Owner: owner,
			Repo:  repoName,
			Title: br.Title,
			Body:  "",
			Head:  br.Name,
			Base:  base,
		}, br.PR)
		if err != nil {
			return fmt.Errorf("ensure PR for %s: %w", br.Name, err)
		}

		if br.PR == nil {
			fmt.Fprintf(os.Stderr, "created PR #%d for %s: %s\n", *prNum, br.Name, url)
		}
		br.PR = prNum
	}

	// Save PR numbers before updating bodies (in case diagram update fails)
	if err := store.Save(); err != nil {
		return err
	}

	// Second pass: update all PR bodies with diagrams
	for i, br := range st.Branches {
		if br.PR == nil {
			continue
		}

		diagram := buildDiagram(st.Branches, st.Base, i, owner, repoName)

		pr, err := client.GetPR(ctx, owner, repoName, *br.PR)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not fetch PR #%d: %v\n", *br.PR, err)
			continue
		}

		newBody := gh.UpdateBody(pr.Body, diagram)
		if newBody != pr.Body {
			fmt.Fprintf(os.Stderr, "updating PR #%d diagram...\n", *br.PR)
			if err := client.UpdatePR(ctx, owner, repoName, *br.PR, gh.UpdatePRInput{Body: newBody}); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not update PR #%d body: %v\n", *br.PR, err)
			}
		}
	}

	fmt.Fprintln(os.Stderr, "all PRs up to date")
	return nil
}

func buildDiagram(branches []stack.Branch, base string, currentIdx int, owner, repo string) string {
	var diagramBranches []gh.DiagramBranch
	for i, br := range branches {
		db := gh.DiagramBranch{
			Title:     br.Title,
			PRNumber:  br.PR,
			IsCurrent: i == currentIdx,
		}
		if br.PR != nil {
			db.PRURL = fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, *br.PR)
		}
		diagramBranches = append(diagramBranches, db)
	}
	return gh.RenderDiagram(diagramBranches, base)
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
