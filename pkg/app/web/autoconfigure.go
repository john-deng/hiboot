package web

import (
	"fmt"
	"github.com/kataras/iris"
	irsctx "github.com/kataras/iris/context"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/log"
)

const Profile = "web"

type configuration struct {
	at.AutoConfiguration

	Properties properties `mapstructure:"web"`
}

func newWebConfiguration() *configuration {
	return &configuration{}
}

func init() {
	app.Register(newWebConfiguration)
}

// Context is the instance of context.Context
func (c *configuration) Context(app *webApp) *Context {
	ctx := &Context{
		Context: irsctx.NewContext(app),
	}

	if c.Properties.View.Enabled {
		v := c.Properties.View
		app.RegisterView(iris.HTML(v.ResourcePath, v.Extension))

		route := app.Get(v.ContextPath, func(ctx iris.Context) {
			ctx.View(v.DefaultPage)
		})
		route.MainHandlerName = fmt.Sprintf("%s%s ", v.ContextPath, v.DefaultPage)
		log.Infof("Mapped \"%v\" onto %v", v.ContextPath, v.DefaultPage)
	}

	app.ContextPool.Attach(func() irsctx.Context {
		return &Context{
			// Optional Part 3:
			Context: irsctx.NewContext(app),
		}
	})

	return ctx
}
