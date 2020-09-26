// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

type property struct {
	name          string
	defaultValue  interface{}
	providerChain *[]Provider
	converter     func(value interface{}) interface{}
}

func (p *property) WithDefaultValue(defaultValue interface{}) {
	p.defaultValue = defaultValue
}
