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

package web

import (
	"os"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"crypto/rsa"
	"github.com/fatih/camelcase"
	"github.com/kataras/iris"
	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/httptest"
	"github.com/kataras/iris/context"
	"github.com/iris-contrib/httpexpect"
	"github.com/hidevopsio/hiboot/pkg/system"
	"github.com/hidevopsio/hiboot/pkg/utils"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/hiboot/pkg/starter/web/jwt"
	"path/filepath"
)

const (
	mainPackageDepth = 3

	pathSep = "/"

	AuthTypeDefault = ""
	AuthTypeAnon    = "anon"
	AuthTypeJwt     = "jwt"
)

type ApplicationInterface interface {
	Init()
	Config() *system.Configuration
	GetSignKey() *rsa.PrivateKey
	Run()
	NewTestServer(t *testing.T) *httpexpect.Expect
}

type Application struct {
	app    *iris.Application
	config *system.Configuration
	jwtEnabled bool
	workDir string
}

type Health struct {
	Status string `json:"status"`
}

type Controller struct {
	ContextMapping string
	AuthType       string
}

const (
	application = "application"
	config      = "config"
	yaml        = "yaml"
)

var (
	Controllers []interface{}
)


func (wa *Application) Init() {
	log.Println("application init")

	wa.workDir = utils.GetWorkingDir("")

	log.Println("working dir: ", wa.workDir)

	builder := &system.Builder{
		Path:       filepath.Join(wa.workDir, config),
		Name:       application,
		FileType:   yaml,
		Profile:    os.Getenv("APP_PROFILES_ACTIVE"),
		ConfigType: system.Configuration{},
	}
	cp, err := builder.Build()
	if err == nil {
		wa.config = cp.(*system.Configuration)
		log.SetLevel(wa.config.Logging.Level)
	} else {
		log.SetLevel(log.DebugLevel)
	}


	err = jwt.Init(wa.workDir)
	if err != nil {
		wa.jwtEnabled = false
		log.Error(err.Error())
	} else {
		wa.jwtEnabled = true
	}
}

func (wa *Application) Config() *system.Configuration {
	return wa.config
}

func (wa *Application) Run() {
	serverPort := ":8080"
	if wa.config != nil {
		serverPort = fmt.Sprintf(":%v", wa.config.Server.Port)
	}
	// TODO: WithCharset should be configurable
	wa.app.Run(iris.Addr(fmt.Sprintf(serverPort)), iris.WithCharset("UTF-8"), iris.WithoutVersionChecker)
}

func (wa *Application) NewTestServer(t *testing.T) *httpexpect.Expect {
	return httptest.New(t, wa.app)
}

func healthHandler(app *iris.Application) *router.Route {
	return app.Get("/health", func(ctx context.Context) {
		health := Health{
			Status: "UP",
		}
		ctx.JSON(health)
	})
}

func (wa *Application) handle(method reflect.Method, object interface{}, ctx context.Context) {
	//log.Debug("NumIn: ", method.Type.NumIn())
	inputs := make([]reflect.Value, method.Type.NumIn())

	inputs[0] = reflect.ValueOf(object)
	inputs[1] = reflect.ValueOf(ctx)
	method.Func.Call(inputs)
}

func Add(controller interface{})  {
	Controllers = append(Controllers, controller)
}

func NewApplication(controllers ... interface{}) (*Application, error) {
	wa := &Application{}

	wa.Init()

	app := iris.New()

	// The only one Required:
	// here is how you define how your own context will
	// be created and acquired from the iris' generic context pool.
	app.ContextPool.Attach(func() context.Context {
		return &Context{
			// Optional Part 3:
			Context: context.NewContext(app),
		}
	})

	wa.app = app

	healthHandler(app)

	if len(controllers) == 0 {
		controllers = Controllers
		if len(controllers) == 0 {
			return nil, &system.NotFoundError{Name: "controller"}
		}
	}

	if ! wa.jwtEnabled {
		err := wa.register(controllers, AuthTypeAnon, AuthTypeDefault, AuthTypeJwt)
		if err != nil {
			return nil, err
		}
	} else {
		err := wa.register(controllers, AuthTypeAnon, AuthTypeDefault)
		if err != nil {
			return nil, err
		}

		app.Use(jwt.GetHandler().Serve)

		err = wa.register(controllers, AuthTypeJwt)
		if err != nil {
			return nil, err
		}
	}

	return wa, nil
}

func (wa *Application)register(controllers []interface{}, auths... string) error {
	app := wa.app
	for _, c := range controllers {
		field := reflect.ValueOf(c)

		fieldType := field.Type()
		log.Debug("fieldType: ", fieldType)
		fieldName := fieldType.Elem().Name()
		log.Debug("fieldName: ", fieldName)

		controller := field.Interface()
		log.Debug("controller: ", controller)

		fieldAuth := field.Elem().FieldByName("AuthType")
		if ! fieldAuth.IsValid() {
			return &system.InvalidControllerError{Name: fieldName}
		}
		a := fmt.Sprintf("%v", fieldAuth.Interface())
		log.Debug(a)
		if ! utils.StringInSlice(a, auths) {
			continue
		}

		cp := field.Elem().FieldByName("ContextMapping")
		if ! cp.IsValid() {
			return &system.InvalidControllerError{Name: fieldName}
		}
		contextMapping := fmt.Sprintf("%v", cp.Interface())

		fieldNames := camelcase.Split(fieldName)
		controllerName := ""
		if len(fieldNames) >= 2 {
			controllerName = strings.Replace(fieldName, fieldNames[len(fieldNames)-1], "", 1)
			controllerName = utils.LowerFirst(controllerName)
		}
		log.Debug("controllerName: ", controllerName)
		if contextMapping == "" {
			contextMapping = pathSep + controllerName
		}

		numOfMethod := field.NumMethod()
		log.Debug("methods: ", numOfMethod)

		beforeMethod, ok := fieldType.MethodByName("Before")
		var party iris.Party
		if ok {
			log.Debug("contextPath: ", contextMapping)
			log.Debug("beforeMethod.Name: ", beforeMethod.Name)
			party = app.Party(contextMapping, func(ctx context.Context) {
				wa.handle(beforeMethod, controller, ctx)
			})
		}

		for mi := 0; mi < numOfMethod; mi++ {
			method := fieldType.Method(mi)
			methodName := method.Name
			log.Debug("method: ", methodName)

			ctxMap := camelcase.Split(methodName)
			httpMethod := strings.ToUpper(ctxMap[0])
			apiContextMapping := strings.Replace(methodName, ctxMap[0], "", 1)
			apiContextMapping = pathSep + utils.LowerFirst(apiContextMapping)

			if party == nil {
				relativePath := filepath.Join(contextMapping, apiContextMapping)
				log.Debug("relativePath: ", relativePath)
				app.Handle(httpMethod, relativePath, func(ctx context.Context) {
					wa.handle(method, controller, ctx)
				})
			} else {
				log.Debug("contextMapping: ", apiContextMapping)
				party.Handle(httpMethod, apiContextMapping, func(ctx context.Context) {
					wa.handle(method, controller, ctx)
				})
			}

		}

	}
	return nil
}
