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

var env = &struct {
	variables map[string]*variable
	providers map[Provider]bool
	settings  Settings
}{
	variables: make(map[string]*variable),
	providers: map[Provider]bool{
		CmlArgumentsProvider():         true,
		JsonConfigurationProvider():    true,
		YamlConfigurationProvider():    true,
		EnvironmentVariablesProvider(): true,
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

var lock = sync.Mutex{}

type Settings struct {
	FailOnMissingRequired bool
	DefaultSources        []Source
}

func addVar(v *variable) error {
	lock.Lock()
	defer lock.Unlock()

	if _, found := env.variables[v.name]; found {
		return ErrVariableAlreadyExists
	}

	env.variables[v.name] = v

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
	errors := err.NewErrors()
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
	var v, found = env.variables[name]

	if found {
		v.mutex.Lock()
		defer v.mutex.Unlock()
		if v.cachedValue != nil {
			return v.cachedValue.value
		}
	}

	var sources *[]Source
	var convert func(value interface{}) interface{}
	if !found || len(v.sources) == 0 {
		// either the variable was not found or no sources were defined
		sources = &env.settings.DefaultSources
		convert = nil
	} else {
		sources = &v.sources
		convert = v.converter
	}

	// search the provider chain for the property
	var value interface{}
	for _, source := range *sources {
		value = source.Provider().Get(name, source.Config())
		if value != nil {
			if convert != nil {
				value = convert(value)
			}
			break
		}
	}

	// value not found, check if it's a defined variable to get the default value
	if found {
		if value == nil {
			value = v.defaultValue
		}
		v.cachedValue = &valuePlaceholder{value: value}
	}

	// update variable value
	return value
}

// Refresh
// Refreshes Provider configurations. A provider does not need to guarantee a
// refresh, but should have an error-free implementation then.
func Refresh() []error {
	var result []error
	for provider := range env.providers {
		if e := provider.Refresh(); e != nil {
			result = append(result, e)
		}
	}

	return nil
}
