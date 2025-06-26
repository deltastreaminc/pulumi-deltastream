// Copyright 2024, Pulumi Corporation.  All rights reserved.
// g o:build nodejs || all
// + build nodejs all

package examples

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestBasicTs(t *testing.T) {
	apiKey := os.Getenv("DS_API_KEY")
	orgID := os.Getenv("DS_ORGANIZATION_ID")
	serverUri := os.Getenv("DS_SERVER_URI")

	opts := getJSBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              filepath.Join(getCwd(t), "basic-ts"),
		DestroyOnCleanup: true,
		Config: map[string]string{
			"DS_API_KEY":         apiKey,
			"DS_ORGANIZATION_ID": orgID,
			"DS_SERVER_URI":      serverUri,
		},
	})

	integration.ProgramTest(t, &opts)
}
