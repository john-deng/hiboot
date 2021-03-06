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

// Package grpc provides the hiboot starter for injectable grpc client and server dependency
package grpc

import (
	"google.golang.org/grpc"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/factory"
	"hidevops.io/hiboot/pkg/utils/reflector"
	"reflect"
)

const (
	// Profile is the profile of grpc, it should be as same as the package name
	Profile = "grpc"
)

type configuration struct {
	app.Configuration
	Properties properties `mapstructure:"grpc"`

	instantiateFactory factory.InstantiateFactory
}

type grpcService struct {
	name string
	cb   interface{}
	svc  interface{}
}

var (
	grpcServers []*grpcService
	grpcClients []*grpcService
)

// RegisterServer register server from application
func RegisterServer(register interface{}, server interface{}) {
	svrName := reflector.GetLowerCamelFullName(server)
	svr := &grpcService{
		name: svrName,
		cb:   register,
		svc:  server,
	}
	app.Register(server)
	grpcServers = append(grpcServers, svr)
}

// Server alias to RegisterServer
var Server = RegisterServer

// RegisterClient register client from application
func RegisterClient(name string, clientConstructors ...interface{}) {
	for _, clientConstructor := range clientConstructors {
		svr := &grpcService{
			name: name,
			cb:   clientConstructor,
		}
		grpcClients = append(grpcClients, svr)

		// pre-allocate client in order to pass dependency check
		typ, ok := reflector.GetObjectType(clientConstructor)
		if ok {
			// NOTE: it's very important !!!
			// To register grpc client and grpc.ClientConn in advance.
			// client should depends on grpc.clientFactory
			metaData := &factory.MetaData{
				Object:  reflect.New(typ).Interface(),
				Depends: []string{"grpc.clientFactory"},
			}
			app.Register(metaData)
		}
	}
	// Just register grpc.ClientConn in order to pass the dependency check
	app.Register(new(grpc.ClientConn))
}

// Client register client from application, it is a alias to RegisterClient
var Client = RegisterClient

func init() {
	app.Register(newConfiguration)
}

func newConfiguration(instantiateFactory factory.InstantiateFactory) *configuration {
	c := &configuration{
		instantiateFactory: instantiateFactory,
	}

	// we need to specify dependencies for runtime dependency injection
	var dep []string
	for _, srv := range grpcServers {
		if srv.svc != nil {
			dep = append(dep, srv.name)
		}
	}
	c.RuntimeDeps.Set(c.ServerFactory, dep)

	return c
}

// ClientConnector is the interface that connect to grpc client
// it can be injected to struct at runtime
func (c *configuration) ClientConnector() ClientConnector {
	return newClientConnector(c.instantiateFactory)
}

// GrpcClientFactory create gRPC Clients that registered by application
func (c *configuration) ClientFactory(cc ClientConnector) ClientFactory {
	return newClientFactory(c.instantiateFactory, c.Properties, cc)
}

// GrpcServer create new gRpc Server
func (c *configuration) Server() (grpcServer *grpc.Server) {
	// just return if grpc server is not enabled
	if c.Properties.Server.Enabled {
		grpcServer = grpc.NewServer()
	}
	return
}

// GrpcServerFactory create gRPC servers that registered by application
// go:depends
func (c *configuration) ServerFactory(grpcServer *grpc.Server) ServerFactory {
	return newServerFactory(c.instantiateFactory, c.Properties, grpcServer)
}
