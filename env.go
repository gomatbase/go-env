// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import "github.com/gomatbase/go-env/providers"

var defaultProviderChain = []Provider{providers.CmlArgumentsProvider(), providers.JsonConfigurationProvider(), providers.YamlConfigurationProvider(), providers.EnvironmentVariablesProvider()}

var env = &struct {
	properties map[string]*property
}{
	properties: make(map[string]*property),
}

func AddProperty(name string) *property {
	p := &property{name: name}
	env.properties[name] = p
	return p
}

// initializes environment with provided configuration
func Load() []error {
	var errors []error
	for _, provider := range defaultProviderChain {
		if e := provider.Load(); e != nil {
			errors = append(errors, e)
		}
	}

	return errors
}

// validates if all non-string properties have been provided by a suitable format
func Validate() {

}

// Gets the value of a property if it's provided. Returns nil if not.
func Get(name string) interface{} {
	p, hit := env.properties[name]
	var providerChain *[]Provider
	var aliases *[]string
	var convert func(value interface{}) interface{}
	if !hit || p.providerChain == nil {
		// no property was configured, but we still follow the default chain (CML and then OS ENV)
		providerChain = &defaultProviderChain
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
	errors := []error{}
	for _, provider := range defaultProviderChain {
		if e := provider.Refresh(); e != nil {
			errors = append(errors, e)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
