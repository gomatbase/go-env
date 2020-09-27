// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package providers

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

type yamlConfigurationProvider struct {
	options YamlConfigurationProviderOptions
	yaml    map[interface{}]interface{}
}

type YamlConfigurationProviderOptions struct {
	FileFromCml               bool
	CmlSwitch                 string
	CmlPropertyOverride       bool
	CmlPropertyOverrideSwitch string
	Filename                  string
}

var defaultYamlConfigurationProviderOptions = YamlConfigurationProviderOptions{
	FileFromCml:               true,
	CmlSwitch:                 "y",
	CmlPropertyOverride:       true,
	CmlPropertyOverrideSwitch: "Y",
}

var yamlConfigurationProviderDefaultInstance *yamlConfigurationProvider
var ycpMutex = sync.Mutex{}

// Gets or creates the default YAML configuration Provider instance (Singleton) with default Options
func YamlConfigurationProvider() *yamlConfigurationProvider {
	if yamlConfigurationProviderDefaultInstance == nil {
		return YamlConfigurationProviderWithOptions(defaultYamlConfigurationProviderOptions)
	}
	return yamlConfigurationProviderDefaultInstance

}

// Gets or creates the default JSON Configuration Provider instance (Singleton) with given Options. Options are
// ignored if there is already a default instance initialized
func YamlConfigurationProviderWithOptions(options YamlConfigurationProviderOptions) *yamlConfigurationProvider {
	if yamlConfigurationProviderDefaultInstance == nil {
		ycpMutex.Lock() // lock only for the moment where the default instance might be updated
		if yamlConfigurationProviderDefaultInstance == nil {
			yamlConfigurationProviderDefaultInstance = NewYamlConfigurationProviderWithOptions(options)
		}
		ycpMutex.Unlock()
	}
	return yamlConfigurationProviderDefaultInstance
}

// Creates a new JSON configuration Provider
func NewYamlConfigurationProvider() *yamlConfigurationProvider {
	return NewYamlConfigurationProviderWithOptions(defaultYamlConfigurationProviderOptions)
}

// Creates a new JSON configuration Provider with given options
func NewYamlConfigurationProviderWithOptions(options YamlConfigurationProviderOptions) *yamlConfigurationProvider {
	ycp := &yamlConfigurationProvider{
		options: options,
	}
	_ = ycp.Refresh()
	return ycp
}

// Loads the yaml configuration file. This is the only time when the filename is
// resolved as the source is not expected to change for a refresh.
func (ycp *yamlConfigurationProvider) Load() error {
	if ycp.options.FileFromCml {
		if v := CmlArgumentsProvider().Get(ycp.options.CmlSwitch); v != nil {
			ycp.options.Filename = v.(string)
		} else {
			ycp.options.Filename = ""
		}
		ycp.yaml = make(map[interface{}]interface{})
	}
	return ycp.Refresh()
}

// Reloads the configuration file. If no yaml file is configured, it is a nil operation.
func (ycp *yamlConfigurationProvider) Refresh() error {
	if ycp.options.Filename != "" {
		if b, e := ioutil.ReadFile(ycp.options.Filename); e != nil {
			log.Printf("Unable to read yaml file : \"%v\"", e)
			ycp.yaml = make(map[interface{}]interface{})
			return e
		} else if e = yaml.Unmarshal(b, &ycp.yaml); e != nil {
			return e
		}
	}
	return nil
}

// Gets the given property if available.
func (ycp *yamlConfigurationProvider) Get(name string) interface{} {
	// first check if we allow cml override, and if we do, try to get it from there
	if ycp.options.CmlPropertyOverride {
		if v := CmlArgumentsProvider().Get(ycp.options.CmlPropertyOverrideSwitch + name); v != nil {
			return v
		}
	}
	// first split the parcels from the dot notation
	parcels := strings.Split(name, ".")
	var currentBlock *map[interface{}]interface{}
	var currentValue interface{} = ycp.yaml
	for _, p := range parcels {
		if b, isType := currentValue.(map[interface{}]interface{}); !isType {
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
