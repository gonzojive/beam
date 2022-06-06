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

package vet_test

import (
	"context"
	"testing"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet/bazel/testbazelpipeline"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet/testpipeline"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name                                                                       string
		c                                                                          func(beam.Scope)
		performant, allExported, usesDefaultReflectionShims, requiresRegistrations bool
	}{
		{name: "Performant", c: testpipeline.Performant, performant: true},
		{name: "FunctionReg", c: testpipeline.FunctionReg, allExported: true, usesDefaultReflectionShims: true, requiresRegistrations: true},
		{name: "ShimNeeded", c: testpipeline.ShimNeeded, usesDefaultReflectionShims: true},
		{name: "TypeReg", c: testpipeline.TypeReg, usesDefaultReflectionShims: true, requiresRegistrations: true},
		{name: "TypeReg", c: testbazelpipeline.ConstructPipeline, performant: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p, s := beam.NewPipelineWithRoot()
			test.c(s)
			e, err := vet.Evaluate(context.Background(), p)
			if err != nil {
				t.Fatalf("failed to evaluate testpipeline.Pipeline: %v", err)
			}
			if e.Performant() != test.performant {
				code, err := e.GenerateToString("somepackage")
				if err != nil {
					t.Errorf("unable to generate code for non-performant pipeline: %v", err)
				}
				t.Fatalf("e.Performant() = %v, want %v; generated code:\n%s", e.Performant(), test.performant, code)
			}
			// Abort early for performant pipelines.
			if test.performant {
				return
			}
			if e.AllExported() != test.allExported {
				t.Errorf("e.AllExported() = %v, want %v", e.AllExported(), test.allExported)
			}
			if e.RequiresRegistrations() != test.requiresRegistrations {
				t.Errorf("e.RequiresRegistrations() = %v, want %v", e.RequiresRegistrations(), test.requiresRegistrations)
			}
			if e.UsesDefaultReflectionShims() != test.usesDefaultReflectionShims {
				t.Errorf("e.UsesDefaultReflectionShims() = %v, want %v", e.UsesDefaultReflectionShims(), test.usesDefaultReflectionShims)
			}
		})
	}
}
