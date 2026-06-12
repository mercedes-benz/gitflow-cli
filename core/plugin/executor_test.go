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

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_NoImage_AlwaysNative(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "standard", Image: ""}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_CLIFlag_OverridesPluginConfig(t *testing.T) {
	viper.Set("plugins.npm.mode", "native-mode")
	ExecutorModeOverride = ModeDocker
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, ModeDocker, executor.mode())
}

func TestExecutor_Mode_CLINative_OverridesPluginDocker(t *testing.T) {
	viper.Set("plugins.npm.mode", "docker-mode")
	ExecutorModeOverride = ModeNative
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, ModeNative, executor.mode())
}

func TestExecutor_Mode_PluginConfigDocker_UsedWhenNoFlag(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""
	viper.Set("plugins.npm.mode", "docker-mode")
	defer viper.Reset()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, ModeDocker, executor.mode())
}

func TestExecutor_Command_NativeMode(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	cmd := executor.Command("/tmp/project", "npm", "version")

	assert.Equal(t, []string{"npm", "version"}, cmd.Args)
	assert.Equal(t, "/tmp/project", cmd.Dir)
}

func TestExecutor_Command_DockerRunMode(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	cmd := executor.Command("/tmp/project", "npm", "version")

	assert.Equal(t, []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"node:20-slim", "npm",
		"version",
	}, cmd.Args)
}

func TestExecutor_Command_CustomCommand(t *testing.T) {
	viper.Set("plugins.mvn.command", "/opt/maven/bin/mvn")
	ExecutorModeOverride = ""
	defer viper.Reset()

	executor := Executor{PluginName: "mvn", Image: "maven:3.9"}

	cmd := executor.Command("/tmp/project", "mvn", "versions:set")

	assert.Equal(t, []string{"/opt/maven/bin/mvn", "versions:set"}, cmd.Args)
}

func TestExecutor_Command_CustomImage(t *testing.T) {
	viper.Set("plugins.npm.image", "node:22-alpine")
	ExecutorModeOverride = ModeDocker
	defer func() {
		ExecutorModeOverride = ""
		viper.Reset()
	}()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	cmd := executor.Command("/tmp/project", "npm", "version")

	assert.Equal(t, []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"node:22-alpine", "npm",
		"version",
	}, cmd.Args)
}

func TestExecutor_Command_DefaultImage_WhenNoConfigOverride(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	cmd := executor.Command("/tmp/project", "npm", "version")

	assert.Contains(t, cmd.Args, "node:20-slim")
}

func TestExecutor_RequiredTools_NativeMode(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, []string{"npm"}, executor.RequiredTools([]string{"npm"}))
}

func TestExecutor_RequiredTools_DockerMode(t *testing.T) {
	ExecutorModeOverride = ModeDocker
	defer func() { ExecutorModeOverride = "" }()

	executor := Executor{PluginName: "npm", Image: "node:20-slim"}

	assert.Equal(t, []string{"docker"}, executor.RequiredTools([]string{"npm"}))
}

func TestExecutor_ConfigValues_ModeImageCommand(t *testing.T) {
	viper.Reset()
	ExecutorModeOverride = ""
	defer viper.Reset()

	viper.Set("plugins.mvn.mode", "docker-mode")
	viper.Set("plugins.mvn.image", "maven:3.9-custom")
	viper.Set("plugins.mvn.command", "/opt/maven/bin/mvn")
	viper.Set("plugins.npm.mode", "native-mode")
	viper.Set("plugins.npm.image", "node:22-alpine")
	viper.Set("plugins.npm.command", "/usr/local/bin/npm")
	viper.Set("plugins.python.mode", "docker-mode")
	viper.Set("plugins.python.command", "toml-cli")

	t.Run("mvn uses docker-mode from config", func(t *testing.T) {
		executor := Executor{PluginName: "mvn", Image: "maven:3.9-eclipse-temurin-17"}

		assert.Equal(t, ModeDocker, executor.mode())
	})

	t.Run("mvn uses custom image from config", func(t *testing.T) {
		executor := Executor{PluginName: "mvn", Image: "maven:3.9-eclipse-temurin-17"}

		assert.Equal(t, "maven:3.9-custom", executor.resolveImage())
	})

	t.Run("mvn uses custom command from config", func(t *testing.T) {
		executor := Executor{PluginName: "mvn", Image: "maven:3.9-eclipse-temurin-17"}

		assert.Equal(t, "/opt/maven/bin/mvn", executor.resolveCommand("mvn"))
	})

	t.Run("npm uses native-mode from config", func(t *testing.T) {
		executor := Executor{PluginName: "npm", Image: "node:20-slim"}

		assert.Equal(t, ModeNative, executor.mode())
	})

	t.Run("npm uses custom image from config", func(t *testing.T) {
		executor := Executor{PluginName: "npm", Image: "node:20-slim"}

		assert.Equal(t, "node:22-alpine", executor.resolveImage())
	})

	t.Run("npm uses custom command from config", func(t *testing.T) {
		executor := Executor{PluginName: "npm", Image: "node:20-slim"}

		assert.Equal(t, "/usr/local/bin/npm", executor.resolveCommand("npm"))
	})

	t.Run("python uses docker-mode with custom command and default image", func(t *testing.T) {
		executor := Executor{PluginName: "python", Image: "python:3.12-slim"}

		assert.Equal(t, ModeDocker, executor.mode())
		assert.Equal(t, "python:3.12-slim", executor.resolveImage())
		assert.Equal(t, "toml-cli", executor.resolveCommand("toml"))
	})

	t.Run("composer falls back to defaults when not in config", func(t *testing.T) {
		executor := Executor{PluginName: "composer", Image: "composer:2"}

		assert.Equal(t, ModeNative, executor.mode())
		assert.Equal(t, "composer:2", executor.resolveImage())
		assert.Equal(t, "composer", executor.resolveCommand("composer"))
	})
}
