// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"os"
	"strings"
	"sync"
)

type cmlArgumentsProvider struct {
	args     []string
	switches map[string]string
}

type cmlArgumentsSource struct {
	provider *cmlArgumentsProvider
}

func (cmlas *cmlArgumentsSource) Provider() Provider {
	return cmlas.provider
}

func (cmlas *cmlArgumentsSource) Config() interface{} {
	return cmlas
}

const (
	cmlapSTART = iota
	cmlapSWITCH
	cmlapVALUE
)

var cmlArgumentsProviderDefaultInstance *cmlArgumentsProvider
var cmlapMutex = sync.Mutex{}

// CmlArgumentsProvider
// Gets or creates a singleton instance for the CML Arguments Provider
func CmlArgumentsProvider() *cmlArgumentsProvider {
	if cmlArgumentsProviderDefaultInstance == nil {
		cmlapMutex.Lock() // lock only for the moment where the default instance might be updated
		if cmlArgumentsProviderDefaultInstance == nil {
			cmlArgumentsProviderDefaultInstance = NewCmlArgumentsProvider()
		}
		cmlapMutex.Unlock()
	}
	return cmlArgumentsProviderDefaultInstance
}

func CmlArgumentsSource() *cmlArgumentsSource {
	return &cmlArgumentsSource{
		provider: CmlArgumentsProvider(),
	}
}

// NewCmlArgumentsProvider
// Creates a new CML Arguments Provider (doesn't affect default instance)
func NewCmlArgumentsProvider() *cmlArgumentsProvider {
	cmlap := &cmlArgumentsProvider{}
	if e := cmlap.Refresh(); e != nil {
		panic(e)
	} // errors should never happen
	return cmlap
}

// Get
// Gets the value of the given property, if defined.
func (cmlap *cmlArgumentsProvider) Get(name string) interface{} {
	v, found := cmlap.switches[name]
	if found {
		return v
	}
	return nil
}

// Load
// Parses the command line looking for switches and eventually assigning values
// to them. It supports normal switches (-) long named switches (--) and assigns
// single values when used with the assignment operator (=) or the whole value
// set after a space until the next switch or end of arguments.
func (cmlap *cmlArgumentsProvider) Load() error {
	cmlap.args = os.Args
	cmlap.switches = make(map[string]string)

	previousContext := cmlapSTART
	var currentSwitch string
	for _, arg := range cmlap.args[1:] {
		if arg[0] == '-' {
			// argument is a switch check if it's a long switch
			if arg[1] == '-' {
				currentSwitch = arg[2:]
			} else {
				currentSwitch = arg[1:]
			}

			// check if it's a variable
			if currentSwitch[0] == 'V' {
				currentSwitch = currentSwitch[1:]
			}

			// let's check if it holds an assignment
			if i := strings.IndexByte(currentSwitch, '='); i > 0 {
				currentValue := currentSwitch[i+1:]
				currentSwitch = currentSwitch[:i]
				cmlap.switches[currentSwitch] = currentValue
				previousContext = cmlapVALUE
			} else {
				cmlap.switches[currentSwitch] = ""
				previousContext = cmlapSWITCH
			}
		} else if previousContext == cmlapSWITCH {
			if len(cmlap.switches[currentSwitch]) == 0 {
				cmlap.switches[currentSwitch] = arg
			} else {
				cmlap.switches[currentSwitch] = cmlap.switches[currentSwitch] + " " + arg
			}
		}

		// non contextualized values are currently not indexed
	}
	return nil
}

// Refresh
// cml is fixed from start of application, no refresh is possible but we implement it anyway for testing purposes
func (cmlap *cmlArgumentsProvider) Refresh() error {
	return nil
}
