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
			name: "case 0, drainer resource, time first, grouped by loop, text",
			flag: &flag{
				fields: []string{
					"tim",
					"mes",
				},
				group:  "loo",
				output: "text",
				selects: []string{
					"obj:qihx8",
					"res:dra",
				},
			},
			fixture: "basic.json",
		},
		{
			name: "case 1, drainer resource, loop first, grouped by loop, json, with non json logs inbetween",
			flag: &flag{
				fields: []string{
					"loo",
					"mes",
				},
				group:  "loo",
				output: "json",
				selects: []string{
					"obj:qihx8",
					"res:dra",
				},
			},
			fixture: "text.json",
		},
		{
			name: "case 2, drainer resource, loop first, json, with non json logs inbetween",
			flag: &flag{
				fields: []string{
					"loo",
					"mes",
				},
				output: "json",
				selects: []string{
					"obj:qihx8",
					"res:dra",
				},
			},
			fixture: "text.json",
		},
		{
			name: "case 3, service resource, resource first, text",
			flag: &flag{
				fields: []string{
					"res",
					"tim",
				},
				output: "text",
				selects: []string{
					"obj:hixh7",
					"res:ser",
				},
			},
			fixture: "basic.json",
		},
		{
			name: "case 4, all resources for machine deployment, grouped by loop, text",
			flag: &flag{
				fields: []string{
					"res",
					"mes",
				},
				group:  "loo",
				output: "text",
				selects: []string{
					"obj:qihx8",
					"con:mac",
					"res:.*",
				},
			},
			fixture: "basic.json",
		},
		{
			name: "case 5, error stack of warning logs, json",
			flag: &flag{
				fields: []string{
					"res",
					"sta",
				},
				output: "json",
				selects: []string{
					"lev:war",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 6, error stack of warning logs with annotation, json",
			flag: &flag{
				fields: []string{
					"res",
					"ann",
					"sta",
				},
				output: "json",
				selects: []string{
					"lev:war",
				},
			},
			fixture: "error.json",
		},
		{
			name: "case 7, resource, error logs, json",
			flag: &flag{
				fields: []string{
					"res",
				},
				output: "json",
			},
			fixture: "error.json",
		},
		{
			name: "case 8, resource and annotation, error logs, json",
			flag: &flag{
				fields: []string{
					"res",
					"ann",
				},
				output: "json",
			},
			fixture: "error.json",
		},
		{
			name: "case 9, collector errors, json",
			flag: &flag{
				output: "json",
				selects: []string{
					"lev:err",
					"mes:metr",
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
					t.Fatal(err)
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
					t.Fatal(err)
				}
			}

			{
				p := filepath.Join("testdata", unittest.NormalizeFileName(tc.name)+".golden")

				if *update {
					err := ioutil.WriteFile(p, stdout.Bytes(), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				goldenFile, err := ioutil.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(stdout.Bytes(), goldenFile) {
					t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), stdout.String()))
				}
			}
		})
	}
}
