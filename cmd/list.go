package cmd

import (
	"fmt"
	"sort"

	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all stacks",
	RunE:    runList,
}

func runList(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	if len(store.State.Stacks) == 0 {
		ui.Infof("no stacks — run %s to create one", ui.Bold.Render("`stacked new`"))
		return nil
	}

	currentBranch, _ := repo.CurrentBranch()
	currentStackName := store.CurrentStackName()

	// Sort stack names for stable output
	names := make([]string, 0, len(store.State.Stacks))
	for name := range store.State.Stacks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		st := store.State.Stacks[name]
		marker := " "
		if name == currentStackName {
			marker = "*"
		}

		branchCount := len(st.Branches)
		branchWord := "branches"
		if branchCount == 1 {
			branchWord = "branch"
		}

		label := fmt.Sprintf("%s %s", marker, name)
		if name == currentStackName {
			label = fmt.Sprintf("%s %s", marker, ui.Green.Render(name))
		}

		// Check if current branch is in this stack
		detail := ""
		if st.HasBranch(currentBranch) {
			idx := st.IndexOf(currentBranch)
			detail = fmt.Sprintf(" (on #%d)", idx+1)
		}

		fmt.Printf("%s  %s%s  %s\n", label, ui.Dim.Render(fmt.Sprintf("%d %s", branchCount, branchWord)), detail, ui.Dim.Render("base: "+st.Base))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
