package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getInputDirectory(t *testing.T) {
	mainFile := "./cmd/example-event-post/main.go"
	inputDir := getInputDirectory(mainFile)
	assert.Equal(t, "./cmd/example-event-post", inputDir)
}

func Test_getOutputPath(t *testing.T) {
	testcases := []struct {
		name     string
		mainFile string
		expDir   string
	}{
		{
			name:     "non-nested directory",
			mainFile: "./cmd/example-event-post/main.go",
			expDir:   "build/example-event-post",
		},
		{
			name:     "nested directory",
			mainFile: "./cmd/workflows/BP003/PUB-036/main.go",
			expDir:   "build/workflows-BP003-PUB-036",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			outDir := getOutputDir(tc.mainFile)
			assert.Equal(t, tc.expDir, outDir)
		})
	}
}
