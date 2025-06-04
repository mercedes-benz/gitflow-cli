/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

// HookType defines the different hook types
type HookType string

// ReleaseStartHooks groups all hooks for the ReleaseStart workflow
var ReleaseStartHooks = struct {
	BeforeReleaseStartHook HookType
	AfterWriteVersionHook  HookType
}{
	BeforeReleaseStartHook: "ReleaseStart_BeforeReleaseStartHook",
	AfterWriteVersionHook:  "ReleaseStart_AfterWriteVersionHook",
}

// HotfixStartHooks groups all hooks for the HotfixStart workflow
var HotfixStartHooks = struct {
	BeforeHotfixStartHook HookType
}{
	BeforeHotfixStartHook: "HotfixStart_BeforeHotfixStartHook",
}

// HotfixFinishHooks groups all hooks for the HotfixFinish workflow
var HotfixFinishHooks = struct {
	AfterMergeIntoDevelopmentHook HookType
}{
	AfterMergeIntoDevelopmentHook: "HotfixFinish_AfterMergeIntoDevelopmentHook",
}

// HookFunction is the signature for hook functions
type HookFunction func(repository Repository) error

// HookRegistry manages the registration and execution of hooks
type HookRegistry struct {
	hooks map[HookType]map[string]HookFunction
}

// NewHookRegistry creates a new hook registry
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookType]map[string]HookFunction),
	}
}

// RegisterHook registers a hook callback for a specific hook type
func (r *HookRegistry) RegisterHook(pluginName string, hookType HookType, hookFunction HookFunction) {
	if _, exists := r.hooks[hookType]; !exists {
		r.hooks[hookType] = make(map[string]HookFunction)
	}
	r.hooks[hookType][pluginName] = hookFunction
}

// ExecuteHook runs a hook if it is registered for the specified plugin
func (r *HookRegistry) ExecuteHook(plugin Plugin, hookType HookType, repository Repository) error {
	if hookFunction, ok := r.hooks[hookType][plugin.String()]; ok {
		return hookFunction(repository)
	}
	return nil
}

// GlobalHooks is the global hook registry
var GlobalHooks = NewHookRegistry()
