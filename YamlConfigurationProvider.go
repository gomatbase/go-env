// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type yamlConfigurationProvider struct {
	options   YamlConfigurationProviderOptions
	timestamp time.Time
	lock      sync.Mutex
	yaml      *map[interface{}]interface{}
}

type yamlConfigurationSource struct {
	provider *yamlConfigurationProvider
	name     *string
}

func (ycs *yamlConfigurationSource) Provider() Provider {
	return ycs.provider
}

func (ycs *yamlConfigurationSource) Config() interface{} {
	return ycs
}

func (ycs *yamlConfigurationSource) Name(name string) *yamlConfigurationSource {
	ycs.name = &name
	return ycs
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

// YamlConfigurationProvider
// Gets or creates the default YAML configuration Provider instance (Singleton) with default Options
func YamlConfigurationProvider() *yamlConfigurationProvider {
	if yamlConfigurationProviderDefaultInstance == nil {
		return YamlConfigurationProviderWithOptions(defaultYamlConfigurationProviderOptions)
	}
	return yamlConfigurationProviderDefaultInstance

}

// YamlConfigurationProviderWithOptions
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

// NewYamlConfigurationProvider
// Creates a new JSON configuration Provider
func NewYamlConfigurationProvider() *yamlConfigurationProvider {
	return NewYamlConfigurationProviderWithOptions(defaultYamlConfigurationProviderOptions)
}

// NewYamlConfigurationProviderWithOptions
// Creates a new Yaml configuration Provider with given options
func NewYamlConfigurationProviderWithOptions(options YamlConfigurationProviderOptions) *yamlConfigurationProvider {
	ycp := &yamlConfigurationProvider{
		options: options,
	}
	_ = ycp.Load()
	return ycp
}

func YamlConfigurationSource() *yamlConfigurationSource {
	return &yamlConfigurationSource{
		provider: YamlConfigurationProvider(),
	}
}

// Load
// Loads the yaml configuration file. This is the only time when the filename is
// resolved as the source is not expected to change for a refresh.
func (ycp *yamlConfigurationProvider) Load() error {
	if ycp.options.FileFromCml {
		if v := CmlArgumentsProvider().Get(ycp.options.CmlSwitch, nil); v != nil {
			ycp.options.Filename = v.(string)
		} else {
			ycp.options.Filename = ""
		}
	}
	_, e := ycp.Refresh()
	return e
}

// Refresh
// Reloads the configuration file. If no yaml file is configured, it is a nil operation.
func (ycp *yamlConfigurationProvider) Refresh() (bool, error) {
	if ycp.options.Filename != "" {
		stat, e := os.Stat(ycp.options.Filename)
		if e != nil {
			return false, e
		}
		ycp.lock.Lock()
		defer ycp.lock.Unlock()
		if !stat.ModTime().After(ycp.timestamp) {
			return false, nil
		}
		if b, e := ioutil.ReadFile(ycp.options.Filename); e != nil {
			log.Printf("Unable to read yaml file : \"%v\"", e)
			return false, e
		} else if e = yaml.Unmarshal(b, &ycp.yaml); e != nil {
			return false, e
		}
	}
	return false, nil
}

// Get
// Gets the given property if available.
func (ycp *yamlConfigurationProvider) Get(name string, config interface{}) interface{} {
	// If no yaml has been loaded, let's just return nil
	if ycp.yaml == nil {
		return nil
	}

	variableName := name
	// let's check if a configuration is passed and if it's the right type
	if config != nil {
		if source, isType := config.(*yamlConfigurationSource); isType {
			if source.name != nil {
				variableName = *source.name
			}
		}
	}

	// first check if we allow cml override, and if we do, try to get it from there
	if ycp.options.CmlPropertyOverride {
		if v := CmlArgumentsProvider().Get(ycp.options.CmlPropertyOverrideSwitch+variableName, nil); v != nil {
			return v
		}
	}
	// first split the parcels from the dot notation
	parcels := strings.Split(variableName, ".")
	var currentBlock *map[interface{}]interface{}
	var currentValue interface{} = *ycp.yaml
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
