package factory

import (
	"github.com/stretchr/testify/assert"
	"hidevops.io/hiboot/pkg/system/types"
	"hidevops.io/hiboot/pkg/utils/reflector"
	"reflect"
	"testing"
)

type foo struct {
	name string
}

type fooBarService struct {
	HelloWorld HelloWorld `inject:""`
	foo        *foo
}

type Hello string
type HelloWorld string
type HelloHiboot string

type helloConfiguration struct {
	Configuration
}

func (c *helloConfiguration) Hello() Hello {
	return Hello("Hello World")
}

func (c *helloConfiguration) HelloWorld(h Hello) HelloWorld {
	return HelloWorld(h + "World")
}

func (c *helloConfiguration) HelloHiboot(h Hello) HelloHiboot {
	return HelloHiboot(h + "Hello Hiboot")
}

func newFooBarService(foo *foo) *fooBarService {
	return &fooBarService{foo: foo}
}

func TestUtils(t *testing.T) {
	helloConfig := new(helloConfiguration)
	helloTyp := reflect.TypeOf(helloConfig)
	numOfMethod := helloTyp.NumMethod()
	//log.Debug("methods: ", numOfMethod)
	methodTestData := make([]*MetaData, 0)
	for mi := 0; mi < numOfMethod; mi++ {
		method := helloTyp.Method(mi)
		// append inst to f.components
		methodTestData = append(methodTestData, NewMetaData(helloConfig, method))
	}

	t.Run("should parse instance name via object", func(t *testing.T) {
		md := NewMetaData("", new(fooBarService))
		assert.Equal(t, "fooBarService", md.TypeName)
		assert.Equal(t, "factory", md.PkgName)
		assert.NotEqual(t, nil, md.Object)
	})

	t.Run("should parse instance name via object", func(t *testing.T) {
		md := NewMetaData("", new(fooBarService))
		assert.Equal(t, "fooBarService", md.TypeName)
		assert.Equal(t, "factory", md.PkgName)
		assert.NotEqual(t, nil, md.Object)
	})

	t.Run("should parse instance name via object with eliminator", func(t *testing.T) {
		md := NewMetaData(new(fooBarService))
		assert.Equal(t, "fooBarService", md.TypeName)
		assert.Equal(t, "factory.fooBarService", md.Name)
		assert.NotEqual(t, nil, md.Object)
	})

	t.Run("should parse object instance name via constructor", func(t *testing.T) {
		md := NewMetaData("", newFooBarService)
		assert.Equal(t, "fooBarService", md.TypeName)
		assert.Equal(t, types.Func, md.Kind)
	})

	t.Run("should parse object pkg name", func(t *testing.T) {
		type service struct{}
		svc := new(service)
		md := NewMetaData(svc)
		assert.Equal(t, "factory.service", md.Name)
		assert.Equal(t, svc, md.Object)
	})

	t.Run("should parse object instance name", func(t *testing.T) {
		type service struct{}
		svc := new(service)
		md := NewMetaData("foo", svc)
		assert.Equal(t, "service", md.TypeName)
		assert.Equal(t, "factory.foo", md.Name)
		assert.Equal(t, svc, md.Object)
	})

	t.Run("should parse object pkg name", func(t *testing.T) {
		type service struct{}
		svc := new(service)
		md := NewMetaData(&MetaData{Object: new(service)})
		assert.Equal(t, "factory.service", md.Name)
		assert.Equal(t, svc, md.Object)
	})

	t.Run("should parse object pkg name", func(t *testing.T) {
		type service struct{}
		svc := new(service)
		name, object := ParseParams(svc)
		assert.Equal(t, "factory.service", name)
		assert.Equal(t, svc, object)
	})

	t.Run("should parse object pkg name", func(t *testing.T) {
		type service struct{}
		svc := new(service)
		name, object := ParseParams("factory.fooService", svc)
		assert.Equal(t, "factory.fooService", name)
		assert.Equal(t, svc, object)
	})

	t.Run("should parse object pkg name", func(t *testing.T) {
		name, _ := ParseParams("factory.fooService")
		assert.Equal(t, "factory.fooService", name)
	})

	t.Run("should parse func dependencies", func(t *testing.T) {
		fn := newFooBarService
		ft, ok := reflector.GetObjectType(fn)
		assert.Equal(t, true, ok)

		deps := parseDependencies(fn, types.Func, ft)
		assert.Equal(t, []string{"factory.foo"}, deps)
	})

	t.Run("should append dep", func(t *testing.T) {
		dep := appendDep("", "a.b")
		dep = appendDep(dep, "c.d")
		assert.Equal(t, "a.b,c.d", dep)
	})
}
