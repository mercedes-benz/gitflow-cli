/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_Mode_DefaultIsNative(t *testing.T) {
	viper.Reset()
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

func TestExecutor_Mode_CLIFlag_OverridesPluginConfig(t *testing.T) {
	viper.Set("plugins.test-plugin.mode", "native-mode")
	ExecutorModeOverride = ModeDocker
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeDocker, executor.mode())
}

func TestExecutor_Mode_CLINative_OverridesPluginDocker(t *testing.T) {
	viper.Set("plugins.test-plugin.mode", "docker-mode")
	ExecutorModeOverride = ModeNative
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_PluginConfigDocker_UsedWhenNoFlag(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""
	viper.Set("plugins.test-plugin.mode", "docker-mode")
	defer viper.Reset()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	assert.Equal(t, ModeDocker, executor.mode())
}

func TestExecutor_Command_NativeMode(t *testing.T) {
	viper.Reset()
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

func TestExecutor_Command_CustomCommand(t *testing.T) {
	viper.Set("plugins.test-plugin.command", "/opt/custom/bin/test-cmd")
	ExecutorModeOverride = ""
	defer viper.Reset()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	cmd := executor.Command("/tmp/project", "test-cmd", "arg1")

	assert.Equal(t, []string{"/opt/custom/bin/test-cmd", "arg1"}, cmd.Args)
}

func TestExecutor_Command_CustomImage(t *testing.T) {
	viper.Set("plugins.test-plugin.image", "test-image:2.0-custom")
	ExecutorModeOverride = ModeDocker
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	cmd := executor.Command("/tmp/project", "test-cmd", "arg1")

	assert.Equal(t, []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"test-image:2.0-custom", "test-cmd",
		"arg1",
	}, cmd.Args)
}

func TestExecutor_Command_DefaultImage_WhenNoConfigOverride(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "test-plugin", Image: "test-image:1.0"}

	cmd := executor.Command("/tmp/project", "test-cmd", "arg1")

	assert.Contains(t, cmd.Args, "test-image:1.0")
}

func TestExecutor_RequiredTools_NativeMode(t *testing.T) {
	viper.Reset()
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

func TestExecutor_ConfigValues_ModeImageCommand(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""
	defer viper.Reset()

	viper.Set("plugins.plugin-a.mode", "docker-mode")
	viper.Set("plugins.plugin-a.image", "image-a:custom")
	viper.Set("plugins.plugin-a.command", "/opt/bin/cmd-a")
	viper.Set("plugins.plugin-b.mode", "native-mode")
	viper.Set("plugins.plugin-b.image", "image-b:custom")
	viper.Set("plugins.plugin-b.command", "/usr/local/bin/cmd-b")
	viper.Set("plugins.plugin-c.mode", "docker-mode")
	viper.Set("plugins.plugin-c.command", "cmd-c-custom")

	t.Run("plugin-a uses docker-mode from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-a", Image: "image-a:default"}

		assert.Equal(t, ModeDocker, executor.mode())
	})

	t.Run("plugin-a uses custom image from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-a", Image: "image-a:default"}

		assert.Equal(t, "image-a:custom", executor.resolveImage())
	})

	t.Run("plugin-a uses custom command from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-a", Image: "image-a:default"}

		assert.Equal(t, "/opt/bin/cmd-a", executor.resolveCommand("cmd-a"))
	})

	t.Run("plugin-b uses native-mode from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-b", Image: "image-b:default"}

		assert.Equal(t, ModeNative, executor.mode())
	})

	t.Run("plugin-b uses custom image from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-b", Image: "image-b:default"}

		assert.Equal(t, "image-b:custom", executor.resolveImage())
	})

	t.Run("plugin-b uses custom command from config", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-b", Image: "image-b:default"}

		assert.Equal(t, "/usr/local/bin/cmd-b", executor.resolveCommand("cmd-b"))
	})

	t.Run("plugin-c uses docker-mode with custom command and default image", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-c", Image: "image-c:default"}

		assert.Equal(t, ModeDocker, executor.mode())
		assert.Equal(t, "image-c:default", executor.resolveImage())
		assert.Equal(t, "cmd-c-custom", executor.resolveCommand("cmd-c"))
	})

	t.Run("unconfigured plugin falls back to defaults", func(t *testing.T) {
		executor := Executor{PluginName: "plugin-d", Image: "image-d:default"}

		assert.Equal(t, ModeNative, executor.mode())
		assert.Equal(t, "image-d:default", executor.resolveImage())
		assert.Equal(t, "cmd-d", executor.resolveCommand("cmd-d"))
	})
}
