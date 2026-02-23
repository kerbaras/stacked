package cmd

import (
	"fmt"

	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the stack graph",
	RunE:  runStatus,
}

var verboseFlag bool

func runStatus(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	store, err := loadStore(repo)
	if err != nil {
		return err
	}

	st, _, currentBranch, err := mustCurrentStack(repo, store)
	if err != nil {
		return err
	}

	syncHint(repo, st)

	var infos []ui.BranchInfo
	for _, br := range st.Branches {
		info := ui.BranchInfo{
			Name:      br.Name,
			Title:     br.Title,
			PRNumber:  br.PR,
			IsCurrent: br.Name == currentBranch,
		}

		if verboseFlag {
			hash, err := repo.BranchTip(br.Name)
			if err == nil {
				info.Hash = hash
			}
		}

		parent := st.Parent(br.Name)
		if parent != "" {
			needsRebase, err := repo.NeedsRebase(br.Name, parent)
			if err == nil {
				info.NeedsRebase = needsRebase
			}
		}

		infos = append(infos, info)
	}

	fmt.Print(ui.RenderGraph(infos, st.Base, verboseFlag))
	return nil
}

func init() {
	statusCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "include commit hashes")
	rootCmd.AddCommand(statusCmd)
}
