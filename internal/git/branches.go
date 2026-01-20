package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	iexec "github.com/labset/git-extensions/internal/exec"
)

type Branch struct {
	Name string
	Date string
}

func GetDefaultBranch() (string, error) {
	var result string

	// Try to get the default branch from remote
	err := iexec.WithOutput("git symbolic-ref refs/remotes/origin/HEAD --short", func(output string) error {
		parts := strings.Split(output, "/")
		if len(parts) > 1 {
			result = parts[len(parts)-1]
		}
		return nil
	})
	if err == nil && result != "" {
		return result, nil
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if cmd.Run() == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

func GetCurrentBranch() (string, error) {
	var result string
	err := iexec.WithOutput("git rev-parse --abbrev-ref HEAD", func(output string) error {
		result = output
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return result, nil
}

func GetPurgeableBranches(defaultBranch string) ([]Branch, []string) {
	var warnings []string

	currentBranch, err := GetCurrentBranch()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not detect current branch: %v", err))
		return nil, warnings
	}

	merged, err := getMergedBranches(defaultBranch)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not detect merged branches: %v", err))
	}

	squashed, err := getSquashedBranches(defaultBranch)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not detect squashed branches: %v", err))
	}

	// Combine and dedupe
	seen := make(map[string]bool)
	var purgeable []Branch

	for _, name := range append(merged, squashed...) {
		if name == defaultBranch || name == currentBranch {
			continue
		}
		if !seen[name] {
			seen[name] = true
			purgeable = append(purgeable, Branch{Name: name})
		}
	}

	return purgeable, warnings
}

func getMergedBranches(defaultBranch string) ([]string, error) {
	var branches []string
	err := iexec.WithOutput("git branch --merged "+defaultBranch+" --format=%(refname:short)", func(output string) error {
		for _, line := range strings.Split(output, "\n") {
			if line != "" {
				branches = append(branches, line)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}
	return branches, nil
}

func getSquashedBranches(defaultBranch string) ([]string, error) {
	var allBranches []string
	err := iexec.WithOutput("git branch --format=%(refname:short)", func(output string) error {
		for _, line := range strings.Split(output, "\n") {
			if line != "" {
				allBranches = append(allBranches, line)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var squashed []string
	for _, branch := range allBranches {
		if branch == defaultBranch {
			continue
		}
		if isSquashMerged(branch, defaultBranch) {
			squashed = append(squashed, branch)
		}
	}
	return squashed, nil
}

func isSquashMerged(branch, defaultBranch string) bool {
	var mergeBase string
	err := iexec.WithOutput("git merge-base "+defaultBranch+" "+branch, func(output string) error {
		mergeBase = output
		return nil
	})
	if err != nil {
		return false
	}

	var branchTree string
	err = iexec.WithOutput("git rev-parse "+branch+"^{tree}", func(output string) error {
		branchTree = output
		return nil
	})
	if err != nil {
		return false
	}

	var tempCommit string
	err = iexec.WithOutput("git commit-tree "+branchTree+" -p "+mergeBase+" -m temp", func(output string) error {
		tempCommit = output
		return nil
	})
	if err != nil {
		return false
	}

	var isSquashed bool
	err = iexec.WithOutput("git cherry "+defaultBranch+" "+tempCommit, func(output string) error {
		if output == "" {
			isSquashed = true
			return nil
		}
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "+") {
				isSquashed = false
				return nil
			}
		}
		isSquashed = true
		return nil
	})
	if err != nil {
		return false
	}

	return isSquashed
}

func DeleteBranches(branches []string) error {
	if len(branches) == 0 {
		return nil
	}

	args := append([]string{"branch", "-D"}, branches...)
	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete branches: %s", stderr.String())
	}
	return nil
}

func GetRecentBranches() ([]Branch, error) {
	var branches []Branch
	err := iexec.WithOutput("git for-each-ref refs/heads/ --sort=-committerdate --format=%(committerdate:short)|%(refname:short)", func(output string) error {
		for _, line := range strings.Split(output, "\n") {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				branches = append(branches, Branch{Date: parts[0], Name: parts[1]})
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent branches: %w", err)
	}

	return branches, nil
}

func SwitchBranch(branch string) error {
	return iexec.WithOutput(fmt.Sprintf("git checkout %s", branch), func(output string) error {
		return nil
	})
}
