/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

func TestVersionFileSelection(t *testing.T) {
	t.Run("OnlyPyprojectTomlFileExists", func(t *testing.T) {
		testVersionFile(t, []string{"pyproject.toml"}, "pyproject.toml")
	})

	t.Run("OnlySetupPyFileExists", func(t *testing.T) {
		testVersionFile(t, []string{"setup.py"}, "setup.py")
	})

	t.Run("BothPyprojectTomlAndSetupPyFilesExist", func(t *testing.T) {
		testVersionFile(t, []string{"pyproject.toml", "setup.py"}, "pyproject.toml")
	})
}

func testVersionFile(t *testing.T, files []string, expected string) {
	tmpDir := t.TempDir()
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	original := core.ProjectPath
	core.ProjectPath = tmpDir
	defer func() { core.ProjectPath = original }()

	testPlugin := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
	core.CheckVersionFile(testPlugin)

	if testPlugin.VersionFileName() != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, testPlugin.VersionFileName())
	}
}
