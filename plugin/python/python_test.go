/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useDockerMode(t *testing.T) {
	t.Helper()
	mode := os.Getenv("GITFLOW_TEST_MODE")
	if mode == "" {
		mode = plugin.ModeDocker
	}
	if mode != plugin.ModeDocker {
		return
	}
	plugin.ExecutorModeOverride = plugin.ModeDocker
	t.Cleanup(func() { plugin.ExecutorModeOverride = "" })
}


//go:embed testdata/e2e/pyproject_pep621.toml.tpl
var pyprojectTemplate string

//go:embed testdata/e2e/pyproject_poetry.toml.tpl
var poetryTemplate string

//go:embed testdata/e2e/setup.cfg.tpl
var setupCfgTemplate string

//go:embed testdata/e2e/setup.py.tpl
var setupPyTemplate string

var testConfigs = []plugin.TestConfig{
	{
		Name:             "python_pyproject",
		PluginName:       "python",
		DockerImage:      pluginConfig.DockerImage,
		VersionQualifier: "dev",
		VersionFileName:  "pyproject.toml",
		Template:         pyprojectTemplate,
		EmptyContent:     []byte{},
	},
	{
		Name:             "python_poetry",
		PluginName:       "python",
		DockerImage:      pluginConfig.DockerImage,
		VersionQualifier: "dev",
		VersionFileName:  "pyproject.toml",
		Template:         poetryTemplate,
	},
	{
		Name:             "python_setup_cfg",
		PluginName:       "python",
		DockerImage:      pluginConfig.DockerImage,
		VersionQualifier: "dev",
		VersionFileName:  "setup.cfg",
		Template:         setupCfgTemplate,
		EmptyContent:     []byte{},
	},
	{
		Name:             "python_setup_py",
		PluginName:       "python",
		DockerImage:      pluginConfig.DockerImage,
		VersionQualifier: "dev",
		VersionFileName:  "setup.py",
		Template:         setupPyTemplate,
		EmptyContent:     []byte{},
	},
}

func TestE2E_ReleaseStart(t *testing.T) {
	for _, tc := range testConfigs {
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunReleaseStart(t, tc)
		})
	}
}

func TestE2E_ReleaseStart_BeforeHook(t *testing.T) {
	for _, tc := range testConfigs {
		if tc.EmptyContent == nil {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunBeforeReleaseStartHook(t, tc)
		})
	}
}

func TestE2E_ReleaseFinish(t *testing.T) {
	for _, tc := range testConfigs {
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunReleaseFinish(t, tc)
		})
	}
}

func TestE2E_HotfixStart(t *testing.T) {
	for _, tc := range testConfigs {
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunHotfixStart(t, tc)
		})
	}
}

func TestE2E_HotfixStart_BeforeHook(t *testing.T) {
	for _, tc := range testConfigs {
		if tc.EmptyContent == nil {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunBeforeHotfixStartHook(t, tc)
		})
	}
}

func TestE2E_HotfixFinish(t *testing.T) {
	for _, tc := range testConfigs {
		t.Run(tc.Name, func(t *testing.T) {
			workflow.RunHotfixFinish(t, tc)
		})
	}
}

// setupFromTestdata copies a fixture file into a temp dir with the given target name.
func setupFromTestdata(t *testing.T, fixture, targetFileName string) (core.Repository, *pythonPlugin) {
	t.Helper()
	tmpDir := t.TempDir()

	content, err := os.ReadFile(filepath.Join("testdata", "unit", fixture))
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, targetFileName), content, 0644))

	p := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
	p.Config.VersionFileName = targetFileName

	return core.NewRepository(tmpDir, ""), p
}

// setupEmpty creates an empty target file in a temp dir.
func setupEmpty(t *testing.T, targetFileName string) (core.Repository, *pythonPlugin) {
	t.Helper()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, targetFileName), []byte(""), 0644))

	p := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
	p.Config.VersionFileName = targetFileName

	return core.NewRepository(tmpDir, ""), p
}

