// Copyright 2018 John Deng (hi.devops.io@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package factory provides InstantiateFactory and ConfigurableFactory interface
package factory

import (
	"hidevops.io/hiboot/pkg/system"
	"hidevops.io/hiboot/pkg/utils/reflector"
	"reflect"
)

const (
	// InstantiateFactoryName is the instance name of factory.instantiateFactory
	InstantiateFactoryName = "factory.instantiateFactory"
	// ConfigurableFactoryName is the instance name of factory.configurableFactory
	ConfigurableFactoryName = "factory.configurableFactory"
)

// Factory interface
type Factory interface{}

// InstantiateFactory instantiate factory interface
type InstantiateFactory interface {
	Initialized() bool
	SetInstance(params ...interface{}) (err error)
	GetInstance(params ...interface{}) (retVal interface{})
	GetInstances(name string) (retVal []interface{})
	Items() map[string]interface{}
	AppendComponent(c ...interface{})
	BuildComponents() (err error)
	CustomProperties() map[string]interface{}
}

// ConfigurableFactory configurable factory interface
type ConfigurableFactory interface {
	InstantiateFactory
	SystemConfiguration() *system.Configuration
	Configuration(name string) interface{}
	//Initialize(configurations cmap.ConcurrentMap) (err error)
	BuildSystemConfig() (systemConfig *system.Configuration, err error)
	Build(configs []*MetaData)
}

// Configuration configuration interface
type Configuration interface {
}

type depsMap map[string][]string

// Deps the dependency mapping of configuration
type Deps struct {
	deps depsMap
}

func (c *Deps) ensure() {
	if c.deps == nil {
		c.deps = make(depsMap)
	}
}

// Get get the dependencies mapping
func (c *Deps) Get(name string) (deps []string) {
	c.ensure()

	deps = c.deps[name]

	return
}

// Set set dependencies
func (c *Deps) Set(dep interface{}, value []string) {
	c.ensure()
	var name string
	val := reflect.ValueOf(dep)
	kind := val.Kind()
	switch kind {
	case reflect.Func:
		name = reflector.GetFuncName(dep)
	case reflect.String:
		name = dep.(string)
	default:
		return
	}
	c.deps[name] = value

	return
}
