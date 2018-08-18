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

package cli

import (
	"reflect"
	"os"
	"github.com/hidevopsio/hiboot/pkg/inject"
	"strings"
	"sync"
	"runtime"
	"path/filepath"
	"github.com/hidevopsio/hiboot/pkg/utils/sort"
	"github.com/hidevopsio/hiboot/pkg/utils/gotest"
)

type Application interface {
	Run()
	Init(cmd ...Command) error
	Root() Command
	SetRoot(root Command)
}

type application struct {
	root Command
}

type CommandNameValue struct {
	Name string
	Command interface{}
}

var (
	commandContainer  map[string][]Command
	commandNames []string
	app *application
	once sync.Once
)

func init() {
	commandContainer = make(map[string][]Command)
	commandNames = make([]string, 0)
}

func AddCommand(parentPath string, commands ...Command) {
	// de-duplication
	if commandContainer[parentPath] == nil {
		commandNames = append(commandNames, parentPath)
	}
	for _, command := range commands {
		commandContainer[parentPath] = append(commandContainer[parentPath], command)
	}
}

func GetApplication() Application {
	once.Do(func() {
		app = new(application)
		app.SetRoot(new(rootCommand))
	})
	return app
}

func NewApplication(cmd ...Command) Application {
	a := GetApplication()
	a.Init(cmd...)
	return a
}

func (a *application) injectCommand(cmd Command)  {
	fullname := "root"
	if cmd != nil {
		fullname = cmd.FullName()
	}
	for _, child := range cmd.Children() {
		inject.IntoObject(reflect.ValueOf(child))
		child.SetFullName(fullname + "." + child.GetName())
		a.injectCommand(child)
	}
}

func (a *application) Init(cmd ...Command) error  {

	basename := filepath.Base(os.Args[0])
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	var root Command
	root = new(rootCommand)
	numOfCmd := len(cmd)
	if cmd != nil && numOfCmd > 0 {
		if numOfCmd == 1 {
			root = cmd[0]
		} else {
			root.Add(cmd...)
		}
	}
	root.SetName("root")
	inject.IntoObject(reflect.ValueOf(root))
	Register(root)
	a.SetRoot(root)
	if !gotest.IsRunning() {
		a.Root().EmbeddedCommand().Use = basename
	}

	if a.root != nil && a.root.HasChild() {
		a.injectCommand(a.root)
	} else {
		// parse commands
		parentContainer := make(map[string]Command)
		fullname := "root"
		sort.SortByLen(commandNames)
		parentContainer[fullname] = a.root
		for _, cmdName := range commandNames {
			commands := commandContainer[cmdName]
			parent := parentContainer[cmdName]
			if parent == nil {
				parent = a.root
			}
			for _, command := range commands {
				inject.IntoObject(reflect.ValueOf(command))
				parent.Add(command)
				fullname := cmdName + "." + command.GetName()
				parentContainer[fullname] = command
				command.SetFullName(fullname)
			}
		}
	}
	return nil
}

func (a *application) SetRoot(root Command) {
	a.root = root
}

func (a *application) Root() Command {
	return a.root
}

func (a *application) Run() {

	//log.Debug(commandContainer)
	if a.root != nil{
		if err := a.root.Exec(); err != nil {
			os.Exit(1)
		}
	}
}



