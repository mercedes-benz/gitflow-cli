/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/spf13/viper"
)

const (
	ModeDocker = "docker"
	ModeNative = "native"
)

// activeContainers maps plugin names to running container IDs.
// Tests register containers here so that Executor uses "docker exec" instead of "docker run".
var (
	activeContainers   = make(map[string]string)
	activeContainersMu sync.RWMutex
)

// RegisterContainer registers a running container ID for a plugin.
// When set, the executor uses "docker exec" on this container instead of "docker run".
func RegisterContainer(pluginName, containerID string) {
	activeContainersMu.Lock()
	defer activeContainersMu.Unlock()
	activeContainers[pluginName] = containerID
	log.Printf("[executor] Registered container for plugin=%s containerID=%s", pluginName, containerID[:12])
}

// UnregisterContainer removes the container registration for a plugin.
func UnregisterContainer(pluginName string) {
	activeContainersMu.Lock()
	defer activeContainersMu.Unlock()
	delete(activeContainers, pluginName)
	log.Printf("[executor] Unregistered container for plugin=%s", pluginName)
}

// Executor executes CLI commands either natively or inside a Docker container.
type Executor struct {
	PluginName string
	Image      string
}

// Command returns an *exec.Cmd that runs the given tool with args.
// In native mode, the command runs directly on the host.
// In docker mode with a registered container, it uses "docker exec" on the running container.
// In docker mode without a registered container, it uses "docker run --rm" (ephemeral container).
func (e *Executor) Command(workDir string, name string, args ...string) *exec.Cmd {
	command := e.resolveCommand(name)

	if e.mode() == ModeNative {
		log.Printf("[executor] native: %s %v (dir=%s)", command, args, workDir)
		cmd := exec.Command(command, args...)
		cmd.Dir = workDir
		return cmd
	}

	if containerID := e.containerID(); containerID != "" {
		dockerArgs := []string{"exec", "-w", "/work", containerID, command}
		dockerArgs = append(dockerArgs, args...)
		log.Printf("[executor] docker exec: container=%s, command=%s %v", containerID[:12], command, args)
		return exec.Command("docker", dockerArgs...)
	}

	dockerArgs := []string{
		"run", "--rm",
		"-v", workDir + ":/work",
		"-w", "/work",
		e.Image, command,
	}
	dockerArgs = append(dockerArgs, args...)
	log.Printf("[executor] docker run: image=%s, command=%s %v (dir=%s)", e.Image, command, args, workDir)
	return exec.Command("docker", dockerArgs...)
}

// RequiredTools returns the tools that must be available on the system.
// In docker mode, only "docker" is required. In native mode, the plugin's own tools are needed.
func (e *Executor) RequiredTools(nativeTools []string) []string {
	if e.mode() == ModeNative {
		return nativeTools
	}
	return []string{"docker"}
}

func (e *Executor) mode() string {
	key := fmt.Sprintf("plugins.%s.executor", e.PluginName)
	mode := viper.GetString(key)
	if mode == ModeNative {
		return ModeNative
	}
	return ModeDocker
}

// resolveCommand returns the configured command for this plugin, or falls back to the given name.
func (e *Executor) resolveCommand(name string) string {
	key := fmt.Sprintf("plugins.%s.command", e.PluginName)
	if command := viper.GetString(key); command != "" {
		return command
	}
	return name
}

func (e *Executor) containerID() string {
	activeContainersMu.RLock()
	defer activeContainersMu.RUnlock()
	return activeContainers[e.PluginName]
}
