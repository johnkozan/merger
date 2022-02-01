package bundle

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//                                  |                           |                                  |                           |
// 100a - 101a - 102a - 103a - 104a - 106a - 107a - 108a - 109a - 110a - 111a - 112a - 113a - 114a - 115a - 116a - 117a - 118a - 120a
//            \- 102b - 103b                     \- 108b - 109b - 110b
//                                                             \- 110c - 111c

func TestBundler_String(t *testing.T) {
	c := struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		blockLimit                 uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		name:                      "file 0",
		files:                     []string{},
		lastMergeBlockID:          "00000099a",
		blockLimit:                105,
		expectedCompleted:         true,
		expectedHighestBlockLimit: 104,
	}

	bundler := NewBundler(5, c.blockLimit)
	bundler.lastMergeOneBlockFile = &OneBlockFile{
		ID: c.lastMergeBlockID,
	}

	str := bundler.String()
	require.Contains(t, str, "bundle_size")
	require.Contains(t, str, "last_merge_block_num")
	require.Contains(t, str, "inclusive_lower_block_num")
	require.Contains(t, str, "exclusive_highest_block_limit")

}

func TestBundler_LastMergeOneBlockFile(t *testing.T) {
	c := struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		blockLimit                 uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		name:                      "file 0",
		files:                     []string{},
		lastMergeBlockID:          "00000099a",
		blockLimit:                105,
		expectedCompleted:         true,
		expectedHighestBlockLimit: 104,
	}

	bundler := NewBundler(5, c.blockLimit)
	bundler.lastMergeOneBlockFile = &OneBlockFile{
		ID: c.lastMergeBlockID,
	}
	last := bundler.LastMergeOneBlockFile()
	require.IsType(t, OneBlockFile{}, *last)
	require.Equal(t, last.ID, c.lastMergeBlockID)

}

func TestBundler_ExclusiveHighestBlockLimit(t *testing.T) {
	c := struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		blockLimit                 uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		name:                      "file 0",
		files:                     []string{},
		lastMergeBlockID:          "00000099a",
		blockLimit:                105,
		expectedCompleted:         true,
		expectedHighestBlockLimit: 104,
	}

	bundler := NewBundler(5, c.blockLimit)
	bundler.lastMergeOneBlockFile = &OneBlockFile{
		ID: c.lastMergeBlockID,
	}
	limit := bundler.ExclusiveHighestBlockLimit()
	require.Equal(t, "uint64", reflect.TypeOf(limit).Name())
	require.Equal(t, c.blockLimit, limit)
}

func TestBundler_AddOneBlockFileUglyPatch(t *testing.T) {
	c := struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		blockLimit                 uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		name: "file 0",
		files: []string{
			"0000000100-20210728T105016.0-00000100a-00000099a-90-suffix",
		},
		lastMergeBlockID:          "00000099a",
		blockLimit:                105,
		expectedCompleted:         true,
		expectedHighestBlockLimit: 104,
	}

	bundler := NewBundler(2, c.blockLimit)
	bundler.lastMergeOneBlockFile = &OneBlockFile{
		ID: c.lastMergeBlockID,
	}
	for _, f := range c.files {
		exists := bundler.AddOneBlockFile(MustNewOneBlockFile(f))
		require.False(t, exists)
		exists = bundler.AddOneBlockFile(MustNewOneBlockFile(f))
		require.True(t, exists)
	}
}

func TestBundler_AddPreMergedOneBlockFiles(t *testing.T) {
	c := struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		blockLimit                 uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		name: "file 0",
		files: []string{
			"0000000100-20210728T105016.0-00000100a-00000099a-90-suffix",
		},
		lastMergeBlockID:          "00000099a",
		blockLimit:                105,
		expectedCompleted:         true,
		expectedHighestBlockLimit: 104,
	}

	bundler := NewBundler(2, c.blockLimit)
	bundler.lastMergeOneBlockFile = &OneBlockFile{
		ID: c.lastMergeBlockID,
	}

	bundler.BundlePreMergedOneBlockFiles([]*OneBlockFile{})
	last := bundler.LastMergeOneBlockFile()
	limit := bundler.ExclusiveHighestBlockLimit()
	require.IsType(t, OneBlockFile{}, *last)
	require.Equal(t, last.CanonicalName, "")
	require.Equal(t, limit, c.blockLimit)

	for _, f := range c.files {
		bundler.BundlePreMergedOneBlockFiles([]*OneBlockFile{MustNewOneBlockFile(f)})
	}

	last = bundler.LastMergeOneBlockFile()
	limit = bundler.ExclusiveHighestBlockLimit()
	require.IsType(t, OneBlockFile{}, *last)
	require.Equal(t, last.CanonicalName, strings.Split(c.files[0], "-suffix")[0])
	require.Equal(t, limit, bundler.bundleSize+c.blockLimit)
}

