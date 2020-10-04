// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import "sync"

type property struct {
	name             string
	required         bool
	aliases          []string
	defaultValue     interface{}
	providerRefChain []providerRef
	chain            []Provider
	converter        func(value interface{}) interface{}
	mutex            sync.Mutex
}

func (p *property) providerChain() *[]Provider {
	if p.providerRefChain != nil && p.chain == nil {
		p.mutex.Lock()
		if p.chain == nil {
			p.chain = make([]Provider, len(p.providerRefChain))
			for i, v := range p.providerRefChain {
				p.chain[i] = v.provider
			}
		}
		p.mutex.Unlock()
	}
	return &p.chain
}

type providerRef struct {
	provider Provider
	settings interface{}
}

func (p *property) WithDefaultValue(defaultValue interface{}) *property {
	p.defaultValue = defaultValue
	return p
}

func (p *property) WithConverter(converter func(value interface{}) interface{}) *property {
	p.converter = converter
	return p
}

func (p *property) WithAliases(alias ...string) *property {
	p.aliases = append(p.aliases, alias...)
	return p
}

func (p *property) Required() *property {
	p.required = true
	return p
}

func (p *property) From(provider Provider) *property {
	p.providerRefChain = append(p.providerRefChain, providerRef{
		provider: provider,
	})
	return p
}

func (p *property) WithProviderSettings(settings interface{}) *property {
	if p.providerRefChain == nil {
		panic("Trying to add settings to non reference provider")
	}
	providerIndex := len(p.providerRefChain) - 1
	if p.providerRefChain[providerIndex].settings != nil {
		panic("Trying to set provider settings twice")
	}
	p.providerRefChain[providerIndex].settings = settings
	return p
}
