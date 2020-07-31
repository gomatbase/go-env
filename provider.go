// Copyright 2020 GOM. All rights reserved.
// Since 29/07/2020 By GOM
// Licensed under MIT License

package env

import "os"

type Provider interface {
	Get(names ...string) *string
}

type environmentVariablesProvider struct{}

func (evp *environmentVariablesProvider) Get(names ...string) *string {
	for _, name := range names {
		if value, found := os.LookupEnv(name); found {
			return &value
		}
	}
	return nil
}

type cmlArgumentsProvider struct {
	args []string
}

func (pp *cmlArgumentsProvider) Get(names ...string) *string {
	return nil
}