func TestBundler_IsComplete(t *testing.T) {

	cases := []struct {
		name                       string
		files                      []string
		lastMergeBlockID           string
		exclusiveHighestBlockLimit uint64
		expectedCompleted          bool
		expectedLowerBlockNumLimit uint64
		expectedHighestBlockLimit  uint64
	}{
		{
			name: "file 0",
			files: []string{
				"0000000100-20210728T105016.0-00000100a-00000099a-90-suffix",
				"0000000101-20210728T105016.0-00000101a-00000100a-100-suffix",
				"0000000102-20210728T105016.0-00000102a-00000101a-100-suffix",
				"0000000102-20210728T105016.0-00000102b-00000101a-100-suffix",
				"0000000103-20210728T105016.0-00000103b-00000102b-100-suffix",
				"0000000103-20210728T105016.0-00000103a-00000102a-100-suffix",
				"0000000104-20210728T105016.0-00000104a-00000103a-100-suffix",
				"0000000106-20210728T105016.0-00000106a-00000104a-100-suffix",
			},
			lastMergeBlockID:           "00000099a",
			exclusiveHighestBlockLimit: 105,
			expectedCompleted:          true,
			expectedHighestBlockLimit:  104,
		},
		{
			name: "file 0 incomplete",
			files: []string{
				"0000000100-20210728T105016.0-00000100a-00000099a-90-suffix",
				"0000000101-20210728T105016.0-00000101a-00000100a-100-suffix",
				"0000000102-20210728T105016.0-00000102a-00000101a-100-suffix",
				"0000000102-20210728T105016.0-00000102b-00000101a-100-suffix",
				"0000000103-20210728T105016.0-00000103b-00000102b-100-suffix",
				"0000000103-20210728T105016.0-00000103a-00000102a-100-suffix",
				"0000000104-20210728T105016.0-00000104a-00000103a-100-suffix",
			},
			lastMergeBlockID:           "00000099a",
			exclusiveHighestBlockLimit: 105,
			expectedCompleted:          false,
			expectedHighestBlockLimit:  0,
		},
		{
			name: "file 0 no longest chain",
			files: []string{
				"0000000100-20210728T105016.0-00000100a-00000099a-90-suffix",
				"0000000101-20210728T105016.0-00000101a-00000100a-100-suffix",
				"0000000102-20210728T105016.0-00000102a-00000101a-100-suffix",
				"0000000102-20210728T105016.0-00000102b-00000101a-100-suffix",
				"0000000103-20210728T105016.0-00000103b-00000102b-100-suffix",
				"0000000103-20210728T105016.0-00000103a-00000102a-100-suffix",
			},
			lastMergeBlockID:           "00000099a",
			exclusiveHighestBlockLimit: 105,
			expectedCompleted:          false,
			expectedHighestBlockLimit:  0,
		},
		{
			name: "file 5",
			files: []string{
				"0000000106-20210728T105016.0-00000106a-00000104a-90-suffix",
				"0000000107-20210728T105016.0-00000107a-00000106a-106-suffix",
				"0000000108-20210728T105016.0-00000108b-00000107a-106-suffix",
				"0000000109-20210728T105016.0-00000109b-00000108b-106-suffix",
				"0000000110-20210728T105016.0-00000110b-00000109b-106-suffix",
				"0000000110-20210728T105016.0-00000110c-00000109b-106-suffix",
				"0000000111-20210728T105016.0-00000111c-00000110c-106-suffix",
				"0000000108-20210728T105016.0-00000108a-00000107a-106-suffix",
				"0000000109-20210728T105016.0-00000109a-00000108a-106-suffix",
				"0000000110-20210728T105016.0-00000110a-00000109a-106-suffix",
				"0000000111-20210728T105016.0-00000111a-00000110a-106-suffix",
				"0000000112-20210728T105016.0-00000112a-00000111a-106-suffix",
			},
			lastMergeBlockID:           "00000104a",
			exclusiveHighestBlockLimit: 110,
			expectedCompleted:          true,
			expectedHighestBlockLimit:  109,
		},
		{
			name: "file 10",
			files: []string{
				"0000000107-20210728T105016.1-00000107a-00000106a-90-suffix",
				"0000000108-20210728T105016.2-00000108b-00000107a-107-suffix",
				"0000000109-20210728T105016.3-00000109b-00000108b-107-suffix",
				"0000000110-20210728T105016.4-00000110b-00000109b-107-suffix",
				"0000000110-20210728T105016.5-00000110c-00000109b-107-suffix",
				"0000000111-20210728T105016.6-00000111c-00000110c-107-suffix",
				"0000000108-20210728T105016.7-00000108a-00000107a-107-suffix",
				"0000000109-20210728T105016.8-00000109a-00000108a-107-suffix",
				"0000000110-20210728T105016.9-00000110a-00000109a-107-suffix",
				"0000000111-20210728T105016.10-00000111a-00000110a-107-suffix",
				"0000000112-20210728T105016.11-00000112a-00000111a-107-suffix",
				"0000000113-20210728T105016.12-00000113a-00000112a-107-suffix",
				"0000000114-20210728T105016.13-00000114a-00000113a-107-suffix",
				"0000000115-20210728T105016.14-00000115a-00000114a-107-suffix",
			},
			lastMergeBlockID:           "00000109a",
			exclusiveHighestBlockLimit: 115,
			expectedCompleted:          true,
			expectedHighestBlockLimit:  114,
		},
		{
			name: "file 15",
			files: []string{
				"0000000115-20210728T105016.0-00000115a-00000114a-90-suffix",
				"0000000116-20210728T105016.0-00000116a-00000115a-115-suffix",
				"0000000117-20210728T105016.0-00000117a-00000116a-115-suffix",
				"0000000118-20210728T105016.0-00000118a-00000117a-115-suffix",
				"0000000120-20210728T105016.0-00000120a-00000118a-115-suffix",
			},
			lastMergeBlockID:           "00000114a",
			exclusiveHighestBlockLimit: 120,
			expectedCompleted:          true,
			expectedHighestBlockLimit:  118,
		},
		{
			name: "multiple root",
			files: []string{
				"0000000115-20210728T105016.0-00000115a-00000114a-90-suffix",
				"0000000116-20210728T105016.0-00000116a-00000115a-115-suffix",
				"0000000117-20210728T105016.0-00000117a-00000116a-115-suffix",
				"0000000118-20210728T105016.0-00000118a-00000117a-115-suffix",
				"0000000120-20210728T105016.0-00000120a-00000118a-115-suffix",

				"0000000300-20210728T105016.0-00000300a-00000299a-150-suffix",
				"0000000301-20210728T105016.0-00000301a-00000300a-150-suffix",
				"0000000302-20210728T105016.0-00000302a-00000301a-150-suffix",
				"0000000303-20210728T105016.0-00000303a-00000302a-150-suffix",
				"0000000304-20210728T105016.0-00000304a-00000303a-150-suffix",
				"0000000305-20210728T105016.0-00000305a-00000304a-150-suffix",
			},
			lastMergeBlockID:           "00000114a",
			exclusiveHighestBlockLimit: 120,
			expectedCompleted:          true,
			expectedHighestBlockLimit:  118,
		},
		{
			name: "No lib",
			files: []string{
				"0000000115-20210728T105016.0-00000115a-00000114a-90-suffix",
				"0000000116-20210728T105016.0-00000116a-00000115a-90-suffix",
				"0000000117-20210728T105016.0-00000117a-00000116a-90-suffix",
				"0000000118-20210728T105016.0-00000118a-00000117a-90-suffix",
				//"0000000120-20210728T105016.0-00000120a-00000118a-90-suffix",

				"0000000300-20210728T105016.0-00000300a-00000299a-150-suffix",
				"0000000301-20210728T105016.0-00000301a-00000300a-150-suffix",
				"0000000302-20210728T105016.0-00000302a-00000301a-150-suffix",
				"0000000303-20210728T105016.0-00000303a-00000302a-150-suffix",
				"0000000304-20210728T105016.0-00000304a-00000303a-150-suffix",
				"0000000305-20210728T105016.0-00000305a-00000304a-150-suffix",
			},
			lastMergeBlockID:           "00000114a",
			exclusiveHighestBlockLimit: 120,
			expectedCompleted:          false,
			expectedHighestBlockLimit:  0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bundler := NewBundler(5, c.exclusiveHighestBlockLimit)
			bundler.lastMergeOneBlockFile = &OneBlockFile{
				ID: c.lastMergeBlockID,
			}
			for _, f := range c.files {
				bundler.AddOneBlockFile(MustNewOneBlockFile(f))
			}
			completed, highestBlockLimit := bundler.BundleCompleted()
			require.Equal(t, c.expectedCompleted, completed)
			require.Equal(t, int(c.expectedHighestBlockLimit), int(highestBlockLimit))
		})
	}
}

