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
		expPath  string
	}{
		{
			name:     "non-nested directory",
			mainFile: "./cmd/example-event-post/main.go",
			expPath:  "build/example-event-post/bootstrap",
		},
		{
			name:     "nested directory",
			mainFile: "./cmd/workflows/BP003/PUB-036/main.go",
			expPath:  "build/workflows-BP003-PUB-036/bootstrap",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			outPath := getOutputPath(tc.mainFile)
			assert.Equal(t, tc.expPath, outPath)
		})
	}
}
