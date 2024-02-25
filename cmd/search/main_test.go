package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_getConfigLoader(t *testing.T) {

	readFile := func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
		f, err := os.Open("testdata/search.json")
		require.Nil(t, err)
		return &s3.GetObjectOutput{Body: f}, nil
	}
	loadSearches := getConfigLoader(readFile, "bucket")
	searches, err := loadSearches(context.Background())
	require.Nil(t, err)
	require.Len(t, searches, 1)
}