func MustTestNewMergedOneBlockFile(fileName string) *OneBlockFile {
	fi := MustNewOneBlockFile(fileName)
	fi.Merged = true
	return fi
}

func TestBundler_MergeableFiles(t *testing.T) {
	cases := []struct {
		name                     string
		files                    []*OneBlockFile
		lastMergeBlockID         string
		blockLimit               uint64
		expectedIDs              []string
		expectedLastMergeBlockID string
	}{
		{
			name: "file 0",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.04-00000102b-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.05-00000103b-00000102b-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
				MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),
			},
			lastMergeBlockID:         "00000099a",
			blockLimit:               105,
			expectedIDs:              []string{"00000100a", "00000101a", "00000102a", "00000102b", "00000103b", "00000103a", "00000104a"},
			expectedLastMergeBlockID: "00000104a",
		},
		{
			name: "file 5",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000106-20210728T105016.01-00000106a-00000104a-90-suffix"),
				MustNewOneBlockFile("0000000107-20210728T105016.02-00000107a-00000106a-106-suffix"),
				MustNewOneBlockFile("0000000108-20210728T105016.03-00000108b-00000107a-106-suffix"),
				MustNewOneBlockFile("0000000109-20210728T105016.04-00000109b-00000108b-106-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.05-00000110b-00000109b-106-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.06-00000110c-00000109b-106-suffix"),
				MustNewOneBlockFile("0000000111-20210728T105016.07-00000111c-00000110c-106-suffix"),
				MustNewOneBlockFile("0000000108-20210728T105016.08-00000108a-00000107a-106-suffix"),
				MustNewOneBlockFile("0000000109-20210728T105016.09-00000109a-00000108a-106-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.10-00000110a-00000109a-106-suffix"),
				MustNewOneBlockFile("0000000111-20210728T105016.11-00000111a-00000110a-106-suffix"),
				MustNewOneBlockFile("0000000112-20210728T105016.12-00000112a-00000111a-106-suffix"),
			},
			lastMergeBlockID:         "00000104a",
			blockLimit:               110,
			expectedIDs:              []string{"00000106a", "00000107a", "00000108b", "00000109b", "00000108a", "00000109a"},
			expectedLastMergeBlockID: "00000109a",
		},
		{
			name: "file 10",
			files: []*OneBlockFile{
				MustTestNewMergedOneBlockFile("0000000107-20210728T105016.01-00000107a-00000106a-90-suffix"),
				MustTestNewMergedOneBlockFile("0000000108-20210728T105016.02-00000108b-00000107a-107-suffix"),
				MustTestNewMergedOneBlockFile("0000000109-20210728T105016.03-00000109b-00000108b-107-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.04-00000110b-00000109b-107-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.05-00000110c-00000109b-107-suffix"),
				MustNewOneBlockFile("0000000111-20210728T105016.06-00000111c-00000110c-107-suffix"),
				MustTestNewMergedOneBlockFile("0000000108-20210728T105016.07-00000108a-00000107a-107-suffix"),
				MustTestNewMergedOneBlockFile("0000000109-20210728T105016.08-00000109a-00000108a-107-suffix"),
				MustNewOneBlockFile("0000000110-20210728T105016.09-00000110a-00000109a-107-suffix"),
				MustNewOneBlockFile("0000000111-20210728T105016.10-00000111a-00000110a-107-suffix"),
				MustNewOneBlockFile("0000000112-20210728T105016.11-00000112a-00000111a-107-suffix"),
				MustNewOneBlockFile("0000000113-20210728T105016.12-00000113a-00000112a-107-suffix"),
				MustNewOneBlockFile("0000000114-20210728T105016.13-00000114a-00000113a-107-suffix"),
				MustNewOneBlockFile("0000000115-20210728T105016.14-00000115a-00000114a-107-suffix"),
			},
			lastMergeBlockID:         "00000109a",
			blockLimit:               115,
			expectedIDs:              []string{"00000110b", "00000110c", "00000111c", "00000110a", "00000111a", "00000112a", "00000113a", "00000114a"},
			expectedLastMergeBlockID: "00000114a",
		},
		{
			name: "file 15",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000115-20210728T105016.0-00000115a-00000114a-90-suffix"),
				MustNewOneBlockFile("0000000116-20210728T105016.0-00000116a-00000115a-115-suffix"),
				MustNewOneBlockFile("0000000117-20210728T105016.0-00000117a-00000116a-115-suffix"),
				MustNewOneBlockFile("0000000118-20210728T105016.0-00000118a-00000117a-115-suffix"),
				MustNewOneBlockFile("0000000120-20210728T105016.0-00000120a-00000118a-115-suffix"),
			},
			lastMergeBlockID:         "00000114a",
			blockLimit:               120,
			expectedIDs:              []string{"00000115a", "00000116a", "00000117a", "00000118a"},
			expectedLastMergeBlockID: "00000118a",
		},
		{
			name: "file with holes",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.0-00000100a-00000099a-99-suffix"),

				MustNewOneBlockFile("0000000115-20210728T105016.0-00000115a-00000114a-100-suffix"),
				MustNewOneBlockFile("0000000116-20210728T105016.0-00000116a-00000115a-115-suffix"),

				MustNewOneBlockFile("0000000117-20210728T105016.0-00000117b-00000116b-115-suffix"),
				MustNewOneBlockFile("0000000118-20210728T105016.0-00000118b-00000117b-115-suffix"),

				MustNewOneBlockFile("0000000117-20210728T105016.1-00000117a-00000116a-115-suffix"),
				MustNewOneBlockFile("0000000118-20210728T105016.1-00000118a-00000117a-115-suffix"),
				MustNewOneBlockFile("0000000120-20210728T105016.0-00000120a-00000118a-115-suffix"),
			},
			lastMergeBlockID:         "00000114a",
			blockLimit:               120,
			expectedIDs:              []string{"00000100a", "00000115a", "00000116a", "00000117b", "00000118b", "00000117a", "00000118a"},
			expectedLastMergeBlockID: "00000118a",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bundler := NewBundler(5, c.blockLimit)
			bundler.lastMergeOneBlockFile = &OneBlockFile{ID: c.lastMergeBlockID}
			for _, f := range c.files {
				bundler.AddOneBlockFile(f)
			}
			completed, highestBlockLimit := bundler.BundleCompleted()
			require.True(t, completed)
			mergeableFiles := bundler.ToBundle(highestBlockLimit)
			bundler.Commit(highestBlockLimit)

			ids := ToIDs(mergeableFiles)
			require.Equal(t, c.expectedIDs, ids)
			require.Equal(t, c.expectedLastMergeBlockID, bundler.lastMergeOneBlockFile.ID)
		})
	}
}

