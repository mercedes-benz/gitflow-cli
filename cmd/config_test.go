/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDefaultConfig_CreatesFileWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir) // Windows compatibility

	err := initDefaultConfig()
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, ".gitflow-cli.yaml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "production: main")
	assert.Contains(t, string(content), "development: develop")
	assert.Contains(t, string(content), "release: release")
	assert.Contains(t, string(content), "hotfix: hotfix")
	assert.Contains(t, string(content), "push: true")
	assert.Contains(t, string(content), "rollback: false")
	assert.Contains(t, string(content), "docker-fallback: true")
	assert.Contains(t, string(content), "logging: \"off\"")
}

func TestInitDefaultConfig_DoesNotOverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	configPath := filepath.Join(tmpDir, ".gitflow-cli.yaml")
	existingContent := []byte("core:\n  production: master\n")
	require.NoError(t, os.WriteFile(configPath, existingContent, 0644))

	err := initDefaultConfig()
	require.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, content)
}
