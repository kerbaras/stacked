package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Move to the next branch up in the stack",
	RunE:  runUp,
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Move to the next branch down in the stack",
	RunE:  runDown,
}

var gotoCmd = &cobra.Command{
	Use:   "goto <n>",
	Short: "Jump to the nth branch in the stack (1-indexed)",
	Args:  cobra.ExactArgs(1),
	RunE:  runGoto,
}

func runUp(cmd *cobra.Command, args []string) error {
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

	idx := st.IndexOf(currentBranch)
	if idx < 0 {
		return fmt.Errorf("current branch %q not found in stack", currentBranch)
	}
	if idx >= len(st.Branches)-1 {
		return fmt.Errorf("already at the top of the stack")
	}

	target := st.Branches[idx+1].Name
	if err := repo.Checkout(target, false); err != nil {
		return fmt.Errorf("checkout %s: %w", target, err)
	}

	fmt.Printf("switched to %s\n", target)
	return nil
}

func runDown(cmd *cobra.Command, args []string) error {
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

	idx := st.IndexOf(currentBranch)
	if idx < 0 {
		return fmt.Errorf("current branch %q not found in stack", currentBranch)
	}
	if idx == 0 {
		return fmt.Errorf("already at the bottom of the stack")
	}

	target := st.Branches[idx-1].Name
	if err := repo.Checkout(target, false); err != nil {
		return fmt.Errorf("checkout %s: %w", target, err)
	}

	fmt.Printf("switched to %s\n", target)
	return nil
}

func runGoto(cmd *cobra.Command, args []string) error {
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid number: %s", args[0])
	}

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

	if n < 1 || n > len(st.Branches) {
		return fmt.Errorf("branch number %d out of range (1-%d)", n, len(st.Branches))
	}

	target := st.Branches[n-1].Name
	if err := repo.Checkout(target, false); err != nil {
		return fmt.Errorf("checkout %s: %w", target, err)
	}

	fmt.Printf("switched to %s\n", target)
	return nil
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(gotoCmd)
}