func TestBundler_Complicated(t *testing.T) {

	files := []*OneBlockFile{
		MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
		MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
		MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
		MustNewOneBlockFile("0000000102-20210728T105016.04-00000102b-00000101a-100-suffix"),
		MustNewOneBlockFile("0000000103-20210728T105016.05-00000103b-00000102b-100-suffix"),
		MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
		MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
		MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),
		MustNewOneBlockFile("0000000107-20210728T105016.09-00000107a-00000106a-100-suffix"),
		MustNewOneBlockFile("0000000108-20210728T105016.10-00000108b-00000107a-100-suffix"),
		MustNewOneBlockFile("0000000109-20210728T105016.11-00000109b-00000108b-100-suffix"),
		MustNewOneBlockFile("0000000110-20210728T105016.12-00000110b-00000109b-100-suffix"),
		MustNewOneBlockFile("0000000110-20210728T105016.13-00000110c-00000109b-100-suffix"),
		MustNewOneBlockFile("0000000111-20210728T105016.14-00000111c-00000110c-100-suffix"),
		MustNewOneBlockFile("0000000108-20210728T105016.15-00000108a-00000107a-100-suffix"),
		MustNewOneBlockFile("0000000109-20210728T105016.16-00000109a-00000108a-100-suffix"),
		MustNewOneBlockFile("0000000110-20210728T105016.17-00000110a-00000109a-100-suffix"),
		MustNewOneBlockFile("0000000111-20210728T105016.18-00000111a-00000110a-100-suffix"),
		MustNewOneBlockFile("0000000112-20210728T105016.19-00000112a-00000111a-100-suffix"),
		MustNewOneBlockFile("0000000113-20210728T105016.20-00000113a-00000112a-100-suffix"),
		MustNewOneBlockFile("0000000114-20210728T105016.21-00000114a-00000113a-100-suffix"),
		MustNewOneBlockFile("0000000115-20210728T105016.22-00000115a-00000114a-100-suffix"),
		MustNewOneBlockFile("0000000116-20210728T105016.23-00000116a-00000115a-100-suffix"),
		MustNewOneBlockFile("0000000117-20210728T105016.24-00000117a-00000116a-100-suffix"),
		MustNewOneBlockFile("0000000118-20210728T105016.25-00000118a-00000117a-100-suffix"),
		MustNewOneBlockFile("0000000120-20210728T105016.26-00000120a-00000118a-100-suffix"),
	}

	bundler := NewBundler(5, 105)
	bundler.lastMergeOneBlockFile = &OneBlockFile{ID: "00000099a"}
	for _, f := range files {
		bundler.AddOneBlockFile(f)
	}

	completed, highestBlockLimit := bundler.BundleCompleted()
	require.True(t, completed)
	mergeableFiles := bundler.ToBundle(highestBlockLimit)
	bundler.Commit(highestBlockLimit)

	ids := ToIDs(mergeableFiles)
	require.Equal(t, []string{"00000100a", "00000101a", "00000102a", "00000102b", "00000103b", "00000103a", "00000104a"}, ids)

	completed, highestBlockLimit = bundler.BundleCompleted()
	require.True(t, completed)
	mergeableFiles = bundler.ToBundle(highestBlockLimit)
	bundler.Commit(highestBlockLimit)
	ids = ToIDs(mergeableFiles)
	require.Equal(t, []string{"00000106a", "00000107a", "00000108b", "00000109b", "00000108a", "00000109a"}, ids)

	completed, highestBlockLimit = bundler.BundleCompleted()
	require.True(t, completed)
	mergeableFiles = bundler.ToBundle(highestBlockLimit)
	bundler.Commit(highestBlockLimit)

	ids = ToIDs(mergeableFiles)
	require.Equal(t, []string{"00000110b", "00000110c", "00000111c", "00000110a", "00000111a", "00000112a", "00000113a", "00000114a"}, ids)

	completed, highestBlockLimit = bundler.BundleCompleted()
	require.True(t, completed)
	mergeableFiles = bundler.ToBundle(highestBlockLimit)
	bundler.Commit(highestBlockLimit)

	ids = ToIDs(mergeableFiles)
	require.Equal(t, []string{"00000115a", "00000116a", "00000117a", "00000118a"}, ids)

	longestOneBlockChain := bundler.LongestOneBlockFileChain()
	require.IsType(t, []*OneBlockFile{}, longestOneBlockChain)
	require.Equal(t, len(longestOneBlockChain), 19)

	longestChain := bundler.LongestChain()
	require.IsType(t, []string{}, longestChain)
	require.Equal(t, len(longestOneBlockChain), len(longestChain))
}

