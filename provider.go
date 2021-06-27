// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

// A Provider is a component that is able to extract values from a specific source, when present. They can
// be registered in the env package as a source of values
type Provider interface {
	// GetProperty the value for the given property. nil should be returned if no property is found or an error with it exists
	Get(name string) interface{}

	// Load Loads the values it should provide
	Load() error

	// Refresh the provider sources. Useful for sources which are mutable during a single execution (like file sources)
	Refresh() error
}
