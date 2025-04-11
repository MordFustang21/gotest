package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findTests(t *testing.T) {
	tests := []struct {
		file  string
		tests []Test
	}{
		{
			file: "testdata/nested_t_run_string.go",
			tests: []Test{
				{Name: "Test_Nested", File: "nested_t_run_string.go", FilePath: "testdata/nested_t_run_string.go", LineNumber: 7},
				{Name: "Test_Nested/L1", File: "nested_t_run_string.go", FilePath: "testdata/nested_t_run_string.go", LineNumber: 7},
				{Name: "Test_Nested/L1/L2", File: "nested_t_run_string.go", FilePath: "testdata/nested_t_run_string.go", LineNumber: 7},
			},
		},
		{
			file: "testdata/t_run_for_loop.go",
			tests: []Test{
				{Name: "Test_ForLoop", File: "t_run_for_loop.go", FilePath: "testdata/t_run_for_loop.go", LineNumber: 5},
				{Name: "Test_ForLoop/test1", File: "t_run_for_loop.go", FilePath: "testdata/t_run_for_loop.go", LineNumber: 5},
				{Name: "Test_ForLoop/test2", File: "t_run_for_loop.go", FilePath: "testdata/t_run_for_loop.go", LineNumber: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			found := findTests(tt.file)
			// verify tests are found
			assert.Equal(t, tt.tests, found)
		})
	}
}

func Benchmark_findTests(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tests, err := GetTestsFromDir("/Users/dlaird/projects/docuverse-server/", true)
		if err != nil {
			b.Fatal(err)
		}
		_ = tests
	}
}