// TestVersionFileSelection tests correct priority: pyproject.toml > setup.cfg > setup.py
func TestVersionFileSelection(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"OnlyPyprojectToml", []string{"pyproject.toml"}, "pyproject.toml"},
		{"OnlySetupCfg", []string{"setup.cfg"}, "setup.cfg"},
		{"OnlySetupPy", []string{"setup.py"}, "setup.py"},
		{"PyprojectTomlHasHighestPriority", []string{"pyproject.toml", "setup.cfg", "setup.py"}, "pyproject.toml"},
		{"SetupCfgBeforeSetupPy", []string{"setup.cfg", "setup.py"}, "setup.cfg"},
		{"PyprojectTomlBeforeSetupCfg", []string{"pyproject.toml", "setup.cfg"}, "pyproject.toml"},
		{"PyprojectTomlBeforeSetupPy", []string{"pyproject.toml", "setup.py"}, "pyproject.toml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, file := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, file), []byte(""), 0644))
			}

			original := core.ProjectPath
			core.ProjectPath = tmpDir
			defer func() { core.ProjectPath = original }()

			p := &pythonPlugin{Plugin: plugin.NewFactory().NewPlugin(pluginConfig)}
			core.CheckVersionFile(p)

			assert.Equal(t, tt.expected, p.VersionFileName())
		})
	}
}

// TestReadVersion_PyprojectPEP621 tests reading version from PEP 621 pyproject.toml
func TestReadVersion_PyprojectPEP621(t *testing.T) {
	useDockerMode(t)
	t.Run("StandardVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621.toml", "pyproject.toml")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", v.String())
	})

	t.Run("VersionWithQualifier", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621_qualifier.toml", "pyproject.toml")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "2.0.0-dev", v.String())
	})

	t.Run("EmptyFile_ReturnsError", func(t *testing.T) {
		repo, p := setupEmpty(t, "pyproject.toml")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("NoVersionField_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621_no_version.toml", "pyproject.toml")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("InvalidVersion_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621_invalid_version.toml", "pyproject.toml")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})
}

// TestReadVersion_PyprojectPoetry tests reading version from Poetry pyproject.toml
func TestReadVersion_PyprojectPoetry(t *testing.T) {
	useDockerMode(t)
	t.Run("StandardVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_poetry.toml", "pyproject.toml")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", v.String())
	})

	t.Run("VersionWithQualifier", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_poetry_qualifier.toml", "pyproject.toml")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "5.2.1-dev", v.String())
	})

	t.Run("NoVersionField_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_poetry_no_version.toml", "pyproject.toml")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})
}

// TestReadVersion_SetupCfg tests reading version from setup.cfg
func TestReadVersion_SetupCfg(t *testing.T) {
	useDockerMode(t)
	t.Run("StandardVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup.cfg", "setup.cfg")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", v.String())
	})

	t.Run("VersionWithQualifier", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_cfg_qualifier.cfg", "setup.cfg")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "4.0.0-dev", v.String())
	})

	t.Run("EmptyFile_ReturnsError", func(t *testing.T) {
		repo, p := setupEmpty(t, "setup.cfg")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("NoMetadataSection_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_cfg_no_metadata.cfg", "setup.cfg")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("NoVersionInMetadata_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_cfg_no_version.cfg", "setup.cfg")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("InvalidVersion_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_cfg_invalid_version.cfg", "setup.cfg")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})
}

// TestReadVersion_SetupPy tests reading version from setup.py
func TestReadVersion_SetupPy(t *testing.T) {
	useDockerMode(t)
	t.Run("StandardVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup.py", "setup.py")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", v.String())
	})

	t.Run("SingleQuotes", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_py_single_quotes.py", "setup.py")
		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "2.0.0-dev", v.String())
	})

	t.Run("EmptyFile_ReturnsError", func(t *testing.T) {
		repo, p := setupEmpty(t, "setup.py")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("NoVersionKeyword_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_py_no_version.py", "setup.py")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})

	t.Run("InvalidVersion_ReturnsError", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_py_invalid_version.py", "setup.py")
		_, err := p.ReadVersion(repo)
		assert.Error(t, err)
	})
}

