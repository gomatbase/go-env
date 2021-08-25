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
	sources      []*source
	chain        []Provider
	converter    func(value interface{}) interface{}
	listener     func(oldValue interface{}, newValue interface{})
	mutex        sync.Mutex
}

type source struct {
	source      Source
	cachedValue *valuePlaceholder
}

func Var(name string) *variable {
	return &variable{name: name, sources: make([]*source, 0)}
}

func (v *variable) Default(defaultValue interface{}) *variable {
	v.defaultValue = defaultValue
	return v
}

func (v *variable) From(s Source) *variable {
	source := &source{
		source: s,
	}
	v.sources = append(v.sources, source)
	return v
}

func (v *variable) Required() *variable {
	v.required = true
	return v
}

func (v *variable) ListeningWith(listener func(oldValue interface{}, newValue interface{})) *variable {
	v.listener = listener
	return v
}

func (v *variable) Add() error {
	return addVar(v)
}
