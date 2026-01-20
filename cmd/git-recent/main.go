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
		Use:   "git-recent",
		Short: "Switch to a recently used branch",
		Long:  "Interactively select and switch to a branch, sorted by most recent commit date. Displays the commit date for each branch.",
		RunE:  run,
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	branches, err := git.GetRecentBranches()
	if err != nil {
		return fmt.Errorf("failed to get recent branches: %w", err)
	}

	if len(branches) == 0 {
		fmt.Println("No other branches found.")
		return nil
	}

	options := make([]huh.Option[string], len(branches))
	for i, branch := range branches {
		label := fmt.Sprintf("%s  %s", branch.Date, branch.Name)
		options[i] = huh.NewOption(label, branch.Name)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a branch to switch to").
				Description(fmt.Sprintf("Found %d branches", len(branches))).
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if selected == "" {
		fmt.Println("No branch selected.")
		return nil
	}

	if err := git.SwitchBranch(selected); err != nil {
		return err
	}

	fmt.Printf("Switched to branch '%s'\n", selected)
	return nil
}
