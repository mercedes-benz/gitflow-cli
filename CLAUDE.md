# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Workflow Behavior

See [README.md](README.md) for the complete user-facing documentation: workflow steps (release start/finish, hotfix start/finish), CLI flags (`--no-push`, `--docker-mode`, `--native-mode`, `--yes`), configuration keys, and plugin execution modes. When modifying workflow logic, always verify that the README still accurately describes the behavior.

## Build & Run

```bash
go build -o gitflow-cli .
go install
go run .
```

## Testing

Run all tests (unit + e2e):
```bash
go test ./...
```

Run a specific plugin's tests:
```bash
go test ./plugin/mvn/ -run TestReleaseStart -v
```

Run only fallback e2e tests:
```bash
go test ./e2e/workflow/ -v
```

Run configuration e2e tests:
```bash
go test ./e2e/ -v
```

The e2e tests create temporary git repos (bare remote + local clone), commit version files from embedded templates, invoke the CLI via `cmd.Execute()` in-process, and assert branch/tag/version state. No external git server is needed.

## Architecture

This is a Go CLI (Cobra + Viper) that automates the gitflow branching model (release start/finish, hotfix start/finish) with automatic semantic version bumping.

### Registry pattern (fundamental rule)

Plugins register themselves — `core`, `e2e`, and other packages MUST NOT know about individual plugins. This applies to:

- **Runtime**: plugins self-register via `init()` using `core.RegisterPlugin()`
- **Testing**: each plugin owns its own e2e tests in `<name>_test.go`, calling shared workflow functions from `e2e/workflow/`
- **Imports**: only `plugin/plugin.go` imports individual plugin packages (via blank imports to trigger `init()`)

When adding cross-cutting functionality, always ask: "Can the plugin register itself?" If yes, use a registry. Never create a central list that enumerates plugins.

### Core flow

`main.go` → `cmd/root.go` (Cobra commands) → `core.Start()`/`core.Finish()` → plugin detection → workflow execution (branch, merge, tag, push).

### Plugin system

Plugins implement `core.Plugin` interface (ReadVersion, WriteVersion, VersionFileName, VersionQualifier, RequiredTools). They self-register via `init()` functions using `core.RegisterPlugin()`.

- `core/plugin/` — base `Plugin` struct, `Config`, `TestConfig`, and `Factory` that injects the global `HookRegistry`
- `plugin/standard/` — fallback plugin using `version.txt` (also registered via `RegisterFallbackPlugin`)
- `plugin/mvn/` — Maven (`pom.xml`)
- `plugin/npm/` — npm (`package.json`)
- `plugin/composer/` — Composer (`composer.json`)
- `plugin/road/` — road manifest (`road.yaml`)
- `plugin/python/` — Python (`pyproject.toml`, `setup.cfg`, `setup.py`)

Plugin detection: iterates `pluginRegistry` in order, first plugin whose version file exists in the project wins. Falls back to `standard` plugin.

### Plugin file structure

Each plugin is self-contained in `plugin/<name>/`:

- `<name>.go` — plugin implementation, `init()` with `core.RegisterPlugin()`
- `<name>_test.go` — unit tests AND e2e tests (calls `e2e/workflow.Run*()`)
- `testdata/e2e/<file>.tpl` — version file templates (embedded via `//go:embed`)
- `testdata/unit/` — unit test fixtures (if needed)

### E2E test architecture

- `e2e/workflow/` — exported test functions (`RunReleaseStart`, `RunHotfixFinish`, etc.) that define the generic workflow assertions. This package is a library — it provides test logic but does not call it.
- `e2e/test_env.go` — `GitTestEnv` (repo setup, git commands, assertions)
- Each plugin's `_test.go` imports `e2e/workflow` and calls the shared functions with its own `TestConfig`
- Fallback tests (no-plugin behavior) live in `plugin/standard/standard_test.go` (the standard plugin IS the fallback)
- Configuration tests (custom branch names) live in `cmd/root_test.go`

### Hook system

`core/hook.go` defines `HookRegistry` with typed hooks (`ReleaseStartHooks`, `HotfixStartHooks`, `HotfixFinishHooks`). Plugins register hooks during `init()` via `Plugin.RegisterHook()`. Hooks run at specific workflow points (e.g., before release start, after merge into develop).

### Repository abstraction

`core/repository.go` — `Repository` interface wraps all git operations (checkout, merge, tag, push, rollback). Every method shells out to `git` via `exec.Command`. The `Rollback` method resets the repo to remote state when `workflow.rollback: true` is configured.

### Version handling

`core/version.go` — `Version` struct with Major/Minor/Incremental/Qualifier. `ParseVersion` uses regex `(\d+)\.(\d+)\.(\d+)(?:-(\w+))?$`. Version increment logic: `Next(Minor)` bumps minor and resets incremental to 0; `Next(Incremental)` bumps patch.

### Configuration

Reads `$HOME/.gitflow-cli.yaml` (or `--config` flag) via Viper. Settings are grouped under `branches:` (production, development, release, hotfix), `workflow:` (push, rollback, docker-fallback), and `logging:`.

## Adding a new plugin

1. Create `plugin/<name>/<name>.go`
2. Define a `Config` with name, version file, qualifier, required tools
3. Embed `plugin.Plugin` from `core/plugin`
4. Implement `ReadVersion` and `WriteVersion`
5. In `init()`: use `plugin.NewFactory().NewPlugin(config)`, register hooks, call `core.RegisterPlugin()`
6. Add blank import to `plugin/plugin.go`
7. Create `plugin/<name>/testdata/e2e/<file>.tpl` with version file template
8. Create `plugin/<name>/<name>_test.go` with e2e tests calling `e2e/workflow.Run*()`

## License headers

All `.go` files require the SPDX header:
```
/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/
```
