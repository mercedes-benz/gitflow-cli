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

	// create initial commit
	env.CommitFile("main", "README.md", "# Gitflow Test Repository", "initial commit")

	return env
}

// CommitFile creates a file with given content, adds it, commits it, and pushes it to the remote
func (env *GitTestEnv) CommitFile(branch, file, content, message string) {
	env.t.Helper()

	path := filepath.Join(env.LocalPath, file)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(env.t, err, "Failed to create file: %s", path)

	// First check if the branch exists locally
	_, err = env.ExecuteGitAllowError("rev-parse", "--verify", branch)
	if err != nil {
		env.ExecuteGit("checkout", "-b", branch)
	} else {
		env.ExecuteGit("checkout", branch)
	}

	env.ExecuteGit("add", path)
	env.ExecuteGit("commit", "-m", message)
	env.ExecuteGit("push", "-u", "origin", branch)
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

// GetCurrentBranch gets the name of the current branch
func (env *GitTestEnv) GetCurrentBranch() string {
	env.t.Helper()
	output := env.ExecuteGit("rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(output)
}

// AssertFileEquals checks if a file in a branch has the expected content
// index specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertFileEquals(path, expectedContent, branch string, index ...int) {
	env.t.Helper()

	commitRef := branch
	if len(index) > 0 && index[0] > 0 {
		commitRef = fmt.Sprintf("%s~%d", branch, index[0])
	}

	fileContent := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, path))
	assert.Equal(env.t, expectedContent, fileContent, "File %s in %s has unexpected content", path, commitRef)
}

// AssertCommitMessageEquals checks if the commit message at the given branch and index matches the expected message
// index specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertCommitMessageEquals(expectedMessage, branch string, index ...int) {
	env.t.Helper()

	indexValue := 0
	if len(index) > 0 && index[0] > 0 {
		indexValue = index[0]
	}

	actualMessage := env.getCommitMessage(branch, indexValue)
	assert.Equal(env.t, expectedMessage, actualMessage, "Commit message of %s~%d should be '%s' but was '%s'", branch, indexValue, expectedMessage, actualMessage)
}

// AssertTagEquals checks if the tag at the given branch and index matches the expected tag
// index specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertTagEquals(expectedTag, branch string, index ...int) {
	env.t.Helper()

	indexValue := 0
	if len(index) > 0 && index[0] > 0 {
		indexValue = index[0]
	}

	actualTag := env.getTag(branch, indexValue)
	assert.Equal(env.t, expectedTag, actualTag, "Tag of %s~%d should be '%s' but was '%s'", branch, indexValue, expectedTag, actualTag)
}

// GetCommitMessage gets the message of a specific commit
// index specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) getCommitMessage(commit string, index ...int) string {
	env.t.Helper()

	commitOffset := "HEAD"
	if len(index) > 0 && index[0] > 0 {
		// If index is provided and > 0, get older commits
		commitOffset = fmt.Sprintf("HEAD~%d", index[0])
	}

	args := []string{"log", "-1", "--pretty=%B"}
	if commit != "" {
		// If commit is specified, use it as the base reference
		if len(index) > 0 && index[0] > 0 {
			// For a specific commit with offset
			args = append(args, fmt.Sprintf("%s~%d", commit, index[0]))
		} else {
			// For the commit itself
			args = append(args, commit)
		}
	} else {
		// If no commit is specified, use the HEAD with potential offset
		args = append(args, commitOffset)
	}

	output := env.ExecuteGit(args...)
	return strings.TrimSpace(output)
}

// GetTag gets all tags pointing to a specific commit
// index specifies which commit to retrieve:
// 0 = HEAD or specified commit (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) getTag(commit string, index ...int) string {
	env.t.Helper()

	commitRef := "HEAD"
	if commit != "" {
		commitRef = commit
	}

	if len(index) > 0 && index[0] > 0 {
		// If index is provided and > 0, get older commits
		if commit != "" {
			commitRef = fmt.Sprintf("%s~%d", commit, index[0])
		} else {
			commitRef = fmt.Sprintf("HEAD~%d", index[0])
		}
	}

	return strings.TrimSpace(env.ExecuteGit("tag", "--points-at", commitRef))
}
