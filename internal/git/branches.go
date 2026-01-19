package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Branch struct {
	Name string
}

func GetDefaultBranch() (string, error) {
	// Try to get the default branch from remote
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	out, err := cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "/")
		if len(parts) > 1 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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
		// Skip default branch and current branch
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
	cmd := exec.Command("git", "branch", "--merged", defaultBranch, "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

func getSquashedBranches(defaultBranch string) ([]string, error) {
	// Get all local branches
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var squashed []string
	for _, branch := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if branch == "" || branch == defaultBranch {
			continue
		}
		if isSquashMerged(branch, defaultBranch) {
			squashed = append(squashed, branch)
		}
	}
	return squashed, nil
}

func isSquashMerged(branch, defaultBranch string) bool {
	// Get the merge-base between the branch and default
	mergeBaseCmd := exec.Command("git", "merge-base", defaultBranch, branch)
	mergeBase, err := mergeBaseCmd.Output()
	if err != nil {
		return false
	}
	mergeBaseStr := strings.TrimSpace(string(mergeBase))

	// Get the tree of the branch
	treeCmd := exec.Command("git", "rev-parse", branch+"^{tree}")
	branchTree, err := treeCmd.Output()
	if err != nil {
		return false
	}
	branchTreeStr := strings.TrimSpace(string(branchTree))

	// Create a temporary commit with the branch's tree on top of merge-base
	commitCmd := exec.Command("git", "commit-tree", branchTreeStr, "-p", mergeBaseStr, "-m", "temp")
	tempCommit, err := commitCmd.Output()
	if err != nil {
		return false
	}
	tempCommitStr := strings.TrimSpace(string(tempCommit))

	// Check if this tree state is reachable from the default branch
	// using git cherry - if all commits show "-", they're already in default
	cherryCmd := exec.Command("git", "cherry", defaultBranch, tempCommitStr)
	cherryOut, err := cherryCmd.Output()
	if err != nil {
		return false
	}

	// If output is empty or shows "-", the changes are in default branch
	output := strings.TrimSpace(string(cherryOut))
	if output == "" {
		return true
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "+") {
			return false
		}
	}
	return true
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
