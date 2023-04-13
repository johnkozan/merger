// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merger

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDstore(t *testing.T) {
	store := NewDStoreIO(
		testLogger,
		testTracer,
		dstore.NewMockStore(nil),
		dstore.NewMockStore(nil),
		dstore.NewMockStore(nil),
		1,
		0,
		100,
	)

	_, ok := store.(ForkAwareIOInterface)
	require.True(t, ok)

	// non-fork-aware
	store = NewDStoreIO(
		testLogger,
		testTracer,
		dstore.NewMockStore(nil),
		dstore.NewMockStore(nil),
		nil,
		1,
		0,
		100,
	)

	_, ok = store.(ForkAwareIOInterface)
	require.False(t, ok)
}

func newDStoreIO(
	oneBlocksStore dstore.Store,
	mergedBlocksStore dstore.Store,
) IOInterface {
	return NewDStoreIO(testLogger, testTracer, oneBlocksStore, mergedBlocksStore, nil, 0, 0, 100)
}

func TestMergerIO_MergeUploadPerfect(t *testing.T) {
	bstream.GetBlockWriterHeaderLen = 0

	files := []*bstream.OneBlockFile{
		block100,
		block101,
	}
	var mergeLastBase string
	var filesRead []string
	var mergeCounter int
	done := make(chan struct{})

	oneBlockStore := dstore.NewMockStore(nil)
	oneBlockStore.OpenObjectFunc = func(_ context.Context, name string) (io.ReadCloser, error) {
		filesRead = append(filesRead, name)
		if len(filesRead) == 2 {
			close(done)
		}
		return ioutil.NopCloser(strings.NewReader("")), nil
	}
	mergedBlocksStore := dstore.NewMockStore(
		func(base string, f io.Reader) (err error) {
			mergeLastBase = base
			mergeCounter++
			return nil
		},
	)

	mio := newDStoreIO(oneBlockStore, mergedBlocksStore)

	err := mio.MergeAndStore(context.Background(), 100, files)
	require.NoError(t, err)
	assert.Equal(t, mergeCounter, 1)
	assert.Equal(t, mergeLastBase, "0000000100")

	expectFilenames := []string{
		"0000000100-0000000000000100a-0000000000000099a-98-suffix", // read header
		"0000000100-0000000000000100a-0000000000000099a-98-suffix",
		"0000000101-0000000000000101a-0000000000000100a-99-suffix",
	}

	select {
	case <-time.After(time.Second):
		t.Error("timeout waiting for read")
	case <-done:
	}
	assert.Equal(t, expectFilenames, filesRead)
}

func TestMergerIO_MergeUploadFiltered(t *testing.T) {
	files := []*bstream.OneBlockFile{
		block98,
		block99,
		block100,
		block101,
	}

	var mergeLastBase string
	var filesRead []string
	var mergeCounter int
	done := make(chan struct{})

	oneBlockStore := dstore.NewMockStore(nil)
	oneBlockStore.OpenObjectFunc = func(_ context.Context, name string) (io.ReadCloser, error) {
		filesRead = append(filesRead, name)
		if len(filesRead) == 2 {
			close(done)
		}
		return ioutil.NopCloser(strings.NewReader("")), nil
	}
	mergedBlocksStore := dstore.NewMockStore(
		func(base string, f io.Reader) (err error) {
			mergeLastBase = base
			mergeCounter++
			return nil
		},
	)

	mio := newDStoreIO(oneBlockStore, mergedBlocksStore)

	err := mio.MergeAndStore(context.Background(), 100, files)
	require.NoError(t, err)
	assert.Equal(t, mergeCounter, 1)
	assert.Equal(t, mergeLastBase, "0000000100")

	expectFilenames := []string{
		"0000000098-0000000000000098a-0000000000000097a-96-suffix", // read header
		// 99 not read
		"0000000100-0000000000000100a-0000000000000099a-98-suffix",
		"0000000101-0000000000000101a-0000000000000100a-99-suffix",
	}

	select {
	case <-time.After(time.Second):
		t.Error("timeout waiting for read")
	case <-done:
	}
	assert.Equal(t, expectFilenames, filesRead)
}

func TestMergerIO_MergeUploadNoFiles(t *testing.T) {
	files := []*bstream.OneBlockFile{}

	oneBlockStore := dstore.NewMockStore(nil)
	mergedBlocksStore := dstore.NewMockStore(nil)
	mio := newDStoreIO(oneBlockStore, mergedBlocksStore)

	err := mio.MergeAndStore(context.Background(), 114, files)
	require.Error(t, err)
}
func TestMergerIO_MergeUploadFilteredToZero(t *testing.T) {
	files := []*bstream.OneBlockFile{
		block102Final100,
		block103Final101,
	}
	oneBlockStore := dstore.NewMockStore(nil)
	mergedBlocksStore := dstore.NewMockStore(nil)
	mio := newDStoreIO(oneBlockStore, mergedBlocksStore)

	block102Final100.MemoizeData = []byte{0x0, 0x1, 0x2, 0x3}
	block103Final101.MemoizeData = []byte{0x0, 0x1, 0x2, 0x3}

	err := mio.MergeAndStore(context.Background(), 114, files)
	require.NoError(t, err)
}
