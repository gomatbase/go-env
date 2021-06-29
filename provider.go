// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

// A Provider is a component that is able to extract values from a specific source, when present. They can
// be registered in the env package as a source of values
type Provider interface {
	// Get the value for the given property. nil should be returned if no property is found or an error with it exists
	Get(name string) interface{}

	// Load Loads the values it should provide
	Load() error

	// Refresh the provider sources. Useful for sources which are mutable during a single execution (like file sources)
	Refresh() error
}

// Source of a variable identifies the provider where the value will come from and the variable configuration to extract
// the value from the provider
type Source interface {
	// Provider returns the variable provider
	Provider() Provider
	// Config is the provider configuration used to extract the variable from the provider. The returned object should
	// be defined and processed by the provider
	Config() interface{}
}