// TestWriteVersion_PyprojectPEP621 tests writing version to PEP 621 pyproject.toml
func TestWriteVersion_PyprojectPEP621(t *testing.T) {
	useDockerMode(t)
	t.Run("ReplaceExistingVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621.toml", "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("2", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", v.String())
	})

	t.Run("AddVersionToEmptyFile", func(t *testing.T) {
		repo, p := setupEmpty(t, "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("1", "0", "0", "dev")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0-dev", v.String())
	})

	t.Run("AddVersionToFileWithoutVersionField", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621_no_version.toml", "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("1", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", v.String())
	})

	t.Run("PreservesOtherFields", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_pep621.toml", "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("9", "0", "0")))

		content, err := os.ReadFile(filepath.Join(repo.Local(), "pyproject.toml"))
		require.NoError(t, err)
		assert.Contains(t, string(content), `description = "some project"`)
	})
}

// TestWriteVersion_PyprojectPoetry tests writing version to Poetry pyproject.toml
func TestWriteVersion_PyprojectPoetry(t *testing.T) {
	useDockerMode(t)
	t.Run("ReplaceExistingVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_poetry.toml", "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("3", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "3.0.0", v.String())
	})

	t.Run("PreservesPoetryStructure", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "pyproject_poetry.toml", "pyproject.toml")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("2", "0", "0")))

		content, err := os.ReadFile(filepath.Join(repo.Local(), "pyproject.toml"))
		require.NoError(t, err)
		assert.Contains(t, string(content), `[tool.poetry.dependencies]`)
		assert.Contains(t, string(content), `description = "poetry project"`)
	})
}

// TestWriteVersion_SetupCfg tests writing version to setup.cfg
func TestWriteVersion_SetupCfg(t *testing.T) {
	useDockerMode(t)
	t.Run("ReplaceExistingVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup.cfg", "setup.cfg")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("2", "0", "0", "dev")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "2.0.0-dev", v.String())
	})

	t.Run("AddVersionToEmptyFile", func(t *testing.T) {
		repo, p := setupEmpty(t, "setup.cfg")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("1", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", v.String())
	})

	t.Run("AddVersionToFileWithoutMetadata", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_cfg_no_metadata.cfg", "setup.cfg")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("1", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", v.String())
	})
}

// TestWriteVersion_SetupPy tests writing version to setup.py
func TestWriteVersion_SetupPy(t *testing.T) {
	useDockerMode(t)
	t.Run("ReplaceExistingVersion", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup.py", "setup.py")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("2", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", v.String())
	})

	t.Run("ReplaceVersionWithSingleQuotes", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup_py_single_quotes.py", "setup.py")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("3", "0", "0", "dev")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "3.0.0-dev", v.String())
	})

	t.Run("CreateFromEmptyFile", func(t *testing.T) {
		repo, p := setupEmpty(t, "setup.py")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("1", "0", "0")))

		v, err := p.ReadVersion(repo)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", v.String())
	})

	t.Run("PreservesOtherKeywords", func(t *testing.T) {
		repo, p := setupFromTestdata(t, "setup.py", "setup.py")
		require.NoError(t, p.WriteVersion(repo, core.NewVersion("2", "0", "0")))

		content, err := os.ReadFile(filepath.Join(repo.Local(), "setup.py"))
		require.NoError(t, err)
		assert.Contains(t, string(content), `description="A test app"`)
	})
}

// TestReadWriteRoundtrip verifies that read after write returns the same version for all formats
func TestReadWriteRoundtrip(t *testing.T) {
	fixtures := []struct {
		name       string
		fixture    string
		targetFile string
	}{
		{"PEP621", "pyproject_pep621.toml", "pyproject.toml"},
		{"Poetry", "pyproject_poetry.toml", "pyproject.toml"},
		{"SetupCfg", "setup.cfg", "setup.cfg"},
		{"SetupPy", "setup.py", "setup.py"},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			useDockerMode(t)
			repo, p := setupFromTestdata(t, f.fixture, f.targetFile)

			original, err := p.ReadVersion(repo)
			require.NoError(t, err)

			newVersion := core.NewVersion("5", "3", "1", "dev")
			require.NoError(t, p.WriteVersion(repo, newVersion))

			readBack, err := p.ReadVersion(repo)
			require.NoError(t, err)
			assert.Equal(t, "5.3.1-dev", readBack.String())
			assert.NotEqual(t, original.String(), readBack.String())
		})
	}
}
