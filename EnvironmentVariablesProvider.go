// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"os"
	"sync"
)

type environmentVariablesProvider struct{}

type environmentVariablesSource struct{}

func (evs *environmentVariablesSource) Provider() Provider {
	return environmentVariablesProviderInstance
}

func (evs *environmentVariablesSource) Config() interface{} {
	return evs
}

var environmentVariablesProviderInstance = &environmentVariablesProvider{}
var evpMutex = sync.Mutex{}

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
func (evp *environmentVariablesProvider) Refresh() error {
	return nil
}

func (evp *environmentVariablesProvider) Get(name string) interface{} {
	if v, found := os.LookupEnv(name); found {
		return v
	}
	return nil
}
