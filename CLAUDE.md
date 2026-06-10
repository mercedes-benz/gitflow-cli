# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

Run a specific test:
```bash
go test ./e2e/workflow/ -run TestReleaseStart/StandardPlugin -v
```

Run only e2e tests:
```bash
go test ./e2e/... -v
```

The e2e tests create temporary git repos (bare remote + local clone), commit version files from templates in `e2e/helper/templates/`, invoke the CLI via `cmd.Execute()` in-process, and assert branch/tag/version state. No external git server is needed.

## Architecture

This is a Go CLI (Cobra + Viper) that automates the gitflow branching model (release start/finish, hotfix start/finish) with automatic semantic version bumping.

### Core flow

`main.go` → `cmd/root.go` (Cobra commands) → `core.Start()`/`core.Finish()` → plugin detection → workflow execution (branch, merge, tag, push).

### Plugin system

Plugins implement `core.Plugin` interface (ReadVersion, WriteVersion, VersionFileName, VersionQualifier, RequiredTools). They self-register via `init()` functions using `core.RegisterPlugin()`.

- `core/plugin/` — base `Plugin` struct, `Config`, and `Factory` that injects the global `HookRegistry`
- `plugin/standard/` — fallback plugin using `version.txt` (also registered via `RegisterFallbackPlugin`)
- `plugin/mvn/` — Maven (`pom.xml`)
- `plugin/npm/` — npm (`package.json`)
- `plugin/composer/` — Composer (`composer.json`)
- `plugin/road/` — road manifest (`road.yaml`)
- `plugin/python/` — Python (not yet implemented, import disabled in `plugin/plugin.go`)

Plugin detection: iterates `pluginRegistry` in order, first plugin whose version file exists in the project wins. Falls back to `standard` plugin.

### Hook system

`core/hook.go` defines `HookRegistry` with typed hooks (`ReleaseStartHooks`, `HotfixStartHooks`, `HotfixFinishHooks`). Plugins register hooks during `init()` via `Plugin.RegisterHook()`. Hooks run at specific workflow points (e.g., before release start, after merge into develop).

### Repository abstraction

`core/repository.go` — `Repository` interface wraps all git operations (checkout, merge, tag, push, undo). Every method shells out to `git` via `exec.Command`. The `UndoAllChanges` method resets the repo to remote state when `undo: true` is configured.

### Version handling

`core/version.go` — `Version` struct with Major/Minor/Incremental/Qualifier. `ParseVersion` uses regex `(\d+)\.(\d+)\.(\d+)(?:-(\w+))?$`. Version increment logic: `Next(Minor)` bumps minor and resets incremental to 0; `Next(Incremental)` bumps patch.

### Configuration

Reads `$HOME/.gitflow-cli.yaml` (or `--config` flag) via Viper. Settings under `core:` key configure branch names (production, development, release, hotfix prefixes), undo behavior, and logging output.

## Adding a new plugin

1. Create `plugin/<name>/<name>.go`
2. Define a `Config` with name, version file, qualifier, required tools
3. Embed `plugin.Plugin` from `core/plugin`
4. Implement `ReadVersion` and `WriteVersion`
5. In `init()`: use `plugin.NewFactory().NewPlugin(config)`, register hooks, call `core.RegisterPlugin()`
6. Add import to `plugin/plugin.go`
7. Add e2e template in `e2e/helper/templates/<file>.tpl` and test cases in `e2e/workflow/`

## License headers

All `.go` files require the SPDX header:
```
/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/
```
