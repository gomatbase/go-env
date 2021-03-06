// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package providers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

type jsonConfigurationProvider struct {
	options JsonConfigurationProviderOptions
	json    map[string]interface{}
}

type JsonConfigurationProviderOptions struct {
	FileFromCml               bool
	CmlSwitch                 string
	CmlPropertyOverride       bool
	CmlPropertyOverrideSwitch string
	Filename                  string
}

var defaultJsonConfigurationProviderOptions = JsonConfigurationProviderOptions{
	FileFromCml:               true,
	CmlSwitch:                 "j",
	CmlPropertyOverride:       true,
	CmlPropertyOverrideSwitch: "J",
}

var jsonConfigurationProviderDefaultInstance *jsonConfigurationProvider
var jcpMutex = sync.Mutex{}

// JsonConfigurationProvider
// Gets or creates the default JSON configuration Provider instance (Singleton) with default Options
func JsonConfigurationProvider() *jsonConfigurationProvider {
	if jsonConfigurationProviderDefaultInstance == nil {
		return JsonConfigurationProviderWithOptions(defaultJsonConfigurationProviderOptions)
	}
	return jsonConfigurationProviderDefaultInstance

}

// JsonConfigurationProviderWithOptions
// Gets or creates the default JSON Configuration Provider instance (Singleton) with given Options. Options are
// ignored if there is already a default instance initialized
func JsonConfigurationProviderWithOptions(options JsonConfigurationProviderOptions) *jsonConfigurationProvider {
	if jsonConfigurationProviderDefaultInstance == nil {
		jcpMutex.Lock() // lock only for the moment where the default instance might be updated
		if jsonConfigurationProviderDefaultInstance == nil {
			jsonConfigurationProviderDefaultInstance = NewJsonConfigurationProviderWithOptions(options)
		}
		jcpMutex.Unlock()
	}
	return jsonConfigurationProviderDefaultInstance
}

// NewJsonConfigurationProvider
// Creates a new JSON configuration Provider
func NewJsonConfigurationProvider() *jsonConfigurationProvider {
	return NewJsonConfigurationProviderWithOptions(defaultJsonConfigurationProviderOptions)
}

// NewJsonConfigurationProviderWithOptions
// Creates a new JSON configuration Provider with given options
func NewJsonConfigurationProviderWithOptions(options JsonConfigurationProviderOptions) *jsonConfigurationProvider {
	jcp := &jsonConfigurationProvider{
		options: options,
	}
	_ = jcp.Load()
	return jcp
}

// Load
// Loads the json configuration file. This is the only time when the filename is
// resolved as the source is not expected to change for a refresh.
func (jcp *jsonConfigurationProvider) Load() error {
	if jcp.options.FileFromCml {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlSwitch); v != nil {
			jcp.options.Filename = v.(string)
		} else {
			jcp.options.Filename = ""
		}
		jcp.json = make(map[string]interface{})
	}
	return jcp.Refresh()
}

// Refresh
// Reloads the configuration file. If no json file is configured, it is a nil operation.
func (jcp *jsonConfigurationProvider) Refresh() error {
	if jcp.options.Filename != "" {
		if b, e := ioutil.ReadFile(jcp.options.Filename); e != nil {
			log.Printf("Unable to read json file : \"%v\"", e)
			jcp.json = make(map[string]interface{})
			return e
		} else if e = json.Unmarshal(b, &jcp.json); e != nil {
			return e
		}
	}
	return nil
}

// Get
// Gets the given property from the json file, if available.
func (jcp *jsonConfigurationProvider) Get(name string) interface{} {
	// first check if we allow cml override, and if we do, try to get it from there
	if jcp.options.CmlPropertyOverride {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlPropertyOverrideSwitch + name); v != nil {
			return v
		}
	}
	// first split the parcels from the dot notation
	parcels := strings.Split(name, ".")
	var currentBlock *map[string]interface{}
	var currentValue interface{} = jcp.json
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
