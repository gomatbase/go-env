// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import "sync"

type valuePlaceholder struct {
	value interface{}
}

type variable struct {
	name         string
	required     bool
	defaultValue interface{}
	cachedValue  *valuePlaceholder
	sources      []Source
	chain        []Provider
	converter    func(value interface{}) interface{}
	mutex        sync.Mutex
}

func Var(name string) *variable {
	return &variable{name: name, sources: make([]Source, 0)}
}

func (v *variable) Default(defaultValue interface{}) *variable {
	v.defaultValue = defaultValue
	return v
}

func (v *variable) From(source Source) *variable {
	v.sources = append(v.sources, source)
	return v
}

func (v *variable) Required() *variable {
	v.required = true
	return v
}

func (v *variable) Add() error {
	return addVar(v)
}
