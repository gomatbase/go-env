// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package providers

import (
	"os"
	"strings"
	"sync"
)

type cmlArgumentsProvider struct {
	args     []string
	switches map[string]string
}

const (
	cmlapSTART = iota
	cmlapSWITCH
	cmlapVALUE
)

var cmlArgumentsProviderDefaultInstance *cmlArgumentsProvider
var cmlapMutex = sync.Mutex{}

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

// Creates a new CML Arguments Provider (doesn't affect default instance)
func NewCmlArgumentsProvider() *cmlArgumentsProvider {
	cmlap := &cmlArgumentsProvider{}
	if e := cmlap.Refresh(); e != nil {
		panic(e)
	} // errors should never happen
	return cmlap
}

func (cmlap *cmlArgumentsProvider) Get(name string) interface{} {
	v, found := cmlap.switches[name]
	if found {
		return v
	}
	return nil
}

func (cmlap *cmlArgumentsProvider) Refresh() error {
	// cml is fixed from start of application, no refresh is possible but we implement it anyway for testing purposes
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
