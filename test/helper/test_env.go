/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package helper

import (
	"bytes"
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
	"text/template"

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

	// Initialize local repository with main branch
	cmd = exec.Command("git", "init", "--initial-branch=main")
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

	// Create git testing environment
	env := &GitTestEnv{
		LocalPath:  localPath,
		RemotePath: remotePath,
		t:          t,
	}

	// Create an empty commit to initialize the main branch
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial empty commit")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to create initial empty commit")

	// Push the empty main branch to remote
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to push main branch")

	// Create develop branch
	env.CreateBranch("develop", "main")

	return env
}

// CommitFileFromTemplate creates a file using a template with variables, adds it, commits it, and pushes it to the remote
// The filename will be derived from the template name (e.g., template "version.txt.tpl" creates file "version.txt")
// The commit message is automatically generated based on the branch name
func (env *GitTestEnv) CommitFileFromTemplate(templatePath, bindingValue, commitRef string) {
	env.t.Helper()

	env.ExecuteGit("checkout", commitRef)

	// Read the template file
	templateContent, err := os.ReadFile(templatePath)
	require.NoError(env.t, err, "Failed to read template file: %s", templatePath)

	// Create a new template and parse the content
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(templateContent))
	require.NoError(env.t, err, "Failed to parse template: %s", templatePath)

	// Prepare the data for template rendering
	data := struct {
		Version string
	}{
		Version: bindingValue,
	}

	// Render the template
	var renderedContent bytes.Buffer
	err = tmpl.Execute(&renderedContent, data)
	require.NoError(env.t, err, "Failed to render template: %s", templatePath)

	// Derive the filename from the template name by removing the .tpl extension
	templateBase := filepath.Base(templatePath)
	file := strings.TrimSuffix(templateBase, ".tpl")

	// Write the file
	path := filepath.Join(env.LocalPath, file)
	err = os.WriteFile(path, renderedContent.Bytes(), 0644)
	require.NoError(env.t, err, "Failed to create file: %s", path)

	// Generate commit message based on branch name
	message := fmt.Sprintf("Set up test precondition for %s branch", commitRef)

	env.ExecuteGit("add", path)
	env.ExecuteGit("commit", "-m", message)
	env.ExecuteGit("push", "-u", "origin", commitRef)
}

// CreateBranch creates a new branch from the specified base branch
func (env *GitTestEnv) CreateBranch(branch string, commitRef ...string) {
	env.t.Helper()

	// Create from current HEAD by default
	baseRef := "HEAD"
	if len(commitRef) > 0 && commitRef[0] != "" {
		// If specified, create from the base branch
		baseRef = commitRef[0]
	}

	// Checkout the base branch or commit
	env.ExecuteGit("checkout", baseRef)

	// Create and checkout the new branch
	env.ExecuteGit("checkout", "-b", branch)

	// Push to remote and set up tracking
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
func (env *GitTestEnv) AssertBranchExists(commitRef string) {
	env.t.Helper()
	_, err := env.ExecuteGitAllowError("rev-parse", "--verify", commitRef)
	assert.NoError(env.t, err, "Branch %s does not exist", commitRef)
}

// AssertBranchDoesNotExist checks that a branch does not exist
func (env *GitTestEnv) AssertBranchDoesNotExist(commitRef string) {
	env.t.Helper()
	_, err := env.ExecuteGitAllowError("rev-parse", "--verify", commitRef)
	assert.Error(env.t, err, "Branch %s exists but should not", commitRef)
}

// AssertCurrentBranchEquals checks if the currently checked out branch matches the expected branch name
func (env *GitTestEnv) AssertCurrentBranchEquals(expectedBranch string) {
	env.t.Helper()
	currentBranch := strings.TrimSpace(env.ExecuteGit("rev-parse", "--abbrev-ref", "HEAD"))
	assert.Equal(env.t, expectedBranch, currentBranch, "Current branch should be '%s', but got '%s'", expectedBranch, currentBranch)
}

// AssertFileEquals checks if a file in a branch has the expected content
// If comparing a template file ending with .tpl, it will check the {{.Version}} placeholder against expectedContent
// depth specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertFileEquals(path, expectedContent, commitRef string, depth ...int) {
	env.t.Helper()

	if len(depth) > 0 && depth[0] > 0 {
		commitRef = fmt.Sprintf("%s~%d", commitRef, depth[0])
	}

	fileContent := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, path))

	// Check if this is a template file (.tpl extension)
	if strings.HasSuffix(path, ".tpl") {
		// If it's a template file, parse it to extract the version placeholder
		tmpl, err := template.New("test").Parse(fileContent)
		require.NoError(env.t, err, "Failed to parse template file: %s", path)

		// Create a buffer to render the template with the expected content
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, struct {
			Version string
		}{
			Version: expectedContent,
		})
		require.NoError(env.t, err, "Failed to render template with version: %s", expectedContent)

		// Get the actual file name (without .tpl extension)
		actualFileName := strings.TrimSuffix(filepath.Base(path), ".tpl")

		// Get the actual file content from the repository
		actualFileContent := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, actualFileName))

		// Compare the rendered template with the actual file content
		assert.Equal(env.t, buf.String(), actualFileContent,
			"File %s in %s with version %s has unexpected content",
			actualFileName, commitRef, expectedContent)
	} else {
		// For non-template files, do a direct comparison as before
		assert.Equal(env.t, expectedContent, fileContent,
			"File %s in %s has unexpected content", path, commitRef)
	}
}

