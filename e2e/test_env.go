/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"text/template"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExecuteFunc is the function used to execute the CLI command.
// Must be set before calling ExecuteGitflow (typically to cmd.Execute).
var ExecuteFunc func() error

// GitTestEnv manages local repository and simulated remote repository
type GitTestEnv struct {
	LocalPath  string // Path to local repository
	RemotePath string // Path to simulated remote repository
	t          *testing.T
	dockerMode bool
}

// SetupTestEnvOption configures options for SetupTestEnv
type SetupTestEnvOption func(*testEnvOptions)

// TestEnvOption functions to customize test environment setup
var (
	WithProductionBranch = func(branch string) SetupTestEnvOption {
		return func(opts *testEnvOptions) { opts.productionBranch = branch }
	}
	WithDevelopmentBranch = func(branch string) SetupTestEnvOption {
		return func(opts *testEnvOptions) { opts.developmentBranch = branch }
	}
	WithReleaseBranch = func(prefix string) SetupTestEnvOption {
		return func(opts *testEnvOptions) { opts.releaseBranch = prefix }
	}
	WithHotfixBranch = func(prefix string) SetupTestEnvOption {
		return func(opts *testEnvOptions) { opts.hotfixBranch = prefix }
	}
	WithDockerMode = func(hasImage bool) SetupTestEnvOption {
		return func(opts *testEnvOptions) {
			opts.dockerMode = hasImage && os.Getenv("GITFLOW_TEST_MODE") != "native"
		}
	}
)

// testEnvOptions holds the options for setting up the test environment
type testEnvOptions struct {
	productionBranch  string
	developmentBranch string
	releaseBranch     string
	hotfixBranch      string
	dockerMode        bool
}

// SetupTestEnv creates test environment with local repo and simulated remote
func SetupTestEnv(t *testing.T, options ...SetupTestEnvOption) *GitTestEnv {
	t.Helper()

	// Default options
	opts := &testEnvOptions{
		productionBranch:  "main",
		developmentBranch: "develop",
	}

	// Apply user options
	for _, option := range options {
		option(opts)
	}

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

	// Initialize local repository with production branch
	cmd = exec.Command("git", "init", "--initial-branch="+opts.productionBranch)
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
		dockerMode: opts.dockerMode,
	}

	if opts.dockerMode {
		t.Cleanup(func() { plugin.ExecutorModeOverride = "" })
	}

	// Create an empty commit to initialize the production branch
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial empty commit")
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to create initial empty commit")

	// Push the empty production branch to remote
	cmd = exec.Command("git", "push", "-u", "origin", opts.productionBranch)
	cmd.Dir = localPath
	require.NoError(t, cmd.Run(), "Failed to push production branch")

	// Create development branch
	env.CreateBranch(opts.developmentBranch, opts.productionBranch)

	return env
}

// CommitTemplateContent renders a template string with the given version and commits the result.
func (env *GitTestEnv) CommitTemplateContent(templateContent, fileName, version, commitRef string) {
	env.t.Helper()

	tmpl, err := template.New(fileName).Parse(templateContent)
	require.NoError(env.t, err, "Failed to parse template for %s", fileName)

	data := struct{ Version string }{Version: version}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	require.NoError(env.t, err, "Failed to render template for %s", fileName)

	env.CommitFile(fileName, buf.Bytes(), commitRef)
}

// CommitFile creates a file with the specified content in the repository,
// commits it, and pushes it to the remote.
func (env *GitTestEnv) CommitFile(name string, content []byte, commitRef string) {
	env.t.Helper()

	env.ExecuteGit("checkout", commitRef)

	// Create file with content
	path := filepath.Join(env.LocalPath, name)
	err := os.WriteFile(path, content, 0644)
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

	// Save the original os.Args and restore it when done
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set command line arguments with the --path parameter
	baseArgs := []string{"gitflow-cli", "--path", env.LocalPath}
	if env.dockerMode {
		baseArgs = append(baseArgs, "--docker-mode")
	}
	os.Args = append(baseArgs, args...)
	env.t.Logf("Executing command: gitflow-cli %s", strings.Join(os.Args[1:], " "))

	// Capture output using a pipe
	r, w, err := os.Pipe()
	require.NoError(env.t, err)

	// Save original stdout/stderr and replace with pipe
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	// Start background reader BEFORE execution to prevent pipe deadlock.
	// OS pipes have a limited kernel buffer (~64KB); heavy output (e.g. Maven
	// dependency downloads) fills it and blocks writes if nobody is reading.
	var output []byte
	var readErr error
	done := make(chan struct{})
	go func() {
		output, readErr = io.ReadAll(r)
		close(done)
	}()

	// Recover from any panics that might occur during command execution
	var cmdErr error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				cmdErr = fmt.Errorf("panic during command execution: %v", rec)
				env.t.Logf("PANIC: %v", rec)
				debug.PrintStack()
			}
		}()

		cmdErr = ExecuteFunc()
	}()

	// Restore original stdout/stderr and close the write end to signal EOF to reader
	os.Stdout, os.Stderr = oldStdout, oldStderr
	w.Close()

	// Wait for reader goroutine to finish
	<-done
	require.NoError(env.t, readErr)

	// Log the command output and any errors
	if cmdErr != nil {
		env.t.Logf("Command failed: %v", cmdErr)
	}
	env.t.Logf("Command output for 'gitflow-cli %s':\n%s", strings.Join(args, " "), string(output))

	// If there was an error, fail the test with more information
	if cmdErr != nil {
		env.t.Fatalf("Command failed: %v\nOutput: %s", cmdErr, string(output))
	}

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

