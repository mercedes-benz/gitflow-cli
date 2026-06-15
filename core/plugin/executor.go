/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mercedes-benz/gitflow-cli/core"
)

const (
	ModeDocker = "docker-mode"
	ModeNative = "native-mode"
)

// ExecutorModeOverride is set by CLI flags (--docker-mode/--native-mode) and takes highest priority.
var ExecutorModeOverride string

// ToolFallbackFunc is called when a native tool is not found and docker-fallback is not
// automatically enabled. It asks the user whether to use Docker instead.
// Returns true to proceed with Docker, false to abort.
var ToolFallbackFunc func(tool string, image string) (bool, error)

// Executor executes CLI commands either natively or inside a Docker container.
type Executor struct {
	PluginName  string
	Image       string
	DockerSetup []string
}

// Command returns an *exec.Cmd that runs the given tool with args.
// In native mode, the command runs directly on the host.
// In docker mode, it uses "docker run --rm" with the plugin's image.
func (e *Executor) Command(workDir string, name string, args ...string) *exec.Cmd {
	if e.mode() == ModeNative {
		log.Printf("[executor] native: %s %v (dir=%s)", name, args, workDir)
		cmd := exec.Command(name, args...)
		cmd.Dir = workDir
		return cmd
	}

	image := e.Image
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
		quotedParts = append(quotedParts, shellQuote(name))
		for _, a := range args {
			quotedParts = append(quotedParts, shellQuote(a))
		}
		setupChain := strings.Join(e.DockerSetup, " && ")
		fullCmd := setupChain + " && " + strings.Join(quotedParts, " ")
		dockerArgs = append(dockerArgs, "sh", "-c", fullCmd)
		log.Printf("[executor] docker run: image=%s, setup=%v, command=%s %v (dir=%s)", image, e.DockerSetup, name, args, workDir)
	} else {
		dockerArgs = append(dockerArgs, name)
		dockerArgs = append(dockerArgs, args...)
		log.Printf("[executor] docker run: image=%s, command=%s %v (dir=%s)", image, name, args, workDir)
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

// ResolveMode determines the effective execution mode, applying docker fallback logic
// when native tools are missing. Call this before RequiredTools to allow fallback.
func (e *Executor) ResolveMode(nativeTools []string) error {
	if ExecutorModeOverride != "" || e.Image == "" {
		return nil
	}

	// Check if native tools are available
	for _, tool := range nativeTools {
		if _, err := exec.LookPath(tool); err != nil {
			return e.handleMissingTool(tool)
		}
	}
	return nil
}

func (e *Executor) handleMissingTool(tool string) error {
	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("neither '%s' nor 'docker' found — install one of them", tool)
	}

	// Auto-fallback if configured
	if core.DockerFallback {
		fmt.Fprintf(os.Stderr, "INFO: %s not found, using Docker (%s)\n", tool, e.Image)
		ExecutorModeOverride = ModeDocker
		return nil
	}

	// Ask the user via the callback
	if ToolFallbackFunc != nil {
		proceed, err := ToolFallbackFunc(tool, e.Image)
		if err != nil {
			return err
		}
		if proceed {
			ExecutorModeOverride = ModeDocker
			return nil
		}
	}

	return fmt.Errorf("'%s' not found — install it or enable docker-fallback in config", tool)
}

func (e *Executor) mode() string {
	if e.Image == "" {
		return ModeNative
	}
	if ExecutorModeOverride != "" {
		return ExecutorModeOverride
	}
	return ModeNative
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
