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

// Gets or creates the default JSON configuration Provider instance (Singleton) with default Options
func JsonConfigurationProvider() *jsonConfigurationProvider {
	if jsonConfigurationProviderDefaultInstance == nil {
		return JsonConfigurationProviderWithOptions(defaultJsonConfigurationProviderOptions)
	}
	return jsonConfigurationProviderDefaultInstance

}

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

// Creates a new JSON configuration Provider
func NewJsonConfigurationProvider() *jsonConfigurationProvider {
	return NewJsonConfigurationProviderWithOptions(defaultJsonConfigurationProviderOptions)
}

// Creates a new JSON configuration Provider with given options
func NewJsonConfigurationProviderWithOptions(options JsonConfigurationProviderOptions) *jsonConfigurationProvider {
	jcp := &jsonConfigurationProvider{
		options: options,
		json:    make(map[string]interface{}),
	}
	if options.FileFromCml {
		if v := CmlArgumentsProvider().Get(options.CmlSwitch); v != nil {
			jcp.options.Filename = v.(string)
		}
	}
	_ = jcp.Refresh()
	return jcp
}

func (jcp *jsonConfigurationProvider) Refresh() error {
	if b, e := ioutil.ReadFile(jcp.options.Filename); e != nil {
		log.Print("Unable to read json file : \"%v\"", e)
		return e
	} else if e = json.Unmarshal(b, &jcp.json); e != nil {
		return e
	}
	return nil
}

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