// AssertCommitMessageEquals checks if the first line of the commit message at the given branch and depth matches the expected message
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
func (env *GitTestEnv) AssertTagEquals(expectedTag, commitRef string, depth ...int) {
	env.t.Helper()

	depthValue := 0
	if len(depth) > 0 && depth[0] > 0 {
		depthValue = depth[0]
	}

	actualTag := env.getTag(commitRef, depthValue)
	assert.Equal(env.t, expectedTag, actualTag, "Tag of %s~%d should be '%s' but was '%s'", commitRef, depthValue, expectedTag, actualTag)
}

// AssertTemplateVersionEquals checks if the version in a file matches the expected version
// using inline template content instead of a file path.
func (env *GitTestEnv) AssertTemplateVersionEquals(templateContent, fileName, expectedVersion, commitRef string, depth ...int) {
	env.t.Helper()

	if len(depth) > 0 && depth[0] > 0 {
		commitRef = fmt.Sprintf("%s~%d", commitRef, depth[0])
	}

	// Simple template: content is just "{{.Version}}"
	if strings.TrimSpace(templateContent) == "{{.Version}}" {
		actual := strings.TrimSpace(env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, fileName)))
		assert.Equal(env.t, expectedVersion, actual,
			"Version in %s at %s should be '%s' but was '%s'", fileName, commitRef, expectedVersion, actual)
		return
	}

	// Complex template: use marker replacement to locate version
	parsedTemplate, err := template.New(fileName).Parse(templateContent)
	require.NoError(env.t, err, "Failed to parse template")

	versionMarker := "###VERSION_MARKER###"
	var markerContent bytes.Buffer
	err = parsedTemplate.Execute(&markerContent, struct{ Version string }{Version: versionMarker})
	require.NoError(env.t, err, "Failed to render template with marker")

	markerOutput := markerContent.String()
	markerPos := strings.Index(markerOutput, versionMarker)
	require.True(env.t, markerPos >= 0, "Could not find version marker in rendered template")

	prefix := markerOutput[:markerPos]
	suffix := markerOutput[markerPos+len(versionMarker):]

	actualFileContent := env.ExecuteGit("show", fmt.Sprintf("%s:%s", commitRef, fileName))

	prefixPos := strings.Index(actualFileContent, prefix)
	require.True(env.t, prefixPos >= 0, "Could not find content before version in actual file")

	startPos := prefixPos + len(prefix)
	var endPos int
	if suffix != "" {
		suffixPos := strings.Index(actualFileContent[startPos:], suffix)
		require.True(env.t, suffixPos >= 0, "Could not find content after version in actual file")
		endPos = startPos + suffixPos
	} else {
		endPos = len(actualFileContent)
	}

	actualVersion := strings.TrimSpace(actualFileContent[startPos:endPos])
	assert.Equal(env.t, expectedVersion, actualVersion,
		"Version in %s at %s should be '%s' but was '%s'", fileName, commitRef, expectedVersion, actualVersion)
}

// getCommitMessage gets the message of a specific commit
func (env *GitTestEnv) getCommitMessage(commitRef string, depth ...int) string {
	env.t.Helper()

	commitOffset := "HEAD"
	if len(depth) > 0 && depth[0] > 0 {
		commitOffset = fmt.Sprintf("HEAD~%d", depth[0])
	}

	args := []string{"log", "-1", "--pretty=%B"}
	if commitRef != "" {
		if len(depth) > 0 && depth[0] > 0 {
			args = append(args, fmt.Sprintf("%s~%d", commitRef, depth[0]))
		} else {
			args = append(args, commitRef)
		}
	} else {
		args = append(args, commitOffset)
	}

	output := env.ExecuteGit(args...)
	return strings.TrimSpace(output)
}

// getTag gets all tags pointing to a specific commit
func (env *GitTestEnv) getTag(commit string, depth ...int) string {
	env.t.Helper()

	commitRef := "HEAD"
	if commit != "" {
		commitRef = commit
	}

	if len(depth) > 0 && depth[0] > 0 {
		if commit != "" {
			commitRef = fmt.Sprintf("%s~%d", commit, depth[0])
		} else {
			commitRef = fmt.Sprintf("HEAD~%d", depth[0])
		}
	}

	return strings.TrimSpace(env.ExecuteGit("tag", "--points-at", commitRef))
}
