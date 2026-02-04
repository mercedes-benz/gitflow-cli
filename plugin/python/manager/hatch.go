/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HatchManager handles Hatch configuration using the hatch CLI
type HatchManager struct {
	projectPath string
	filePath    string
}

// NewHatchManager creates a new Hatch manager instance
// It verifies that:
// 1. pyproject.toml exists
// 2. hatch command is available
// 3. The project is a valid Hatch project
func NewHatchManager(projectPath string) (VersionManager, error) {
	filePath := filepath.Join(projectPath, "pyproject.toml")
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("pyproject.toml not found at %s", projectPath)
	}

	// Verify hatch is installed
	if _, err := exec.LookPath("hatch"); err != nil {
		return nil, fmt.Errorf("hatch command not found - please install hatch (pip install hatch)")
	}

	// Verify this is a valid Hatch project by trying to read version
	cmd := exec.Command("hatch", "version")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("not a valid Hatch project or version not configured: %s", string(output))
	}

	return &HatchManager{
		projectPath: projectPath,
		filePath:    filePath,
	}, nil
}

// GetVersion reads the version using the hatch CLI
func (h *HatchManager) GetVersion() (string, error) {
	cmd := exec.Command("hatch", "version")
	cmd.Dir = h.projectPath

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get version from hatch: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute hatch version: %v", err)
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("hatch returned empty version")
	}

	return version, nil
}

// SetVersion writes the version using the hatch CLI
func (h *HatchManager) SetVersion(version string) error {
	cmd := exec.Command("hatch", "version", version)
	cmd.Dir = h.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set version with hatch: %v: %s", err, string(output))
	}

	return nil
}

// GetName returns the name of this manager
func (h *HatchManager) GetName() string {
	return "hatch"
}

// GetFilePath returns the path to the version file
func (h *HatchManager) GetFilePath() string {
	return filepath.Base(h.filePath)
}
