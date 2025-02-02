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

// Package memfs contains a in-memory Beam filesystem. Useful for testing.
package memfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/filesystem"
)

func init() {
	filesystem.Register("memfs", New)
}

var instance = &fs{m: make(map[string][]byte)}

type fs struct {
	m  map[string][]byte
	mu sync.Mutex
}

// New returns the global memory filesystem.
func New(_ context.Context) filesystem.Interface {
	return instance
}

func (f *fs) Close() error {
	return nil
}

func (f *fs) List(_ context.Context, glob string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// As with other functions, the memfs:// prefix is optional.
	globNoScheme := strings.TrimPrefix(glob, "memfs://")

	var ret []string
	for k := range f.m {
		matched, err := filepath.Match(globNoScheme, strings.TrimPrefix(k, "memfs://"))
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %w", err)
		}
		if matched {
			ret = append(ret, k)
		}
	}
	sort.Strings(ret)
	return ret, nil
}

func (f *fs) OpenRead(_ context.Context, filename string) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	normalizedKey := normalize(filename)
	if _, ok := f.m[normalizedKey]; !ok {
		return nil, os.ErrNotExist
	}
	return &bytesReader{instance: f, normalizedKey: normalizedKey}, nil
}

func (f *fs) OpenWrite(_ context.Context, filename string) (io.WriteCloser, error) {
	key := normalize(filename)
	writer := &commitWriter{key: key, instance: f}
	_, err := writer.Write([]byte{})
	return writer, err
	// Create the file if it does not exist.

}

func (f *fs) Size(_ context.Context, filename string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if v, ok := f.m[normalize(filename)]; ok {
		return int64(len(v)), nil
	}
	return -1, os.ErrNotExist
}

// Remove the named file from the filesystem.
func (f *fs) Remove(_ context.Context, filename string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, filename)
	return nil
}

// Rename the old path to the new path.
func (f *fs) Rename(_ context.Context, oldpath, newpath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[newpath] = f.m[oldpath]
	delete(f.m, oldpath)
	return nil
}

// Copier copies the old path to the new path.
func (f *fs) Copy(_ context.Context, oldpath, newpath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[newpath] = f.m[oldpath]
	return nil
}

// Compile time check for interface implementations.
var (
	_ filesystem.Remover = ((*fs)(nil))
	_ filesystem.Renamer = ((*fs)(nil))
	_ filesystem.Copier  = ((*fs)(nil))
)

// Copier copies the old path to the new path.
func (f *fs) write(key string, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	cp := make([]byte, len(value))
	copy(cp, value)

	f.m[normalize(key)] = cp
	return nil
}

// Write stores the given key and value in the global store.
func Write(key string, value []byte) {
	instance.write(key, value)
}

func normalize(key string) string {
	if strings.HasPrefix(key, "memfs://") {
		return key
	}
	return "memfs://" + key
}

type commitWriter struct {
	key      string
	buf      bytes.Buffer
	instance *fs
}

func (w *commitWriter) Write(p []byte) (n int, err error) {
	n, err = w.buf.Write(p)
	if err != nil {
		return n, err
	}

	w.instance.write(w.key, w.buf.Bytes())
	return n, nil
}

func (w *commitWriter) Close() error {
	return nil
}

// bytesReader implements io.Reader, io.Seeker, io.Cloer for memfs "files."
type bytesReader struct {
	instance      *fs
	normalizedKey string
	pos           int64
}

var _ io.ReadSeekCloser = (*bytesReader)(nil)

func (r *bytesReader) Read(p []byte) (int, error) {
	r.instance.mu.Lock()
	defer r.instance.mu.Unlock()

	currentValue, exists := r.instance.m[r.normalizedKey]
	if !exists {
		return 0, os.ErrNotExist
	}
	if int(r.pos) >= len(currentValue) {
		return 0, io.EOF
	}
	wantEnd := int(r.pos) + len(p)
	if wantEnd > len(currentValue) {
		wantEnd = len(currentValue)
	}
	n := wantEnd - int(r.pos)
	copy(p, currentValue[r.pos:])
	r.pos = int64(wantEnd)
	return n, nil
}

func (r *bytesReader) Close() error { return nil }
func (r *bytesReader) Seek(offset int64, whence int) (int64, error) {
	r.instance.mu.Lock()
	defer r.instance.mu.Unlock()

	currentValue, exists := r.instance.m[r.normalizedKey]
	if !exists {
		return 0, os.ErrNotExist
	}
	currentLen := len(currentValue)

	wantPos := r.pos
	switch whence {
	case io.SeekCurrent:
		wantPos = r.pos + offset
	case io.SeekStart:
		wantPos = offset
	case io.SeekEnd:
		wantPos = int64(currentLen) + offset
	}
	if wantPos < 0 {
		return 0, fmt.Errorf("%w: invalid seek position %d is before start of file", errBadSeek, wantPos)
	}
	if int(wantPos) > currentLen {
		return 0, fmt.Errorf("%w: invalid seek position %d is after end of file", errBadSeek, wantPos)
	}
	r.pos = wantPos
	return r.pos, nil
}

var errBadSeek = errors.New("bad seek")
