// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"os"
)

type environmentVariablesProvider struct{}

type environmentVariablesSource struct {
	name *string
}

func (evs *environmentVariablesSource) Provider() Provider {
	return environmentVariablesProviderInstance
}

func (evs *environmentVariablesSource) Config() interface{} {
	return evs
}

func (evs *environmentVariablesSource) Name(name string) *environmentVariablesSource {
	evs.name = &name
	return evs
}

var environmentVariablesProviderInstance = &environmentVariablesProvider{}

// EnvironmentVariablesProvider
// Gets the Environment Variables Provider instance (Singleton)
func EnvironmentVariablesProvider() *environmentVariablesProvider {
	return environmentVariablesProviderInstance
}

func EnvironmentVariablesSource() *environmentVariablesSource {
	return &environmentVariablesSource{}
}

// Load
// Loads the environment variables. This is a nil operation as the environment variables are always taken directly from os calls
func (evp *environmentVariablesProvider) Load() error {
	return nil
}

// Refresh
// Refreshes the environment variables. This is a nil operation as the environment variables are always taken directly from os calls
func (evp *environmentVariablesProvider) Refresh() (bool, error) {
	return false, nil
}

func (evp *environmentVariablesProvider) Get(name string, config interface{}) interface{} {
	variableName := name
	if source, isType := config.(*environmentVariablesSource); isType {
		if source.name != nil {
			variableName = *source.name
		}
	}

	if v, found := os.LookupEnv(variableName); found {
		return v
	}
	return nil
}
