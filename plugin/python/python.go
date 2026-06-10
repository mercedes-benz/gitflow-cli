/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"
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
	RequiredTools:    []string{python3},
}

func init() {
	pluginFactory := plugin.NewFactory()

	p := &pythonPlugin{
		Plugin: pluginFactory.NewPlugin(pluginConfig),
	}

	p.RegisterHook(core.ReleaseStartHooks.BeforeReleaseStartHook, p.beforeReleaseStart)
	p.RegisterHook(core.HotfixStartHooks.BeforeHotfixStartHook, p.beforeHotfixStart)

	core.RegisterPlugin(p)
	fmt.Printf("python plugin init: registered with versionFileNames=%v\n", pluginConfig.VersionFileNames)
}

func (p *pythonPlugin) ReadVersion(repository core.Repository) (core.Version, error) {
	var logs = make([]any, 0)
	defer func() { core.Log(logs...) }()

	filePath := filepath.Join(repository.Local(), p.VersionFileName())

	versionString, err := p.readVersion(filePath, repository.Local())
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

	filePath := filepath.Join(repository.Local(), p.VersionFileName())

	if err := p.writeVersion(filePath, version.String(), repository.Local()); err != nil {
		logs = append(logs, err)
		return err
	}

	logs = append(logs, fmt.Sprintf("Wrote version %s to %s", version.String(), p.VersionFileName()))
	return nil
}

func (p *pythonPlugin) readVersion(filePath, dir string) (string, error) {
	switch p.VersionFileName() {
	case "pyproject.toml":
		return readPyprojectVersion(filePath)
	case "setup.cfg":
		return runPython(dir, readSetupCfgScript, filePath)
	case "setup.py":
		return runPython(dir, readSetupPyScript, filePath)
	default:
		return "", fmt.Errorf("unsupported version file: %s", p.VersionFileName())
	}
}

func (p *pythonPlugin) writeVersion(filePath, version, dir string) error {
	switch p.VersionFileName() {
	case "pyproject.toml":
		return writePyprojectVersion(filePath, version)
	case "setup.cfg":
		_, err := runPython(dir, writeSetupCfgScript, filePath, version)
		return err
	case "setup.py":
		_, err := runPython(dir, writeSetupPyScript, filePath, version)
		return err
	default:
		return fmt.Errorf("unsupported version file: %s", p.VersionFileName())
	}
}

func readPyprojectVersion(filePath string) (string, error) {
	if out, err := exec.Command(toml, "get", "--toml-path", filePath, "project.version").Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if out, err := exec.Command(toml, "get", "--toml-path", filePath, "tool.poetry.version").Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	return "", fmt.Errorf("no version found in pyproject.toml")
}

func writePyprojectVersion(filePath, version string) error {
	if exec.Command(toml, "get", "--toml-path", filePath, "project.version").Run() == nil {
		return runToml("set", "--toml-path", filePath, "project.version", version)
	}
	if exec.Command(toml, "get", "--toml-path", filePath, "tool.poetry.version").Run() == nil {
		return runToml("set", "--toml-path", filePath, "tool.poetry.version", version)
	}
	exec.Command(toml, "add_section", "--toml-path", filePath, "project").Run()
	return runToml("set", "--toml-path", filePath, "project.version", version)
}

func runToml(args ...string) error {
	if output, err := exec.Command(toml, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("toml %s failed: %v: %s", args[0], err, output)
	}
	return nil
}

func runPython(dir, script string, args ...string) (string, error) {
	cmd := exec.Command(python3, append([]string{"-c", script}, args...)...)
	cmd.Dir = dir
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
