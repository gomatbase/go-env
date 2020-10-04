// Copyright 2020 GOM. All rights reserved.
// Since 04/10/2020 By GOM
// Licensed under MIT License

// Provider wrapper for cfenv.
package cfproviders

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
)

type cfEnvironmentProvider struct {
	cfEnvironment   map[string]interface{}
	isCfEnvironment bool
}

var cfEnvironmentProviderDefaultInstance *cfEnvironmentProvider
var cfeMutex = sync.Mutex{}

func CfEnvironmentProvider() *cfEnvironmentProvider {
	if cfEnvironmentProviderDefaultInstance == nil {
		cfeMutex.Lock() // lock only for the moment where the default instance might be updated
		if cfEnvironmentProviderDefaultInstance == nil {
			cfEnvironmentProviderDefaultInstance = NewCfEnvironmentProvider()
		}
		cfeMutex.Unlock()
	}
	return cfEnvironmentProviderDefaultInstance
}

func NewCfEnvironmentProvider() *cfEnvironmentProvider {
	cfe := &cfEnvironmentProvider{}
	return cfe
}

func (cfe *cfEnvironmentProvider) Load() error {
	cfEnvironment := make(map[string]interface{})
	cfEnvironment["vcap"] = make(map[string]interface{})

	if vcapApplication, found := os.LookupEnv("VCAP_APPLICATION"); !found {
		return errors.New("Not running in CF")
	} else {
		var vcapApplicationMap = make(map[string]interface{})
		if e := json.Unmarshal([]byte(vcapApplication), vcapApplicationMap); e != nil {
			return errors.New("Unable to parse VCAP_APPLICATION")
		}
		cfEnvironment["vcap"].(map[string]interface{})["application"] = vcapApplicationMap
	}

	if vcapServices, found := os.LookupEnv("VCAP_SERVICES"); found {
		var vcapServicesMap = make(map[string]interface{})
		if e := json.Unmarshal([]byte(vcapServices), vcapServicesMap); e != nil {
			return errors.New("Unable to parse VCAP_SERVICES")
		}
		cfEnvironment["vcap"].(map[string]interface{})["services"] = vcapServicesMap
	}

	// we only need to initialize/parse the vcap variables. Everything else will be service by the Environment Provider
	cfe.cfEnvironment = cfEnvironment

	return nil
}

func (cfe *cfEnvironmentProvider) Refresh() error {
	if cfe.cfEnvironment == nil {
		errors.New("No CF Environment Initialized")
	}
	// we only initialize the wrapper once.
	return nil
}

func (cfe *cfEnvironmentProvider) Get(name string) interface{} {
	// we only use the wrapper mainly to get the vcap variables, all the rest can be taken from the environment directly.
	if cfe.cfEnvironment == nil {
		errors.New("No CF Environment Initialized")
	}

	// first split the parcels from the dot notation
	parcels := strings.Split(name, ".")
	var currentBlock *map[string]interface{}
	var currentValue interface{} = cfe.cfEnvironment
	for _, p := range parcels {
		if b, isType := currentValue.(map[string]interface{}); !isType {
			return nil
		} else {
			currentBlock = &b
		}
		if v, found := (*currentBlock)[p]; !found {
			return nil
		} else {
			currentValue = v
		}

	}
	return currentValue
}
