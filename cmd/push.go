package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all branches in the stack (force-with-lease)",
	RunE:  runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
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

	syncHint(repo, st)

	for _, br := range st.Branches {
		fmt.Fprintf(os.Stderr, "pushing %s...\n", br.Name)
		if err := repo.Push(br.Name, true); err != nil {
			return fmt.Errorf("push %s: %w", br.Name, err)
		}
	}

	fmt.Fprintln(os.Stderr, "all branches pushed")
	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
