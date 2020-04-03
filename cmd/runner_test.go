package cmd

import (
	"bytes"
	"context"
	goflag "flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/gg/pkg/unittest"
)

var update = goflag.Bool("update", false, "update .golden file")

// Test_Cmd_run tests log parsing based on different flag configurations.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//     go test ./cmd -run Test_Cmd_run -update
//
func Test_Cmd_run(t *testing.T) {
	testCases := []struct {
		name    string
		flag    *flag
		fixture string
	}{
		{
			name: "case 0, drainer resource, loop first, grouped by loop, with non json logs inbetween",
			flag: &flag{
				colour: false,
				fields: []string{
					"loo",
					"mes",
				},
				group: "loo",
				selects: []string{
					"obj:qihx8",
					"res:dra",
				},
			},
			fixture: "text.json",
		},
		{
			name: "case 1, drainer resource, loop first, with non json logs inbetween",
			flag: &flag{
				colour: false,
				fields: []string{
					"loo",
					"mes",
				},
				selects: []string{
					"obj:qihx8",
					"res:dra",
				},
			},
			fixture: "text.json",
		},
		{
			name: "case 2, error stack of warning logs",
			flag: &flag{
				colour: false,
				fields: []string{
					"res",
					"sta",
				},
				selects: []string{
					"lev:war",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 3, error stack of warning logs with annotation",
			flag: &flag{
				colour: false,
				fields: []string{
					"res",
					"ann",
					"sta",
				},
				selects: []string{
					"lev:war",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 4, resource, error logs",
			flag: &flag{
				colour: false,
				fields: []string{
					"res",
				},
			},
			fixture: "error.json",
		},
		// Note that the golden file is empty because the selection does not match
		// anything.
		{
			name: "case 5, resource and annotation, error logs",
			flag: &flag{
				colour: false,
				fields: []string{
					"res",
					"ann",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 6, message and resource, error logs",
			flag: &flag{
				colour: false,
				fields: []string{
					"mes",
					"res",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 7, collector errors",
			flag: &flag{
				colour: false,
				selects: []string{
					"lev:err",
					"mes:metr",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 8, microkit errors",
			flag: &flag{
				colour: false,
				selects: []string{
					"cal:mic",
					"lev:err",
				},
			},
			fixture: "error.json",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var stdin io.Reader
			{
				p := filepath.Join("fixture", tc.fixture)
				file, err := os.Open(p)
				if err != nil {
					t.Fatal(microerror.JSON(err))
				}

				stdin = file
			}

			var stdout *bytes.Buffer
			{
				stdout = &bytes.Buffer{}
			}

			{
				ru := &runner{
					flag:   tc.flag,
					stdin:  stdin,
					stdout: stdout,
				}

				err := ru.run(context.Background(), nil, nil)
				if err != nil {
					t.Fatal(microerror.JSON(err))
				}
			}

			{
				p := filepath.Join("testdata", unittest.NormalizeFileName(tc.name)+".golden")

				if *update {
					err := ioutil.WriteFile(p, stdout.Bytes(), 0644)
					if err != nil {
						t.Fatal(microerror.JSON(err))
					}
				}
				goldenFile, err := ioutil.ReadFile(p)
				if err != nil {
					t.Fatal(microerror.JSON(err))
				}

				if !bytes.Equal(stdout.Bytes(), goldenFile) {
					t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), stdout.String()))
				}
			}
		})
	}
}
