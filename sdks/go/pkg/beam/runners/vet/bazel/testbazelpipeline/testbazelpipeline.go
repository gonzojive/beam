package testbazelpipeline

import (
	"context"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet"
)

func ConstructPipeline(s beam.Scope) {
	inputs := beam.Create(s, 1, 2, 3)
	same := beam.ParDo(s, identInt, inputs)
	beam.ParDo(s, plus1Int, same)
}

func identInt(i int) int {
	return i
}

func plus1Int(i int) int {
	return i + i
}

// GenerateCode returns generated Go code that should be compiled into the
// current package.
func GenerateCode(ctx context.Context) (string, error) {
	p := beam.NewPipeline()
	ConstructPipeline(p.Root())
	eval, err := vet.Evaluate(ctx, p)
	if err != nil {
		return "", err
	}
	return eval.GenerateToString("testbazelpipeline")
}
