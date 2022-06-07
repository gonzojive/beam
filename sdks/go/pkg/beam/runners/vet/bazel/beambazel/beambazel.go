package beambazel

import (
	"context"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
)

// RegisteredPipeline is an object to hold a registered pipeline.
type RegisteredPipeline struct {
	importPath string
	gen        func(context.Context) (*beam.Pipeline, error)
}

// ImportPath returns an import path.
func (rp *RegisteredPipeline) ImportPath() string { return rp.importPath }

// Pipeline generates a pipeline object.
func (rp *RegisteredPipeline) Pipeline(ctx context.Context) (*beam.Pipeline, error) {
	return rp.gen(ctx)
}

var registry []*RegisteredPipeline

// RegisterPipeline registers a function that will generate a beam pipeline.
func RegisterPipeline(importPath string, generator func(context.Context) (*beam.Pipeline, error)) {
	registry = append(registry, &RegisteredPipeline{
		importPath,
		generator,
	})
}

// RegisteredPipelines returns the set of registered generators.
func RegisteredPipelines() []*RegisteredPipeline {
	var out []*RegisteredPipeline
	out = append(out, registry...)
	return out
}
