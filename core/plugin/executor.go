/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

const (
	ModeDocker = "docker-mode"
	ModeNative = "native-mode"
)

// ExecutorModeOverride is set by CLI flags (--docker-mode/--native-mode) and takes highest priority.
var ExecutorModeOverride string

// Executor executes CLI commands either natively or inside a Docker container.
type Executor struct {
	PluginName  string
	Image       string
	DockerSetup []string
}

// Command returns an *exec.Cmd that runs the given tool with args.
// In native mode, the command runs directly on the host.
// In docker mode, it uses "docker run --rm" with the plugin's image.
// If DockerSetup is set, commands are wrapped in sh -c with setup commands prepended.
func (e *Executor) Command(workDir string, name string, args ...string) *exec.Cmd {
	command := e.resolveCommand(name)

	if e.mode() == ModeNative {
		log.Printf("[executor] native: %s %v (dir=%s)", command, args, workDir)
		cmd := exec.Command(command, args...)
		cmd.Dir = workDir
		return cmd
	}

	image := e.resolveImage()
	dockerArgs := []string{
		"run", "--rm",
		"-v", workDir + ":/work",
		"-w", "/work",
	}

	if len(e.DockerSetup) > 0 {
		dockerArgs = append(dockerArgs, "-v", "gitflow-"+e.PluginName+"-cache:/root/.cache")
	}

	dockerArgs = append(dockerArgs, image)

	if len(e.DockerSetup) > 0 {
		quotedParts := make([]string, 0, len(args)+1)
		quotedParts = append(quotedParts, shellQuote(command))
		for _, a := range args {
			quotedParts = append(quotedParts, shellQuote(a))
		}
		setupChain := strings.Join(e.DockerSetup, " && ")
		fullCmd := setupChain + " && " + strings.Join(quotedParts, " ")
		dockerArgs = append(dockerArgs, "sh", "-c", fullCmd)
		log.Printf("[executor] docker run: image=%s, setup=%v, command=%s %v (dir=%s)", image, e.DockerSetup, command, args, workDir)
	} else {
		dockerArgs = append(dockerArgs, command)
		dockerArgs = append(dockerArgs, args...)
		log.Printf("[executor] docker run: image=%s, command=%s %v (dir=%s)", image, command, args, workDir)
	}

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
	if e.Image == "" {
		return ModeNative
	}
	if ExecutorModeOverride != "" {
		return ExecutorModeOverride
	}
	if mode := viper.GetString(fmt.Sprintf("plugins.%s.mode", e.PluginName)); mode != "" {
		return mode
	}
	return ModeNative
}

// resolveImage returns the configured Docker image for this plugin, or falls back to the default.
func (e *Executor) resolveImage() string {
	key := fmt.Sprintf("plugins.%s.image", e.PluginName)
	if image := viper.GetString(key); image != "" {
		return image
	}
	return e.Image
}

// resolveCommand returns the configured command for this plugin, or falls back to the given name.
func (e *Executor) resolveCommand(name string) string {
	key := fmt.Sprintf("plugins.%s.command", e.PluginName)
	if command := viper.GetString(key); command != "" {
		return command
	}
	return name
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`!#&|;(){}[]<>?*~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
