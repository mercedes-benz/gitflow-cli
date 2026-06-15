/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_Mode_DefaultIsNative(t *testing.T) {
	ExecutorModeOverride = ""

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_NoImage_AlwaysNative(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: ""}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_CLIFlag_OverridesDefault(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeDocker, executor.mode())
}

func TestExecutor_Mode_CLINative_Override(t *testing.T) {
	ExecutorModeOverride = ModeNative
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Command_NativeMode(t *testing.T) {
	ExecutorModeOverride = ""

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	cmd := executor.Command("/tmp/project", "test-cmd", "arg1")

	assert.Equal(t, []string{"test-cmd", "arg1"}, cmd.Args)
	assert.Equal(t, "/tmp/project", cmd.Dir)
}

func TestExecutor_Command_DockerRunMode(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	cmd := executor.Command("/tmp/project", "test-cmd", "arg1")

	assert.Equal(t, []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"test-image:1.0", "test-cmd",
		"arg1",
	}, cmd.Args)
}

func TestExecutor_Command_DockerWithSetup(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{
		PluginName:  "test-plugin",
		Image:       "test-image:1.0",
		DockerSetup: []string{"pip install toml-cli"},
	}

	cmd := executor.Command("/tmp/project", "toml", "get", "version")

	assert.Contains(t, cmd.Args, "sh")
	assert.Contains(t, cmd.Args, "-c")
}

func TestExecutor_RequiredTools_NativeMode(t *testing.T) {
	ExecutorModeOverride = ""

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, []string{"test-cmd"}, executor.RequiredTools([]string{"test-cmd"}))
}

func TestExecutor_RequiredTools_DockerMode(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, []string{"docker"}, executor.RequiredTools([]string{"test-cmd"}))
}

func TestExecutor_ResolveMode_NativeToolAvailable(t *testing.T) {
	ExecutorModeOverride = ""
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	// "git" is always available in test environments
	err := executor.ResolveMode([]string{"git"})
	assert.NoError(t, err)
	assert.Equal(t, "", ExecutorModeOverride)
}

func TestExecutor_ResolveMode_FallbackWithConfig(t *testing.T) {
	ExecutorModeOverride = ""
	core.DockerFallback = true
	defer func() {
		ExecutorModeOverride = ""
		core.DockerFallback = false
	}()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	// "nonexistent-tool-xyz" doesn't exist, should trigger docker fallback
	err := executor.ResolveMode([]string{"nonexistent-tool-xyz"})

	// If docker is available, mode switches to docker. If not, error about both missing.
	if err != nil {
		assert.Contains(t, err.Error(), "neither")
	} else {
		assert.Equal(t, ModeDocker, ExecutorModeOverride)
	}
}

func TestExecutor_ResolveMode_NoFallback_NoCallback(t *testing.T) {
	ExecutorModeOverride = ""
	core.DockerFallback = false
	oldFunc := ToolFallbackFunc
	ToolFallbackFunc = nil
	defer func() {
		ExecutorModeOverride = ""
		ToolFallbackFunc = oldFunc
	}()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	err := executor.ResolveMode([]string{"nonexistent-tool-xyz"})

	assert.Error(t, err)
}

func TestExecutor_ResolveMode_SkippedWhenOverrideSet(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	err := executor.ResolveMode([]string{"nonexistent-tool-xyz"})
	assert.NoError(t, err)
}
