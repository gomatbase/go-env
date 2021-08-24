// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import (
	"sync"

	"github.com/gomatbase/go-error"
)

const (
	ErrVariableAlreadyExists = err.Error("Variable name already exists.")
)

type providerRegistry struct {
	variables []*variable
	dirty     bool
	lock      sync.Mutex
}

var env = &struct {
	variables map[string]*variable
	providers map[Provider]*providerRegistry
	settings  Settings
}{
	variables: make(map[string]*variable),
	providers: map[Provider]*providerRegistry{
		CmlArgumentsProvider():         newProviderRegistry(),
		JsonConfigurationProvider():    newProviderRegistry(),
		YamlConfigurationProvider():    newProviderRegistry(),
		EnvironmentVariablesProvider(): newProviderRegistry(),
	},
	settings: Settings{
		DefaultSources: []Source{
			CmlArgumentsSource(),
			JsonConfigurationSource(),
			YamlConfigurationSource(),
			EnvironmentVariablesSource(),
		},
	},
}

func newProviderRegistry() *providerRegistry {
	return &providerRegistry{
		variables: []*variable{},
		dirty:     false,
		lock:      sync.Mutex{},
	}
}

var lock = sync.Mutex{}

type Settings struct {
	FailOnMissingRequired bool
	DefaultSources        []Source
}

func addVar(v *variable) error {
	lock.Lock()

	if _, found := env.variables[v.name]; found {
		lock.Unlock()
		return ErrVariableAlreadyExists
	}

	env.variables[v.name] = v
	lock.Unlock()

	// now let's check each of the providers, register unknown providers, and register the variable with its providers
	if len(v.sources) == 0 {
		// no specific sources provided, let's give it the default ones
		v.sources = make([]*source, len(env.settings.DefaultSources))
		for i, s := range env.settings.DefaultSources {
			v.sources[i] = &source{source: s}
		}
	} else {
		for _, s := range v.sources {
			lock.Lock()
			registry, found := env.providers[s.source.Provider()]
			if !found {
				registry = newProviderRegistry()
				env.providers[s.source.Provider()] = registry
			}
			lock.Unlock()
			registry.lock.Lock()
			registry.variables = append(registry.variables, v)
			registry.lock.Unlock()
		}
	}

	return nil
}

// Load
// initializes environment with provided configuration
func Load() []error {
	var result []error
	for provider := range env.providers {
		if e := provider.Load(); e != nil {
			result = append(result, e)
		}
	}

	return result
}

func FailOnMissingVariables(flag bool) {
	env.settings.FailOnMissingRequired = flag
}

// Validate
// validates if all non-string properties have been provided by a suitable format
func Validate() error {
	errors := err.Errors()
	for name, variable := range env.variables {
		if Get(name) == nil && variable.required {
			errors.AddError(err.Error("Property " + name + " not provided!"))
		}
	}
	if errors.Count() > 0 {
		if env.settings.FailOnMissingRequired {
			panic(errors)
		}
		return errors
	}
	return nil
}

// Get
// Gets the value of a variable if it's provided. Returns nil if not.
func Get(name string) interface{} {
	lock.Lock()
	var v, found = env.variables[name]
	lock.Unlock()

	if found {
		v.mutex.Lock()
		if v.cachedValue != nil {
			defer v.mutex.Unlock()
			return v.cachedValue.value
		}
		v.mutex.Unlock()
	}

	var value interface{}
	if !found {
		// it's for an ad-hoc value, let's go through the default chain
		for _, source := range env.settings.DefaultSources {
			value = source.Provider().Get(name, source.Config())
			if value != nil {
				break
			}
		}
	} else {
		// it's a variable
		for _, s := range v.sources {
			sourceValue := s.source.Provider().Get(name, s.source.Config())
			s.cachedValue = &valuePlaceholder{value: sourceValue} // cache the given value to identify if there were changes in a refresh
			if value == nil && sourceValue != nil {
				value = sourceValue
				if v.converter != nil {
					value = v.converter(value)
				}
			}
		}
		if value == nil {
			value = v.defaultValue
		}
		v.cachedValue = &valuePlaceholder{value: value}
	}

	return value
}

// Refresh
// Refreshes Provider configurations. A provider does not need to guarantee a
// refresh, but should have an error-free implementation then.
func Refresh() error {
	errors := err.Errors()
	lock.Lock()
	for provider, registry := range env.providers {
		if updated, e := provider.Refresh(); e != nil {
			errors.AddError(e)
		} else if updated {
			registry.dirty = true
		}
	}
	lock.Unlock()

	// GOM: needs to be improved... this is a brute-force approach which is ok for now.

	for _, v := range env.variables {
		v.mutex.Lock()
		if v.cachedValue != nil {
			// it was never retrieved, let's initialize the cached value prioritizing the first dirty provider
			v.cachedValue = &valuePlaceholder{}
			var dirtyValue interface{}
			for _, s := range v.sources {
				isDirtyProvider := env.providers[s.source.Provider()].dirty
				sourceValue := s.source.Provider().Get(v.name, s.source.Config())
				s.cachedValue = &valuePlaceholder{value: sourceValue}
				if sourceValue != nil && (v.cachedValue.value == nil || isDirtyProvider && dirtyValue == nil) {
					v.cachedValue.value = sourceValue
					if v.converter != nil {
						v.cachedValue.value = v.converter(v.cachedValue.value)
					}
					if isDirtyProvider && dirtyValue == nil {
						dirtyValue = sourceValue
					}
				}
			}
		} else {
			var newValue interface{}
			for _, s := range v.sources {
				if env.providers[s.source.Provider()].dirty {
					sourceValue := s.source.Provider().Get(v.name, s.source.Config())
					if sourceValue != s.cachedValue.value {
						s.cachedValue.value = sourceValue
						if newValue == nil {
							newValue = sourceValue
						}
					}
				}
			}
			if newValue != nil && newValue != v.cachedValue.value {
				v.cachedValue.value = newValue
				if v.converter != nil {
					v.cachedValue.value = v.converter(v.cachedValue.value)
				}
			}
		}
		v.mutex.Unlock()
	}

	lock.Lock()
	for _, registry := range env.providers {
		registry.dirty = false
	}
	lock.Unlock()

	if errors.Count() > 0 {
		return errors
	}

	return nil
}
