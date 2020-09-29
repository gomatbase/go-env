// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import (
	"errors"
	"github.com/gomatbase/go-env/providers"
)

var env = &struct {
	properties map[string]*property
	settings   Settings
}{
	properties: make(map[string]*property),
	settings: Settings{
		DefaultProviderChain: []Provider{
			providers.CmlArgumentsProvider(),
			providers.JsonConfigurationProvider(),
			providers.YamlConfigurationProvider(),
			providers.EnvironmentVariablesProvider(),
		},
	},
}

type Settings struct {
	IgnoreRequired       bool
	DefaultProviderChain []Provider
}

func AddProperty(name string) *property {
	p := &property{name: name}
	env.properties[name] = p
	return p
}

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

// validates if all non-string properties have been provided by a suitable format
func Validate() []error {
	var result []error
	for name, property := range env.properties {
		if property.required && Get(name) == nil {
			result = append(result, errors.New("Property "+name+" not provided!"))
		}
	}
	if env.settings.IgnoreRequired {
		return result
	}
	panic(result)
}

// Gets the value of a property if it's provided. Returns nil if not.
func Get(name string) interface{} {
	p, hit := env.properties[name]
	var providerChain *[]Provider
	var aliases *[]string
	var convert func(value interface{}) interface{}
	if !hit || p.providerChain == nil {
		// no property was configured, but we still follow the default chain (CML and then OS ENV)
		providerChain = &env.settings.DefaultProviderChain
		aliases = &[]string{name}
		convert = nil
	} else {
		providerChain = p.providerChain
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
