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

package impl

import (
	"testing"
	"github.com/hidevopsio/hi/boot/pkg/log"
	"github.com/hidevopsio/hi/cicd/pkg/ci"
	"os"
)

func init()  {
	log.SetLevel(log.DebugLevel)
}

func TestJavaWarPipeline(t *testing.T)  {

	log.Debug("Test Java War Pipeline")

	javaWarPipeline := &JavaWarPipeline{
		JavaPipeline{
			ci.Pipeline{
				App: "test",
				Project: "demo",
			},
		},
	}

	username := os.Getenv("SCM_USERNAME")
	password := os.Getenv("SCM_PASSWORD")
	javaWarPipeline.Init(&ci.Pipeline{Name: "java-war"})
	javaWarPipeline.Run(username, password, false)
}