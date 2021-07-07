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
	settings  Settings
}{
	variables: make(map[string]*variable),
	settings: Settings{
		DefaultProviderChain: []Provider{
			CmlArgumentsProvider(),
			JsonConfigurationProvider(),
			YamlConfigurationProvider(),
			EnvironmentVariablesProvider(),
		},
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
	IgnoreRequired       bool
	DefaultProviderChain []Provider
	DefaultSources       []Source
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
	for _, provider := range env.settings.DefaultProviderChain {
		if e := provider.Load(); e != nil {
			result = append(result, e)
		}
	}

	return result
}

// // Validate
// // validates if all non-string properties have been provided by a suitable format
// func Validate() []error {
// 	var result []error
// 	for name, property := range env.properties {
// 		if property.required && GetProperty(name) == nil {
// 			result = append(result, err.Error("Property "+name+" not provided!"))
// 		}
// 	}
// 	if result != nil && !env.settings.IgnoreRequired {
// 		panic(result)
// 	}
// 	return result
// }

// Get
// Gets the value of a variable if it's provided. Returns nil if not.
func Get(name string) interface{} {
	var v, found = env.variables[name]

	if found {
		v.mutex.Lock()
		defer v.mutex.Unlock()
		if v.value != nil {
			return v.value
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
		v.value = value
	}

	// update variable value
	return value
}

// Refresh
// Refreshes Provider configurations. A provider does not need to guarantee a
// refresh, but should have an error-free implementation then.
func Refresh() []error {
	var result []error
	for _, provider := range env.settings.DefaultProviderChain {
		if e := provider.Refresh(); e != nil {
			result = append(result, e)
		}
	}

	return nil
}

func SetDefaultChain(provider ...Provider) {
	env.settings.DefaultProviderChain = provider
}