func TestBundler_BackToTheFuture(t *testing.T) {
	// load forkdb with some history ...
	files := []*OneBlockFile{
		MustTestNewMergedOneBlockFile("0000000094-20210728T105016.01-00000094a-00000093a-90-suffix"),
		MustTestNewMergedOneBlockFile("0000000095-20210728T105016.01-00000095a-00000094a-94-suffix"),
		MustTestNewMergedOneBlockFile("0000000096-20210728T105016.01-00000096a-00000095a-94-suffix"),
		MustTestNewMergedOneBlockFile("0000000097-20210728T105016.01-00000097a-00000096a-94-suffix"),
		MustTestNewMergedOneBlockFile("0000000098-20210728T105016.01-00000098a-00000097a-94-suffix"),
		MustTestNewMergedOneBlockFile("0000000099-20210728T105016.01-00000099a-00000098a-94-suffix"),

		MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-94-suffix"),
		MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-94-suffix"),
		MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-94-suffix"),
		MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-94-suffix"),
		MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-94-suffix"),
		MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-94-suffix"),
	}

	bundler := NewBundler(5, 105)
	bundler.lastMergeOneBlockFile = &OneBlockFile{ID: "00000099a"}
	for _, f := range files {
		bundler.AddOneBlockFile(f)
	}

	// Let's merge a first back of block from 100 to 104
	completed, highestBlockLimit := bundler.BundleCompleted()
	require.True(t, completed)
	mergeableFiles := bundler.ToBundle(highestBlockLimit)
	ids := ToIDs(mergeableFiles)
	require.Equal(t, []string{"00000100a", "00000101a", "00000102a", "00000103a", "00000104a"}, ids)
	bundler.Commit(highestBlockLimit)

	// Add a very old file
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000095-20210728T105015.01-00000095b-00000094a-90-suffix"))

	//that new file should not trigger a merge
	completed, highestBlockLimit = bundler.BundleCompleted()
	require.False(t, completed)

	// Add missing file for this back in time fork
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000096-20210728T105015.02-00000096b-00000095b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000097-20210728T105015.03-00000097b-00000096b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000098-20210728T105015.04-00000098b-00000097b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000099-20210728T105015.05-00000099b-00000098b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000100-20210728T105015.06-00000100b-00000099b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000101-20210728T105015.07-00000101b-00000100b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000102-20210728T105015.08-00000102b-00000101b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000103-20210728T105015.09-00000103b-00000102b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000104-20210728T105015.10-00000104b-00000103b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000105-20210728T105015.11-00000105b-00000104b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000106-20210728T105015.12-00000106b-00000105b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000107-20210728T105015.12-00000107b-00000106b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000108-20210728T105015.12-00000108b-00000107b-90-suffix"))
	bundler.AddOneBlockFile(MustNewOneBlockFile("000000109-20210728T105015.12-00000109b-00000108b-90-suffix"))

	completed, highestBlockLimit = bundler.BundleCompleted()
	//should not trigger merge yet
	require.False(t, completed)

	bundler.AddOneBlockFile(MustNewOneBlockFile("000000110-20210728T105015.12-00000110b-00000109b-90-suffix"))

	completed, highestBlockLimit = bundler.BundleCompleted()
	//here we go!
	require.True(t, completed)

	mergeableFiles = bundler.ToBundle(highestBlockLimit)
	ids = ToIDs(mergeableFiles)
	require.Equal(t, []string{
		"00000095b", "00000096b",
		"00000097b", "00000098b",
		"00000099b", "00000100b",
		"00000101b", "00000102b",
		"00000103b", "00000104b",
		"00000105b", "00000106b",
		"00000107b", "00000108b",
		"00000109b", "00000106a"}, ids)
}

