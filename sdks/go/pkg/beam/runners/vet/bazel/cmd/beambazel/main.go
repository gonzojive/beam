// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// beamvet is a command line tool for generating code as part of the bazel
// rules for defining Go pipelines.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

var (
	outFile      = flag.String("output", "", "Output .go file with code that makes the pipeline performant.")
	templateJSON = flag.String("template_json", "", "Template data as JSON.")
)

type templateData struct {
	PipelineImportPath string `json:"pipeline_import_path"`
	ConstructPipeline  string `json:"construct_pipeline"`
	GoPackage          string `json:"go_package"`
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatalf("error generating code: %v", err)
	}
}

func run() error {
	if *outFile == "" {
		return fmt.Errorf("must specify --output")
	}
	if *templateJSON == "" {
		return fmt.Errorf("must specify --template_json")
	}
	data := &templateData{}
	if err := json.Unmarshal([]byte(*templateJSON), data); err != nil {
		return fmt.Errorf("failed top parse template JSON: %w", err)
	}
	if data.ConstructPipeline == "" {
		return fmt.Errorf("construct_pipeline is empty - bad --template_json argument")
	}
	if data.PipelineImportPath == "" {
		return fmt.Errorf("pipeline_import_path is empty - bad --template_json argument")
	}
	if data.GoPackage == "" {
		return fmt.Errorf("go_package is empty - bad --template_json argument")
	}
	str := &strings.Builder{}
	if err := programTemplate.Execute(str, data); err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}
	return ioutil.WriteFile(*outFile, []byte(str.String()), 0664)
}

//func executeToString(t *template.)

var programTemplate = template.Must(template.New("main").Parse(`package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet"

	pipelinelib "{{.PipelineImportPath}}"
)

var (
	outFile = flag.String("output", "", "Output .go file with code that makes the pipeline performant.")
)

func main() {
	flag.Parse()
	if err := run(context.Background()); err != nil {
		log.Fatalf("error generating code: %v", err)
	}
}

func run(ctx context.Context) error {
	if *outFile == "" {
		return fmt.Errorf("must specify --output")
	}
	code, err := GenerateCode(ctx, {{.GoPackage | printf "%q"}})
	if err != nil {
		return fmt.Errorf("error generating code: %v", err)
	}
	if err := ioutil.WriteFile(*outFile, []byte(code), 0664); err != nil {
		return err
	}
	return nil
}

// GenerateCode returns generated Go code that should be compiled into the
// target package.
func GenerateCode(ctx context.Context, packageName string) (string, error) {
	p, err := pipelinelib.{{.ConstructPipeline}}(ctx)
	if err != nil {
		return "", fmt.Errorf("pipeline construction for code generator failed: %w", err)
	}
	eval, err := vet.Evaluate(ctx, p)
	if err != nil {
		return "", err
	}
	return eval.GenerateToString(packageName)
}
`))
