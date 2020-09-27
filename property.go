// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

type property struct {
	name          string
	aliases       []string
	defaultValue  interface{}
	providerChain *[]Provider
	converter     func(value interface{}) interface{}
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
