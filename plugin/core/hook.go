/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

// HookType defines the different hook types
type HookType string

const (
	BeforeReleaseStartHook        HookType = "BeforeReleaseStart"
	AfterUpdateProjectVersionHook HookType = "AfterUpdateProjectVersion"
	// Add more hook types here
)

// HookFunction is the signature for hook functions
type HookFunction func() error

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

// Register registers a hook callback for a specific hook type
func (r *HookRegistry) Register(pluginName string, hookType HookType, fn HookFunction) {
	if _, exists := r.hooks[hookType]; !exists {
		r.hooks[hookType] = make(map[string]HookFunction)
	}
	r.hooks[hookType][pluginName] = fn
}

// Execute runs a hook if it is registered for the specified plugin
func (r *HookRegistry) Execute(pluginName string, hookType HookType) error {
	if fn, ok := r.hooks[hookType][pluginName]; ok {
		return fn()
	}
	return nil
}

// GlobalHooks is the global hook registry
var GlobalHooks = NewHookRegistry()