func TestBundler_Purge(t *testing.T) {

	cases := []struct {
		name                      string
		files                     []*OneBlockFile
		lastMergerBlock           *OneBlockFile
		expectedFileToDeleteCount int
		expectedLongestFirstBlock string
		expectedLibID             string
		expectedCompleted         bool
	}{
		{
			name: "Sunny path",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-101-suffix"),
				MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-101-suffix"),
			},
			lastMergerBlock:           MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-101-suffix"),
			expectedCompleted:         true,
			expectedLongestFirstBlock: "00000101a",
			expectedFileToDeleteCount: 5,
			expectedLibID:             "00000101a",
		},
		{
			name: "Sunny path with fork",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102b-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103b-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-101-suffix"),
				MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-101-suffix"),
			},
			lastMergerBlock:           MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-101-suffix"),
			expectedLongestFirstBlock: "00000101a",
			expectedFileToDeleteCount: 7,
			expectedLibID:             "00000101a",
			expectedCompleted:         true,
		},
		{
			name: "Purger fork",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102b-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103b-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-103-suffix"),
				MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-101-suffix"),
			},
			lastMergerBlock:           MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-103-suffix"),
			expectedLongestFirstBlock: "00000103a",
			expectedFileToDeleteCount: 7,
			expectedLibID:             "00000103a",
			expectedCompleted:         true,
		},
		{
			name: "Purge nothing never merged anything",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
			},
			lastMergerBlock:           nil,
			expectedLongestFirstBlock: "00000100a",
			expectedFileToDeleteCount: 0,
			expectedLibID:             "00000100a",
		},
		{
			name: "Purge nothing",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
			},
			lastMergerBlock:           MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-90-suffix"),
			expectedLongestFirstBlock: "00000100a",
			expectedFileToDeleteCount: 0,
			expectedLibID:             "00000100a",
		},
		{
			name: "Purge multiple root",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100b-00000099b-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101b-00000100b-90-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102b-00000101b-90-suffix"),

				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
				MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
				MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-103-suffix"),
				MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-103-suffix"),
			},
			lastMergerBlock:           MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-103-suffix"),
			expectedLongestFirstBlock: "00000103a",
			expectedFileToDeleteCount: 8,
			expectedLibID:             "00000103a",
			expectedCompleted:         true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bundler := NewBundler(5, 105)
			for _, f := range c.files {
				bundler.AddOneBlockFile(f)
			}

			completed, highestBlockLimit := bundler.BundleCompleted()
			require.Equal(t, c.expectedCompleted, completed)
			if completed {
				bundler.Commit(highestBlockLimit)
			}

			bundler.Purge(func(purgedOneBlockFiles []*OneBlockFile) {
				require.Equal(t, c.expectedFileToDeleteCount, len(purgedOneBlockFiles))
			})

			require.Equal(t, c.expectedLibID, bundler.forkDB.LIBID())

			_, err := bundler.forkDB.Roots()
			if c.expectedLongestFirstBlock == "" {
				require.Errorf(t, err, "no link")
				return
			}
			require.NoError(t, err)

			longest := bundler.longestChain()
			require.Equal(t, c.expectedLongestFirstBlock, longest[0])
		})
	}
}

