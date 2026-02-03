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

	"github.com/pelletier/go-toml/v2"
)

// ManagerDetector detects and instantiates the appropriate Python package manager
type ManagerDetector struct {
	detectors []detectorConfig
}

// detectorConfig defines how to detect and initialize a package manager
type detectorConfig struct {
	name          string
	requiredFiles []string
	requiredTools []string
	priority      int
	detectFunc    func(string) bool
	factory       func(string) (VersionManager, error)
}

// NewManagerDetector creates a new manager detector with detection priority order
// Priority: higher number = higher priority
// Hatch (8) > setup.py (5)
func NewManagerDetector() *ManagerDetector {
	return &ManagerDetector{
		detectors: []detectorConfig{
			{
				name:          "hatch",
				requiredFiles: []string{"pyproject.toml"},
				requiredTools: []string{"hatch"},
				priority:      8,
				detectFunc:    isHatchProject,
				factory:       NewHatchManager,
			},
			{
				name:          "setup.py",
				requiredFiles: []string{"setup.py"},
				requiredTools: []string{},
				priority:      5,
				detectFunc:    nil, // No special detection needed
				factory:       NewSetupPyManager,
			},
		},
	}
}

// Detect finds and returns the appropriate VersionManager based on project files and installed tools
func (d *ManagerDetector) Detect(projectPath string) (VersionManager, error) {
	var candidates []detectorConfig

	// Iterate through all detectors and find candidates
	for _, detector := range d.detectors {
		// Step 1: Check if all required files exist
		if !hasAllFiles(projectPath, detector.requiredFiles) {
			continue
		}

		// Step 2: Check if at least one required tool is installed
		if !hasAnyTool(detector.requiredTools) {
			continue
		}

		// Step 3: Run custom detection function if provided
		if detector.detectFunc != nil && !detector.detectFunc(projectPath) {
			continue
		}

		// This detector is a candidate
		candidates = append(candidates, detector)
	}

	// No candidates found
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no supported Python package manager detected")
	}

	// Select the candidate with the highest priority
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.priority > best.priority {
			best = candidate
		}
	}

	// Try to initialize the selected manager
	mgr, err := best.factory(projectPath)
	if err != nil {
		// If initialization fails, try the next best candidate
		for _, candidate := range candidates {
			if candidate.name == best.name {
				continue // Skip the one we just tried
			}

			mgr, err = candidate.factory(projectPath)
			if err == nil {
				return mgr, nil
			}
		}

		// All candidates failed
		return nil, fmt.Errorf("failed to initialize any Python package manager: %v", err)
	}

	return mgr, nil
}

// isHatchProject checks if the pyproject.toml indicates a Hatch project
func isHatchProject(projectPath string) bool {
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return false
	}

	var config struct {
		BuildSystem struct {
			BuildBackend string `toml:"build-backend"`
		} `toml:"build-system"`
		Tool struct {
			Hatch map[string]interface{} `toml:"hatch"`
		} `toml:"tool"`
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return false
	}

	// Check if using hatchling as build backend OR has [tool.hatch] config
	return config.BuildSystem.BuildBackend == "hatchling.build" || config.Tool.Hatch != nil
}

// hasAllFiles checks if all required files exist in the project directory
func hasAllFiles(projectPath string, files []string) bool {
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err != nil {
			return false
		}
	}
	return true
}

func hasAnyTool(tools []string) bool {
	if len(tools) == 0 {
		return true // No tools required (empty slice)
	}

	for _, tool := range tools {
		if tool == "" {
			return true // No tools required (empty string)
		}
		if _, err := exec.LookPath(tool); err == nil {
			return true // Found at least one tool
		}
	}

	return false // None of the tools found
}
