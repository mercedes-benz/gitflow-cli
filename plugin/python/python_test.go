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
	t.Run("Only_PyprojectToml", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		original := core.ProjectPath
		core.ProjectPath = tmpDir
		defer func() { core.ProjectPath = original }()

		testPlugin := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
		core.CheckVersionFile(testPlugin)

		if testPlugin.VersionFileName() != "pyproject.toml" {
			t.Fatalf("Expected 'pyproject.toml', got '%s'", testPlugin.VersionFileName())
		}
	})

	t.Run("Only_SetupPy", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "setup.py"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		original := core.ProjectPath
		core.ProjectPath = tmpDir
		defer func() { core.ProjectPath = original }()

		testPlugin := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
		core.CheckVersionFile(testPlugin)

		if testPlugin.VersionFileName() != "setup.py" {
			t.Fatalf("Expected 'setup.py', got '%s'", testPlugin.VersionFileName())
		}
	})

	t.Run("Both_PyprojectToml_Priority", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "setup.py"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		original := core.ProjectPath
		core.ProjectPath = tmpDir
		defer func() { core.ProjectPath = original }()

		testPlugin := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
		core.CheckVersionFile(testPlugin)

		if testPlugin.VersionFileName() != "pyproject.toml" {
			t.Fatalf("Expected 'pyproject.toml', got '%s'", testPlugin.VersionFileName())
		}
	})
}
