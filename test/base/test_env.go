/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package base

import (
	"fmt"
	"github.com/mercedes-benz/gitflow-cli/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	// Import the plugin package so that init functions for all plugins are executed automatically
	_ "github.com/mercedes-benz/gitflow-cli/plugin"
)

// GitTestEnv manages local repository and simulated remote repository
type GitTestEnv struct {
	LocalPath  string // Path to local repository
	RemotePath string // Path to simulated remote repository
	t          *testing.T
}

// SetupTestEnv creates test environment with local repo and simulated remote
func SetupTestEnv(t *testing.T) *GitTestEnv {
	t.Helper()

	// Create temporary directories for test repositories
	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "local")
	remotePath := filepath.Join(tmpDir, "remote")

	// Ensure directories exist
	require.NoError(t, os.MkdirAll(localPath, 0755))
	require.NoError(t, os.MkdirAll(remotePath, 0755))

	// Initialize remote repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remotePath
	require.NoError(t, cmd.Run(), "Failed to initialize bare remote repository")

	// Initialize local repository
	cmd = exec.Command("git", "init")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to initialize local repository")

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "noreply@mercedes-benz.com")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run())

	// Add remote to local repository
	cmd = exec.Command("git", "remote", "add", "origin", remotePath)
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to add remote to local repository")

	// Create main branches (production and development)
	env := &GitTestEnv{
		LocalPath:  localPath,
		RemotePath: remotePath,
		t:          t,
	}

	env.WriteFile("README.md", "# Test Repository")
	env.ExecuteGit("add", "README.md")
	env.ExecuteGit("commit", "-m", "Initial commit")
	env.ExecuteGit("branch", "-m", "main")
	env.ExecuteGit("push", "-u", "origin", "main")
	env.ExecuteGit("checkout", "-b", "develop")
	env.ExecuteGit("push", "-u", "origin", "develop")
	env.ExecuteGit("checkout", "main")

	return env
}

// ExecuteGitflow calls the Gitflow functionality directly via the Go API
func (env *GitTestEnv) ExecuteGitflow(args ...string) string {
	env.t.Helper()

	// Set command line arguments with the --path parameter
	os.Args = append([]string{"gitflow-cli", "--path", env.LocalPath}, args...)

	// Capture output using a pipe
	r, w, err := os.Pipe()
	require.NoError(env.t, err)

	// Save original stdout/stderr and replace with pipe
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	// Execute the command
	cmd.Execute()

	// Restore original stdout/stderr and close the write end of pipe
	os.Stdout, os.Stderr = oldStdout, oldStderr
	w.Close()

	// Read the captured output
	output, err := io.ReadAll(r)
	require.NoError(env.t, err)

	// Log the command output
	env.t.Logf("Command output for 'gitflow-cli %s':\n%s", strings.Join(args, " "), string(output))

	return string(output)
}

// ExecuteGit runs a git command in the local repository
func (env *GitTestEnv) ExecuteGit(args ...string) string {
	env.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = env.LocalPath
	output, err := cmd.CombinedOutput()
	require.NoError(env.t, err, "Git command failed: git %s\nOutput: %s", strings.Join(args, " "), output)
	return string(output)
}

// ExecuteGitAllowError runs a git command but doesn't fail the test if it returns an error
func (env *GitTestEnv) ExecuteGitAllowError(args ...string) (string, error) {
	env.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = env.LocalPath
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// WriteFile creates a file in the local repository with the given content
func (env *GitTestEnv) WriteFile(path, content string) {
	env.t.Helper()
	fullPath := filepath.Join(env.LocalPath, path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	require.NoError(env.t, err, "Failed to create file: %s", path)
}

// GetFileContent reads the content of a file in the local repository
func (env *GitTestEnv) GetFileContent(path string) string {
	env.t.Helper()
	fullPath := filepath.Join(env.LocalPath, path)
	content, err := os.ReadFile(fullPath)
	require.NoError(env.t, err, "Failed to read file: %s", path)
	return string(content)
}

// GetFileContentFromCommit gets the content of a file at a specific commit/branch
func (env *GitTestEnv) GetFileContentFromCommit(commitOrBranch, path string) string {
	env.t.Helper()
	output := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitOrBranch, path))
	return output
}

// AssertBranchExists checks if a branch exists
func (env *GitTestEnv) AssertBranchExists(branch string) {
	env.t.Helper()
	_, err := env.ExecuteGitAllowError("rev-parse", "--verify", branch)
	assert.NoError(env.t, err, "Branch %s does not exist", branch)
}

// AssertBranchDoesNotExist checks that a branch does not exist
func (env *GitTestEnv) AssertBranchDoesNotExist(branch string) {
	env.t.Helper()
	_, err := env.ExecuteGitAllowError("rev-parse", "--verify", branch)
	assert.Error(env.t, err, "Branch %s exists but should not", branch)
}

// GetTag gets all tags pointing to the current HEAD
func (env *GitTestEnv) GetTag() string {
	env.t.Helper()
	return strings.TrimSpace(env.ExecuteGit("tag", "--points-at", "HEAD"))
}

// GetCommitMessage gets the message of a specific commit
func (env *GitTestEnv) GetCommitMessage(commit string) string {
	env.t.Helper()
	output := env.ExecuteGit("log", "-1", "--pretty=%B", commit)
	return strings.TrimSpace(output)
}

// CountCommitsBetween counts the number of commits between two refs
func (env *GitTestEnv) CountCommitsBetween(base, head string) int {
	env.t.Helper()
	output := env.ExecuteGit("rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	var count int
	_, err := fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	require.NoError(env.t, err, "Failed to parse commit count")
	return count
}

// GetCurrentBranch gets the name of the current branch
func (env *GitTestEnv) GetCurrentBranch() string {
	env.t.Helper()
	output := env.ExecuteGit("rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(output)
}

// AssertFileInBranchEquals checks if a file in a branch has the expected content
func (env *GitTestEnv) AssertFileInBranchEquals(branch, path, expectedContent string) {
	env.t.Helper()
	content := env.GetFileContentFromCommit(branch, path)
	assert.Equal(env.t, expectedContent, content,
		"File %s in branch %s has unexpected content", path, branch)
}

// AssertCommitsAhead checks if a branch is exactly N commits ahead of another branch
func (env *GitTestEnv) AssertCommitsAhead(branch, baseBranch string, expectedCount int) {
	env.t.Helper()
	count := env.CountCommitsBetween(baseBranch, branch)
	assert.Equal(env.t, expectedCount, count,
		"Branch %s should be %d commits ahead of %s, but is %d commits ahead",
		branch, expectedCount, baseBranch, count)
}
