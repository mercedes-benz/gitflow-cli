/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed scripts/read_setup_cfg.py
var readSetupCfgScript string

//go:embed scripts/write_setup_cfg.py
var writeSetupCfgScript string

//go:embed scripts/read_setup_py.py
var readSetupPyScript string

//go:embed scripts/write_setup_py.py
var writeSetupPyScript string

const (
	python3 = "python3"
	toml    = "toml"
)

type pythonPlugin struct {
	plugin.Plugin
}

var pluginConfig = plugin.Config{
	Name: "python",
	VersionFileNames: []string{
		"pyproject.toml",
		"setup.cfg",
		"setup.py",
	},
	VersionQualifier: "dev",
	RequiredTools:    []string{python3, toml},
	DockerImage:      "python:3.12-slim",
	DockerSetup:      []string{"pip install -q toml-cli"},
}

func init() {
	pluginFactory := plugin.NewFactory()

	p := &pythonPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	p.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, p.beforeReleaseStart)
	p.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, p.beforeHotfixStart)

	core.RegisterPlugin(p)
}

func (p *pythonPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs = make([]any, 0)
	defer func() { core.Log(logs...) }()

	projectPath := repository.Local()

	versionString, err := p.readVersion(projectPath)
	if err != nil {
		logs = append(logs, err)
		return core.Version{}, err
	}

	logs = append(logs, fmt.Sprintf("Read version from %s: %s", p.VersionFileName(), versionString))

	version, err := core.ParseVersion(versionString)
	if err != nil {
		return core.Version{}, fmt.Errorf("failed to parse version: %v", err)
	}

	return version, nil
}

func (p *pythonPlugin) WriteVersion(repository core.Repository, version core.Version) error {
	var logs = make([]any, 0)
	defer func() { core.Log(logs...) }()

	projectPath := repository.Local()

	if err := p.writeVersion(projectPath, version.String()); err != nil {
		logs = append(logs, err)
		return err
	}

	logs = append(logs, fmt.Sprintf("Wrote version %s to %s", version.String(), p.VersionFileName()))
	return nil
}

func (p *pythonPlugin) readVersion(projectPath string) (string, error) {
	switch p.VersionFileName() {
	case "pyproject.toml":
		return p.readPyprojectVersion(projectPath)
	case "setup.cfg":
		return p.runPython(projectPath, readSetupCfgScript, p.VersionFileName())
	case "setup.py":
		return p.runPython(projectPath, readSetupPyScript, p.VersionFileName())
	default:
		return "", fmt.Errorf("unsupported version file: %s", p.VersionFileName())
	}
}

func (p *pythonPlugin) writeVersion(projectPath, version string) error {
	switch p.VersionFileName() {
	case "pyproject.toml":
		return p.writePyprojectVersion(projectPath, version)
	case "setup.cfg":
		_, err := p.runPython(projectPath, writeSetupCfgScript, p.VersionFileName(), version)
		return err
	case "setup.py":
		_, err := p.runPython(projectPath, writeSetupPyScript, p.VersionFileName(), version)
		return err
	default:
		return fmt.Errorf("unsupported version file: %s", p.VersionFileName())
	}
}

func (p *pythonPlugin) readPyprojectVersion(projectPath string) (string, error) {
	cmd := p.Executor.Command(projectPath, toml, "get", "--toml-path", p.VersionFileName(), "project.version")
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	cmd = p.Executor.Command(projectPath, toml, "get", "--toml-path", p.VersionFileName(), "tool.poetry.version")
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	return "", fmt.Errorf("no version found in pyproject.toml")
}

func (p *pythonPlugin) writePyprojectVersion(projectPath, version string) error {
	cmd := p.Executor.Command(projectPath, toml, "get", "--toml-path", p.VersionFileName(), "project.version")
	if cmd.Run() == nil {
		return p.runToml(projectPath, "set", "--toml-path", p.VersionFileName(), "project.version", version)
	}
	cmd = p.Executor.Command(projectPath, toml, "get", "--toml-path", p.VersionFileName(), "tool.poetry.version")
	if cmd.Run() == nil {
		return p.runToml(projectPath, "set", "--toml-path", p.VersionFileName(), "tool.poetry.version", version)
	}
	// No existing section — create project section and set version
	p.Executor.Command(projectPath, toml, "add_section", "--toml-path", p.VersionFileName(), "project").Run()
	return p.runToml(projectPath, "set", "--toml-path", p.VersionFileName(), "project.version", version)
}

func (p *pythonPlugin) runToml(projectPath string, args ...string) error {
	cmd := p.Executor.Command(projectPath, toml, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("toml %s failed: %v: %s", args[0], err, output)
	}
	return nil
}

func (p *pythonPlugin) runPython(projectPath, script string, args ...string) (string, error) {
	cmdArgs := append([]string{"-c", script}, args...)
	cmd := p.Executor.Command(projectPath, python3, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python3 failed: %v: %s", err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func (p *pythonPlugin) beforeReleaseStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	if _, err := p.ReadVersion(repository); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0", p.Config.VersionQualifier)
	if err := p.WriteVersion(repository, initVersion); err != nil {
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(fmt.Sprintf("Set initial project version to %s", initVersion.String()))
	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}

func (p *pythonPlugin) beforeHotfixStart(repository core.Repository) error {
	if err := repository.CheckoutBranch(core.Production.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	if _, err := p.ReadVersion(repository); err == nil {
		return nil
	}

	initVersion := core.NewVersion("1", "0", "0")
	if err := p.WriteVersion(repository, initVersion); err != nil {
		return repository.UndoAllChanges(fmt.Errorf("failed to set initial version: %v", err))
	}

	core.Log(fmt.Sprintf("Set initial project version to %s", initVersion.String()))
	if err := repository.CommitChanges("Set initial project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	return nil
}
