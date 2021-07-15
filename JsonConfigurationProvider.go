// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type jsonConfigurationProvider struct {
	options   JsonConfigurationProviderOptions
	timestamp time.Time
	lock      sync.Mutex
	json      *map[string]interface{}
}

type jsonConfigurationSource struct {
	provider *jsonConfigurationProvider
	name     *string
}

func (jcs *jsonConfigurationSource) Provider() Provider {
	return jcs.provider
}

func (jcs *jsonConfigurationSource) Config() interface{} {
	return jcs
}

func (jcs *jsonConfigurationSource) Name(name string) *jsonConfigurationSource {
	jcs.name = &name
	return jcs
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

// NewJsonConfigurationProviderWithOptions
// Creates a new JSON configuration Provider with given options
func NewJsonConfigurationProviderWithOptions(options JsonConfigurationProviderOptions) *jsonConfigurationProvider {
	jcp := &jsonConfigurationProvider{
		options: options,
	}
	_ = jcp.Load()
	return jcp
}

func JsonConfigurationSource() *jsonConfigurationSource {
	return &jsonConfigurationSource{
		provider: JsonConfigurationProvider(),
	}
}

// Load
// Loads the json configuration file. This is the only time when the filename is
// resolved as the source is not expected to change for a refresh.
func (jcp *jsonConfigurationProvider) Load() error {
	if jcp.options.FileFromCml {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlSwitch, nil); v != nil {
			jcp.options.Filename = v.(string)
		} else {
			jcp.options.Filename = ""
		}
	}
	_, e := jcp.Refresh()
	return e
}

// Refresh
// Reloads the configuration file. If no json file is configured, it is a nil operation.
func (jcp *jsonConfigurationProvider) Refresh() (bool, error) {
	if jcp.options.Filename != "" {
		stat, e := os.Stat(jcp.options.Filename)
		if e != nil {
			return false, e
		}
		jcp.lock.Lock()
		defer jcp.lock.Unlock()
		if !stat.ModTime().After(jcp.timestamp) {
			return false, nil
		}
		if b, e := ioutil.ReadFile(jcp.options.Filename); e != nil {
			log.Printf("Unable to read json file : \"%v\"", e)
			return false, e
		} else {
			jsonObject := make(map[string]interface{})
			if e = json.Unmarshal(b, &jsonObject); e != nil {
				return false, e
			}
			jcp.json = &jsonObject
			return true, nil
		}
	}
	return false, nil
}

// Get
// Gets the given property from the json file, if available.
func (jcp *jsonConfigurationProvider) Get(name string, config interface{}) interface{} {
	// If no json has been loaded, let's just return nil
	if jcp.json == nil {
		return nil
	}

	variableName := name
	// let's check if a configuration is passed and if it's the right type
	if config != nil {
		if source, isType := config.(*jsonConfigurationSource); isType {
			if source.name != nil {
				variableName = *source.name
			}
		}
	}

	// first check if we allow cml override, and if we do, try to get it from there
	if jcp.options.CmlPropertyOverride {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlPropertyOverrideSwitch+variableName, nil); v != nil {
			return v
		}
	}
	// first split the parcels from the dot notation
	parcels := strings.Split(variableName, ".")
	var currentBlock *map[string]interface{}
	var currentValue interface{} = *jcp.json
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
