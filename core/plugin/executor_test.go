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

func TestExecutor_Command_NativeMode(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "native")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1", "arg2")

	assert.Equal(t, "/tmp/project", cmd.Dir)
	assert.Equal(t, []string{"sometool", "arg1", "arg2"}, cmd.Args)
}

func TestExecutor_Command_DockerMode(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "docker")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1", "arg2")

	expectedArgs := []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"some-image:latest", "sometool",
		"arg1", "arg2",
	}
	assert.Equal(t, expectedArgs, cmd.Args)
}

func TestExecutor_Command_DefaultMode_IsDocker(t *testing.T) {
	viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1")

	assert.Contains(t, cmd.Args, "docker")
	assert.Contains(t, cmd.Args, "some-image:latest")
}

func TestExecutor_RequiredTools_DockerMode(t *testing.T) {
	viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	tools := executor.RequiredTools([]string{"sometool"})
	assert.Equal(t, []string{"docker"}, tools)
}

func TestExecutor_RequiredTools_NativeMode(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "native")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	tools := executor.RequiredTools([]string{"sometool"})
	assert.Equal(t, []string{"sometool"}, tools)
}

func TestExecutor_ConfigurationSwitch(t *testing.T) {
	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	// Default: Docker mode
	viper.Reset()
	cmd := executor.Command("/tmp/project", "sometool", "run")
	assert.Equal(t, "docker", cmd.Args[0])

	// Switch to native
	viper.Set("plugins.testplugin.executor", "native")
	cmd = executor.Command("/tmp/project", "sometool", "run")
	assert.Equal(t, []string{"sometool", "run"}, cmd.Args)
	assert.Equal(t, "/tmp/project", cmd.Dir)

	// Switch back to docker
	viper.Set("plugins.testplugin.executor", "docker")
	cmd = executor.Command("/tmp/project", "sometool", "run")
	assert.Equal(t, "docker", cmd.Args[0])
	assert.Contains(t, cmd.Args, "some-image:latest")

	viper.Reset()
}

func TestExecutor_Command_CustomCommand_NativeMode(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "native")
	viper.Set("plugins.testplugin.command", "/opt/custom/mytool")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1")

	assert.Equal(t, []string{"/opt/custom/mytool", "arg1"}, cmd.Args)
	assert.Equal(t, "/tmp/project", cmd.Dir)
}

func TestExecutor_Command_CustomCommand_DockerMode(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "docker")
	viper.Set("plugins.testplugin.command", "custom-tool")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1")

	expectedArgs := []string{
		"docker", "run", "--rm",
		"-v", "/tmp/project:/work",
		"-w", "/work",
		"some-image:latest", "custom-tool",
		"arg1",
	}
	assert.Equal(t, expectedArgs, cmd.Args)
}

func TestExecutor_Command_NoCustomCommand_FallsBackToName(t *testing.T) {
	viper.Set("plugins.testplugin.executor", "native")
	defer viper.Reset()

	executor := &Executor{
		PluginName: "testplugin",
		Image:      "some-image:latest",
	}

	cmd := executor.Command("/tmp/project", "sometool", "arg1")

	assert.Equal(t, []string{"sometool", "arg1"}, cmd.Args)
}