func TestBundler_Boostrap(t *testing.T) {
	mergeFiles := map[uint64][]*OneBlockFile{
		95: {
			MustNewOneBlockFile("0000000095-20210728T105016.07-00000095a-00000094a-90-suffix"),
			MustNewOneBlockFile("0000000096-20210728T105016.07-00000096a-00000095a-90-suffix"),
			MustNewOneBlockFile("0000000097-20210728T105016.07-00000097a-00000096a-95-suffix"),
			MustNewOneBlockFile("0000000098-20210728T105016.07-00000098a-00000097a-95-suffix"),
			MustNewOneBlockFile("0000000098-20210728T105016.07-00000098b-00000097a-95-suffix"),
			MustNewOneBlockFile("0000000099-20210728T105016.07-00000099a-00000098a-95-suffix"),
			MustNewOneBlockFile("0000000099-20210728T105016.07-00000099b-00000098b-95-suffix"),
		},
		100: {
			MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-95-suffix"),
			MustNewOneBlockFile("0000000100-20210728T105016.01-00000100b-00000099b-95-suffix"),
			MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-98-suffix"),
			MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-98-suffix"),
			MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-98-suffix"),
			MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-98-suffix"),
		},
		105: {
			MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-101-suffix"),
			MustNewOneBlockFile("0000000107-20210728T105016.09-00000107a-00000106a-101-suffix"),
			MustNewOneBlockFile("0000000108-20210728T105016.15-00000108a-00000107a-101-suffix"),
			MustNewOneBlockFile("0000000109-20210728T105016.16-00000109a-00000108a-106-suffix"),
		},
		110: {
			MustNewOneBlockFile("0000000110-20210728T105016.17-00000110a-00000109a-108-suffix"),
			MustNewOneBlockFile("0000000111-20210728T105016.18-00000111a-00000110a-108-suffix"),
			MustNewOneBlockFile("0000000112-20210728T105016.19-00000112a-00000111a-108-suffix"),
			MustNewOneBlockFile("0000000113-20210728T105016.20-00000113a-00000112a-111-suffix"),
			MustNewOneBlockFile("0000000114-20210728T105016.21-00000114a-00000113a-113-suffix"),
		},
	}

	testCases := []struct {
		name                            string
		firstExclusiveHighestBlockLimit uint64
		mergeFiles                      map[uint64][]*OneBlockFile
		expectedMergeFilesRead          []int
		expectedFirstBlockNum           uint64
		expectedErr                     string
	}{
		{
			name:                            "Sunny path",
			firstExclusiveHighestBlockLimit: 115,
			mergeFiles:                      mergeFiles,
			expectedMergeFilesRead:          []int{105, 105},
			expectedFirstBlockNum:           106,
		},
		{
			name:                            "First bundle with no merge file existing",
			firstExclusiveHighestBlockLimit: 10,
			mergeFiles:                      mergeFiles,
			expectedMergeFilesRead:          nil,
			expectedFirstBlockNum:           0,
			expectedErr:                     "loading one block files: failed to fetch merged file for low block num: 0: merge file not found",
		},
		{
			name:                            "First bundle with merge file",
			firstExclusiveHighestBlockLimit: 10,
			mergeFiles: map[uint64][]*OneBlockFile{
				0: {
					MustNewOneBlockFile("0000000001-20210728T105016.07-00000001a-00000000a-00-suffix"),
					MustNewOneBlockFile("0000000002-20210728T105016.07-00000002a-00000001a-00-suffix"),
					MustNewOneBlockFile("0000000003-20210728T105016.07-00000003a-00000002a-00-suffix"),
					MustNewOneBlockFile("0000000004-20210728T105016.07-00000004a-00000003a-01-suffix"),
				}},
			expectedMergeFilesRead: []int{0, 0},
			expectedFirstBlockNum:  1,
		},
		{
			name:                            "First bundle with no merge file existing",
			firstExclusiveHighestBlockLimit: 5,
			mergeFiles:                      mergeFiles,
			expectedMergeFilesRead:          nil,
			expectedFirstBlockNum:           0,
			expectedErr:                     "loading one block files: failed to fetch merged file for low block num: 0: merge file not found",
		},
		{
			name:                            "First and last from single file",
			firstExclusiveHighestBlockLimit: 115,
			mergeFiles:                      mergeFiles,
			expectedMergeFilesRead:          []int{105, 105},
			expectedFirstBlockNum:           106,
		},
		{
			name:                            "Find lib over 2 files",
			firstExclusiveHighestBlockLimit: 110,
			mergeFiles:                      mergeFiles,
			expectedMergeFilesRead:          []int{100, 95, 100},
			expectedFirstBlockNum:           95,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			bundler := NewBundler(5, c.firstExclusiveHighestBlockLimit)

			// No links in ForksDB yet
			longestChain := bundler.LongestChain()
			require.Nil(t, longestChain)
			firstBlockNum, err := bundler.LongestChainFirstBlockNum()
			require.Equal(t, firstBlockNum, uint64(0))
			require.Errorf(t, err, "no longest chain available")

			var mergeFileReads []int
			err = bundler.Bootstrap(func(lowBlockNum uint64) ([]*OneBlockFile, error) {
				mergeFileReads = append(mergeFileReads, int(lowBlockNum))

				if oneBlockFiles, found := c.mergeFiles[lowBlockNum]; found {
					return oneBlockFiles, nil
				}
				return nil, errors.New("merge file not found")
			})

			if c.expectedErr != "" {
				require.Error(t, err, c.expectedErr)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, c.expectedMergeFilesRead, mergeFileReads)

			firstBlockNum, err = bundler.LongestChainFirstBlockNum()
			require.NoError(t, err)
			require.Equal(t, int(c.expectedFirstBlockNum), int(firstBlockNum))
		})
	}
}

