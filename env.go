// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import (
	"sync"

	"github.com/gomatbase/go-env/providers"
	"github.com/gomatbase/go-error"
)

const (
	ErrVariableAlreadyExists = err.Error("Variable name already exists.")
)

var env = &struct {
	properties map[string]*property
	variables  map[string]*variable
	settings   Settings
}{
	properties: make(map[string]*property),
	variables:  make(map[string]*variable),
	settings: Settings{
		DefaultProviderChain: []Provider{
			providers.CmlArgumentsProvider(),
			providers.JsonConfigurationProvider(),
			providers.YamlConfigurationProvider(),
			providers.EnvironmentVariablesProvider(),
		},
	},
}

var lock = sync.Mutex{}

type Settings struct {
	IgnoreRequired       bool
	DefaultProviderChain []Provider
}

func AddProperty(name string) *property {
	p := &property{
		name:    name,
		aliases: []string{name},
	}
	env.properties[name] = p
	return p
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

// Validate
// validates if all non-string properties have been provided by a suitable format
func Validate() []error {
	var result []error
	for name, property := range env.properties {
		if property.required && GetProperty(name) == nil {
			result = append(result, err.Error("Property "+name+" not provided!"))
		}
	}
	if result != nil && !env.settings.IgnoreRequired {
		panic(result)
	}
	return result
}

// GetProperty
// Gets the value of a property if it's provided. Returns nil if not.
func GetProperty(name string) interface{} {
	var p, hit = env.properties[name]
	var providerChain *[]Provider
	var aliases *[]string
	var convert func(value interface{}) interface{}
	if !hit || p.providerRefChain == nil {
		// no property was configured, but we still follow the default chain (CML and then OS ENV)
		providerChain = &env.settings.DefaultProviderChain
		aliases = &[]string{name}
		convert = nil
	} else {
		providerChain = p.providerChain()
		aliases = &p.aliases
		convert = p.converter
	}

	// search the provider chain for the property
	var value interface{}
	for _, provider := range *providerChain {
		for _, alias := range *aliases {
			value = provider.Get(alias)
			if value != nil {
				if convert != nil {
					value = convert(value)
				}
				return value
			}
		}
	}

	// value not found, check if it's a defined property to get the default value
	if p != nil {
		return p.defaultValue
	}

	// nothing found and no default value
	return nil
}

// Get
// Gets the value of a variable if it's provided. Returns nil if not.
func Get(name string) interface{} {
	var v, hit = env.variables[name]
	var providerChain *[]Provider
	var aliases *[]string
	var convert func(value interface{}) interface{}
	if !hit || v.providerRefChain == nil {
		// no property was configured, but we still follow the default chain (CML and then OS ENV)
		providerChain = &env.settings.DefaultProviderChain
		aliases = &[]string{name}
		convert = nil
	} else {
		providerChain = v.providerChain()
		aliases = &v.aliases
		convert = v.converter
	}

	// search the provider chain for the property
	var value interface{}
	for _, provider := range *providerChain {
		for _, alias := range *aliases {
			value = provider.Get(alias)
			if value != nil {
				if convert != nil {
					value = convert(value)
				}
				return value
			}
		}
	}

	// value not found, check if it's a defined property to get the default value
	if v != nil {
		return v.defaultValue
	}

	// nothing found and no default value
	return nil
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