// AssertCommitMessageEquals checks if the first line of the commit message at the given branch and depth matches the expected message
// depth specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertCommitMessageEquals(expectedMessage, commitRef string, depth ...int) {
	env.t.Helper()

	depthValue := 0
	if len(depth) > 0 && depth[0] > 0 {
		depthValue = depth[0]
	}

	actualMessage := env.getCommitMessage(commitRef, depthValue)

	// Only compare the first line of the commit message
	firstLine := strings.Split(actualMessage, "\n")[0]

	assert.Equal(env.t, expectedMessage, firstLine, "Commit message of %s~%d should be '%s' but was '%s'", commitRef, depthValue, expectedMessage, firstLine)
}

// AssertTagEquals checks if the tag at the given branch and depth matches the expected tag
// depth specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) AssertTagEquals(expectedTag, commitRef string, depth ...int) {
	env.t.Helper()

	depthValue := 0
	if len(depth) > 0 && depth[0] > 0 {
		depthValue = depth[0]
	}

	actualTag := env.getTag(commitRef, depthValue)
	assert.Equal(env.t, expectedTag, actualTag, "Tag of %s~%d should be '%s' but was '%s'", commitRef, depthValue, expectedTag, actualTag)
}

// GetCommitMessage gets the message of a specific commit
// depth specifies which commit to retrieve:
// 0 = HEAD (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) getCommitMessage(commitRef string, depth ...int) string {
	env.t.Helper()

	commitOffset := "HEAD"
	if len(depth) > 0 && depth[0] > 0 {
		// If depth is provided and > 0, get older commits
		commitOffset = fmt.Sprintf("HEAD~%d", depth[0])
	}

	args := []string{"log", "-1", "--pretty=%B"}
	if commitRef != "" {
		// If commitRef is specified, use it as the base reference
		if len(depth) > 0 && depth[0] > 0 {
			// For a specific commitRef with offset
			args = append(args, fmt.Sprintf("%s~%d", commitRef, depth[0]))
		} else {
			// For the commitRef itself
			args = append(args, commitRef)
		}
	} else {
		// If no commitRef is specified, use the HEAD with potential offset
		args = append(args, commitOffset)
	}

	output := env.ExecuteGit(args...)
	return strings.TrimSpace(output)
}