func TestBundler_IsBlockTooOld(t *testing.T) {
	oneBlockFiles := []*OneBlockFile{
		MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
		MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
		MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
		MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
		MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
		MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),
	}

	oneBlockFilesTwoRoots := []*OneBlockFile{
		MustNewOneBlockFile("000000095-20210728T105016.01-00000095b-00000094a-90-suffix"),
		MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
		MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
		MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
		MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
		MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
		MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),
	}

	cases := []struct {
		name           string
		files          []*OneBlockFile
		blockNum       uint64
		maxFixableFork uint64
		expectedResult bool
	}{
		{
			name:           "in the middle",
			files:          oneBlockFiles,
			blockNum:       102,
			expectedResult: false,
		},
		{
			name:           "in the future",
			files:          oneBlockFiles,
			blockNum:       200,
			expectedResult: false,
		},
		{
			name:           "at first block",
			files:          oneBlockFiles,
			blockNum:       100,
			expectedResult: false,
		},
		{
			name:           "before first block",
			files:          oneBlockFiles,
			blockNum:       99,
			expectedResult: true,
		},
		{
			name:           "too old",
			files:          oneBlockFiles,
			blockNum:       5,
			expectedResult: true,
		},
		{
			name:           "in the middle 2 roots",
			files:          oneBlockFilesTwoRoots,
			blockNum:       102,
			expectedResult: false,
		},
		{
			name:           "in the future 2 roots",
			files:          oneBlockFilesTwoRoots,
			blockNum:       200,
			expectedResult: false,
		},
		{
			name:           "at first block 2 roots",
			files:          oneBlockFilesTwoRoots,
			blockNum:       95,
			expectedResult: true,
		},
		{
			name:           "before first block 2 roots",
			files:          oneBlockFilesTwoRoots,
			blockNum:       94,
			expectedResult: true,
		},
		{
			name:           "too old 2 roots",
			files:          oneBlockFilesTwoRoots,
			blockNum:       5,
			expectedResult: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bundler := NewBundler(5, 105)

			if c.name == "in the middle" {
				// no root yet, can't be too old
				tooOld := bundler.IsBlockTooOld(c.blockNum)
				require.False(t, tooOld)
			}

			for _, f := range c.files {
				bundler.AddOneBlockFile(f)
			}

			tooOld := bundler.IsBlockTooOld(c.blockNum)
			require.Equal(t, c.expectedResult, tooOld)
		})
	}
}

func TestBundler_LongestChain_MultipleSameLength(t *testing.T) {
	c := struct {
		name                      string
		files                     []*OneBlockFile
		lastMergerBlock           *OneBlockFile
		expectedFileToDeleteCount int
		expectedLongestFirstBlock string
		expectedLibID             string
	}{
		name: "Sunny path with fork",
		files: []*OneBlockFile{
			MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
			MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),

			MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
			MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
			MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
			MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),

			MustNewOneBlockFile("0000000102-20210728T105016.03-00000102b-00000101a-100-suffix"),
			MustNewOneBlockFile("0000000103-20210728T105016.06-00000103b-00000102b-100-suffix"),
			MustNewOneBlockFile("0000000104-20210728T105016.07-00000104b-00000103b-100-suffix"),
			MustNewOneBlockFile("0000000106-20210728T105016.08-00000106b-00000104b-100-suffix"),
		},
	}

	bundler := NewBundler(5, 107)
	for _, f := range c.files {
		bundler.AddOneBlockFile(f)
	}

	longestChain := bundler.LongestChain()
	require.Nil(t, longestChain)
}

func TestBundler_LongestOneBlockFileChain_SameBlockTime(t *testing.T) {
	c := struct {
		name                      string
		files                     []*OneBlockFile
		lastMergerBlock           *OneBlockFile
		expectedFileToDeleteCount int
		expectedLongestFirstBlock string
		expectedLibID             string
	}{
		name: "Sunny path with fork",
		files: []*OneBlockFile{
			MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
			MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),

			MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-100-suffix"),
			MustNewOneBlockFile("0000000103-20210728T105016.06-00000103a-00000102a-100-suffix"),
			MustNewOneBlockFile("0000000104-20210728T105016.07-00000104a-00000103a-100-suffix"),
			MustNewOneBlockFile("0000000106-20210728T105016.08-00000106a-00000104a-100-suffix"),

			MustNewOneBlockFile("0000000102-20210728T105016.03-00000102b-00000101a-100-suffix"),
			MustNewOneBlockFile("0000000103-20210728T105016.06-00000103b-00000102b-100-suffix"),
		},
	}

	bundler := NewBundler(5, 107)
	blockTime := c.files[0].BlockTime

	for _, f := range c.files {
		f.BlockTime = blockTime
		bundler.AddOneBlockFile(f)
	}

	longest := bundler.LongestOneBlockFileChain()
	require.Equal(t, len(longest), 6)
}

func TestBundler_SetLIB(t *testing.T) {
	cases := []struct {
		name            string
		files           []*OneBlockFile
		lastMergerBlock *OneBlockFile
		expectedLibID   string
	}{
		{
			name: "Sunny path",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-101-suffix"),
			},
			expectedLibID: "00000101a",
		},
		{
			name: "With hole",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-101-suffix"),

				MustNewOneBlockFile("0000000200-20210728T105016.03-00000200a-00000199a-190-suffix"),
			},
			expectedLibID: "00000101a",
		}, {
			name: "With block after hole",
			files: []*OneBlockFile{
				MustNewOneBlockFile("0000000100-20210728T105016.01-00000100a-00000099a-90-suffix"),
				MustNewOneBlockFile("0000000101-20210728T105016.02-00000101a-00000100a-100-suffix"),
				MustNewOneBlockFile("0000000102-20210728T105016.03-00000102a-00000101a-101-suffix"),

				MustNewOneBlockFile("0000000200-20210728T105016.03-00000200a-00000199a-190-suffix"),

				MustNewOneBlockFile("0000000103-20210728T105016.03-00000103a-00000102a-102-suffix"),
			},
			expectedLibID: "00000102a",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			bundler := NewBundler(5, 107)

			for _, f := range c.files {
				bundler.AddOneBlockFile(f)
			}

			require.Equal(t, c.expectedLibID, bundler.forkDB.LIBID())
		})
	}

}
