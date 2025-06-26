// Copyright 2024, Pulumi Corporation.  All rights reserved.
//g o:build nodejs || all
// + build nodejs all

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestBasicTs(t *testing.T) {
	opts := getJSBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              filepath.Join(getCwd(t), "basic-ts"),
		DestroyOnCleanup: true,
	})

	integration.ProgramTest(t, &opts)
}