// GetTag gets all tags pointing to a specific commit
// depth specifies which commit to retrieve:
// 0 = HEAD or specified commit (latest), 1 = HEAD~1 (previous commit), etc.
func (env *GitTestEnv) getTag(commit string, depth ...int) string {
	env.t.Helper()

	commitRef := "HEAD"
	if commit != "" {
		commitRef = commit
	}

	if len(depth) > 0 && depth[0] > 0 {
		// If depth is provided and > 0, get older commits
		if commit != "" {
			commitRef = fmt.Sprintf("%s~%d", commit, depth[0])
		} else {
			commitRef = fmt.Sprintf("HEAD~%d", depth[0])
		}
	}

	return strings.TrimSpace(env.ExecuteGit("tag", "--points-at", commitRef))
}

// AssertVersionEquals checks if the version in a file in a branch matches the expected version
// It uses the template file to determine how to extract the version from the actual file
// templatePath is the path to the template file (e.g., "version.txt.tpl")
// The actual file name is derived from the template name (e.g., "version.txt")
func (env *GitTestEnv) AssertVersionEquals(templatePath, expectedVersion, commitRef string, depth ...int) {
	env.t.Helper()

	// Derive the actual filename from the template name by removing the .tpl extension
	templateBase := filepath.Base(templatePath)
	versionFileName := strings.TrimSuffix(templateBase, ".tpl")

	// Get the commit reference with optional depth
	if len(depth) > 0 && depth[0] > 0 {
		commitRef = fmt.Sprintf("%s~%d", commitRef, depth[0])
	}

	// Read the template file to understand how to extract the version
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		assert.Fail(env.t, "Failed to read template file: %s: %v", templatePath, err)
		return
	}

	// Get the actual file content from the specified branch/commit
	versionFileContent := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, versionFileName))

	// Parse the template to understand its structure
	tmpl, err := template.New(templateBase).Parse(string(templateContent))
	if err != nil {
		assert.Fail(env.t, "Failed to parse template file: %s: %v", templatePath, err)
		return
	}

	// Create a mock execution of the template with a unique marker for Version
	// This will help us understand where to look for the version in the actual file
	uniqueMarker := "##VERSION_PLACEHOLDER_UNIQUE##"
	var mockOutput bytes.Buffer
	err = tmpl.Execute(&mockOutput, struct {
		Version string
	}{
		Version: uniqueMarker,
	})
	if err != nil {
		assert.Fail(env.t, "Failed to execute template: %v", err)
		return
	}

	mockResult := mockOutput.String()

	// Locate the unique marker in the mock output
	markerIndex := strings.Index(mockResult, uniqueMarker)
	if markerIndex < 0 {
		assert.Fail(env.t, "Could not locate Version placeholder in template: %s", templatePath)
		return
	}

	// Try to determine the actual version from the file content
	// by comparing the structure before and after the marker
	beforeMarker := mockResult[:markerIndex]
	afterMarker := mockResult[markerIndex+len(uniqueMarker):]

	// Find where the version should be in the actual file
	startIndex := strings.Index(versionFileContent, beforeMarker)
	if startIndex < 0 {
		assert.Fail(env.t, "Could not locate start of version in file %s in %s", versionFileName, commitRef)
		return
	}
	startIndex += len(beforeMarker)

	endIndex := -1
	if afterMarker != "" {
		endIndex = strings.Index(versionFileContent[startIndex:], afterMarker)
		if endIndex < 0 {
			assert.Fail(env.t, "Could not locate end of version in file %s in %s", versionFileName, commitRef)
			return
		}
		endIndex += startIndex
	} else {
		endIndex = len(versionFileContent)
	}

	actualVersion := versionFileContent[startIndex:endIndex]
	assert.Equal(env.t, expectedVersion, actualVersion,
		"Version in %s in %s should be '%s' but was '%s'", versionFileName, commitRef, expectedVersion, actualVersion)
}
