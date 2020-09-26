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
	jcp := &yamlConfigurationProvider{
		options: options,
		yaml:    make(map[interface{}]interface{}),
	}
	_ = jcp.Refresh()
	return jcp
}

func (jcp *yamlConfigurationProvider) Refresh() error {
	if jcp.options.FileFromCml {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlSwitch); v != nil {
			jcp.options.Filename = v.(string)
		} else {
			jcp.options.Filename = ""
		}
	}
	if b, e := ioutil.ReadFile(jcp.options.Filename); e != nil {
		log.Printf("Unable to read yaml file : \"%v\"", e)
		jcp.yaml = make(map[interface{}]interface{})
		return e
	} else if e = yaml.Unmarshal(b, &jcp.yaml); e != nil {
		return e
	}
	return nil
}

func (jcp *yamlConfigurationProvider) Get(name string) interface{} {
	// first check if we allow cml override, and if we do, try to get it from there
	if jcp.options.CmlPropertyOverride {
		if v := CmlArgumentsProvider().Get(jcp.options.CmlPropertyOverrideSwitch + name); v != nil {
			return v
		}
	}
	// first split the parcels from the dot notation
	parcels := strings.Split(name, ".")
	var currentBlock *map[interface{}]interface{}
	var currentValue interface{} = jcp.yaml
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
