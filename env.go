// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import "github.com/gomatbase/go-env/providers"

var defaultProviderChain = []Provider{providers.CmlArgumentsProvider(), providers.EnvironmentVariablesProvider()}

type environment struct {
	properties map[string]*property
}

var env = &environment{
	properties: make(map[string]*property),
}

func AddProperty(name string) *property {
	p := &property{name: name}
	env.properties[name] = p
	return p
}

// initializes environment with provided configuration
func Build() {
	Refresh()
}

// validates if all non-string properties have been provided by a suitable format
func Validate() {

}

func Get(name string) interface{} {
	p, hit := env.properties[name]
	var providerChain *[]Provider
	if !hit || p.providerChain == nil {
		// no property was configured, but we still follow the default chain (CML and then OS ENV)
		providerChain = &defaultProviderChain
	} else {
		providerChain = p.providerChain
	}

	// first we try to get the value down the precedence provider chain
	var value interface{}
	for _, provider := range *providerChain {
		value = provider.Get(name)
		if value != nil {
			break
		}
	}

	// If not found, let's assign it the default value if provided
	if value == nil && p != nil {
		value = p.defaultValue
	}

	// then we try to convert it if a converter was provided
	if value != nil && p != nil && p.converter != nil {
		value = p.converter(value)
	}

	// and finally we return the provided value
	return value
}

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
