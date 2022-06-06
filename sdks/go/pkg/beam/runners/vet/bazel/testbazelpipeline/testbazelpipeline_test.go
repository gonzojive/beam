package testbazelpipeline

import (
	"context"
	"testing"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet"
)

func TestVet(t *testing.T) {
	p := beam.NewPipeline()
	ConstructPipeline(p.Root())
	eval, err := vet.Evaluate(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	eval.Performant()
}
