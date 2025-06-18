/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package road

import (
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to set up test environment
func setupTest(t *testing.T, content string) (string, core.Repository, *roadPlugin) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create test file with content
	testFilePath := filepath.Join(tempDir, "road.yaml")
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to write test file")

	// Create repository using core.NewRepository
	repository := core.NewRepository(tempDir, "")

	// Create roadPlugin instance
	roadPlugin := &roadPlugin{
		Plugin: plugin.NewFactory().NewPlugin(pluginConfig),
	}

	return testFilePath, repository, roadPlugin
}

func TestVersionReadWrite(t *testing.T) {

	testCases := []struct {
		name           string
		initialContent string
		expectedResult string
	}{
		{
			name:           "No quotes",
			initialContent: "versionNumber: 1.2.3",
			expectedResult: "versionNumber: 1.2.3-dev",
		},
		{
			name:           "Single quotes",
			initialContent: "versionNumber: '1.2.3'",
			expectedResult: "versionNumber: '1.2.3-dev'",
		},
		{
			name:           "Double quotes",
			initialContent: "versionNumber: \"1.2.3\"",
			expectedResult: "versionNumber: \"1.2.3-dev\"",
		},
		{
			name:           "With spaces",
			initialContent: "versionNumber:    1.2.3   ",
			expectedResult: "versionNumber: 1.2.3-dev",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(test *testing.T) {
			// Set up test environment using helper function
			testFilePath, repository, roadPlugin := setupTest(test, testCase.initialContent)

			// Read version
			originalVersion, err := roadPlugin.ReadVersion(repository)
			require.NoError(test, err, "ReadVersion failed")

			// Add dev qualifier to the original version
			originalVersion.Qualifier = "dev"

			// Write back the version with dev qualifier
			err = roadPlugin.WriteVersion(repository, originalVersion)
			require.NoError(test, err, "WriteVersion failed")

			// Read the resulting file content
			resultBytes, err := os.ReadFile(testFilePath)
			require.NoError(test, err, "Failed to read test file after write")

			// Compare with expected result using assert
			resultContent := string(resultBytes)
			assert.Equal(test, testCase.expectedResult, resultContent, "Version replacement did not produce expected content")
		})
	}
}

// TestVersionNoMatch tests cases where the version is not recognized
func TestVersionNoMatch(t *testing.T) {
	// Test cases with different non-matching formats
	testCases := []struct {
		name           string
		initialContent string
		errorExpected  bool
		errorMessage   string
	}{
		{
			name:           "No versionNumber node",
			initialContent: "otherKey: 1.2.3",
		},
		{
			name:           "versionNumber with leading whitespace",
			initialContent: " versionNumber: 1.2.3",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(test *testing.T) {
			// Set up test environment using helper function
			_, repository, roadPlugin := setupTest(test, testCase.initialContent)

			// Call ReadVersion and check the result
			_, err := roadPlugin.ReadVersion(repository)

			// If an error is expected
			require.Error(test, err, "ReadVersion should fail for this case")

		})
	}
}
