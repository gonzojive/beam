package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/vet/bazel/testbazelpipeline"
)

var (
	outFile = flag.String("output", "", "Output .go file with code that makes the pipeline performant.")
)

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
	code, err := testbazelpipeline.GenerateCode(context.Background())
	if err != nil {
		return fmt.Errorf("error generating code: %v", err)
	}
	if err := ioutil.WriteFile(*outFile, []byte(code), 0664); err != nil {
		return err
	}
	return nil
}
