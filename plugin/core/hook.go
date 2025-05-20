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

// Execute executes the registered hook of a specific type for a specific plugin
func (r *HookRegistry) Execute(pluginName string, hookType HookType) error {
	if handlers, exists := r.hooks[hookType]; exists {
		if fn, ok := handlers[pluginName]; ok {
			return fn()
		}
	}
	return nil
}

// HasHook checks if a hook of a specific type is registered for a plugin
func (r *HookRegistry) HasHook(pluginName string, hookType HookType) bool {
	if handlers, exists := r.hooks[hookType]; exists {
		_, ok := handlers[pluginName]
		return ok
	}
	return false
}

// GlobalHooks is the global hook registry
var GlobalHooks = NewHookRegistry()
