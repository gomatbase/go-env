// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import "sync"

type variable struct {
	name             string
	required         bool
	aliases          []string
	defaultValue     interface{}
	providerRefChain []providerRef
	chain            []Provider
	converter        func(value interface{}) interface{}
	mutex            sync.Mutex
}

func Var(name string) *variable {
	return &variable{name: name}
}

func (v *variable) Default(defaultValue interface{}) *variable {
	v.defaultValue = defaultValue
	return v
}

func (v *variable) Add() error {
	return addVar(v)
}

func (v *variable) providerChain() *[]Provider {
	if v.providerRefChain != nil && v.chain == nil {
		v.mutex.Lock()
		if v.chain == nil {
			v.chain = make([]Provider, len(v.providerRefChain))
			for i, p := range v.providerRefChain {
				v.chain[i] = p.provider
			}
		}
		v.mutex.Unlock()
	}
	return &v.chain
}
