package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/labset/git-extensions/internal/git"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "git-purge",
		Short: "Clean up merged and squashed branches",
		Long:  "Interactively select and delete local branches that have been merged or squashed into the default branch.",
		RunE:  run,
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// 1. Get default branch
	defaultBranch, err := git.GetDefaultBranch()
	if err != nil {
		return fmt.Errorf("failed to detect default branch: %w", err)
	}

	// 2. Get purgeable branches
	branches, warnings := git.GetPurgeableBranches(defaultBranch)

	for _, warning := range warnings {
		fmt.Printf("⚠ %s\n", warning)
	}

	if len(branches) == 0 {
		fmt.Println("No purgeable branches found. All clean!")
		return nil
	}

	// 3. Build options for multi-select
	options := make([]huh.Option[string], len(branches))
	for i, branch := range branches {
		options[i] = huh.NewOption(branch.Name, branch.Name)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select branches to delete").
				Description(fmt.Sprintf("Found %d branches merged/squashed into %s", len(branches), defaultBranch)).
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if len(selected) == 0 {
		fmt.Println("No branches selected.")
		return nil
	}

	// 4. Delete selected branches
	if err := git.DeleteBranches(selected); err != nil {
		return err
	}

	fmt.Printf("Deleted %d branch(es)\n", len(selected))
	return nil
}
