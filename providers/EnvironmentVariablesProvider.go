// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package providers

import (
	"os"
	"sync"
)

type environmentVariablesProvider struct{}

var environmentVariablesProviderDefaultInstance *environmentVariablesProvider
var evpMutex = sync.Mutex{}

// Gets or creates the default Environment Variables Provider instance (Singleton)
func EnvironmentVariablesProvider() *environmentVariablesProvider {
	if environmentVariablesProviderDefaultInstance == nil {
		evpMutex.Lock() // lock only for the moment where the default instance might be updated
		if environmentVariablesProviderDefaultInstance == nil {
			environmentVariablesProviderDefaultInstance = NewEnvironmentVariablesProvider()
		}
		evpMutex.Unlock()
	}
	return environmentVariablesProviderDefaultInstance

}

// Creates a new Environment Variables Provider
func NewEnvironmentVariablesProvider() *environmentVariablesProvider {
	evp := &environmentVariablesProvider{}
	return evp
}

// Loads the environment variables. This is a nil operation as the environment variables are always taken directly from os calls
func (evp *environmentVariablesProvider) Load() error {
	return nil
}

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
