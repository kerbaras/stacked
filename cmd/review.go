package cmd

import (
	"context"
	"fmt"

	gh "github.com/kerbaras/stacked/pkg/github"
	"github.com/kerbaras/stacked/pkg/stack"
	"github.com/kerbaras/stacked/pkg/ui"
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
	var prTasks []ui.Task
	for i := range st.Branches {
		i := i
		br := &st.Branches[i]
		base := st.Parent(br.Name)

		label := fmt.Sprintf("Ensuring PR for %s", ui.BranchName(br.Name))
		if br.PR != nil {
			label = fmt.Sprintf("Checking %s for %s", ui.PRRef(*br.PR), ui.BranchName(br.Name))
		}

		prTasks = append(prTasks, ui.Task{
			Label: label,
			Run: func() error {
				prNum, _, err := gh.EnsurePR(ctx, client, owner, repoName, gh.CreatePRInput{
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
				br.PR = prNum
				return nil
			},
		})
	}

	if err := ui.RunTasks(prTasks); err != nil {
		return err
	}

	// Save PR numbers before updating bodies
	if err := store.Save(); err != nil {
		return err
	}

	// Second pass: update all PR bodies with diagrams
	var diagramTasks []ui.Task
	for i, br := range st.Branches {
		i, br := i, br
		if br.PR == nil {
			continue
		}
		diagramTasks = append(diagramTasks, ui.Task{
			Label: fmt.Sprintf("Updating %s diagram", ui.PRRef(*br.PR)),
			Run: func() error {
				diagram := buildDiagram(st.Branches, st.Base, i, owner, repoName)
				pr, err := client.GetPR(ctx, owner, repoName, *br.PR)
				if err != nil {
					return nil
				}
				newBody := gh.UpdateBody(pr.Body, diagram)
				if newBody != pr.Body {
					if err := client.UpdatePR(ctx, owner, repoName, *br.PR, gh.UpdatePRInput{Body: newBody}); err != nil {
						ui.Warnf("could not update %s body: %v", ui.PRRef(*br.PR), err)
					}
				}
				return nil
			},
		})
	}

	if len(diagramTasks) > 0 {
		if err := ui.RunTasks(diagramTasks); err != nil {
			return err
		}
	}

	ui.Success("all PRs up to date")
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
