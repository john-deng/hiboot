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

package logging

import (
	"github.com/stretchr/testify/assert"
	"hidevops.io/hiboot/pkg/app/fake"
	"hidevops.io/hiboot/pkg/utils/io"
	"reflect"
	"testing"
)

func TestConfiguration(t *testing.T) {
	c := newConfiguration(new(fake.ApplicationContext))
	c.Properties = &properties{
		Level: "debug",
	}

	t.Run("should get nil handler", func(t *testing.T) {
		lh := c.LoggerHandler()
		assert.IsType(t, reflect.Func, reflect.TypeOf(lh).Kind())
	})

	t.Run("should get handler", func(t *testing.T) {
		io.EnsureWorkDir(1, "config/application.yml")
		c.LoggerHandler()
	})
}
