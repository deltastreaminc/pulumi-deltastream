// Copyright 2024, Pulumi Corporation.  All rights reserved.
package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/require"
)

func getJSBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"@deltastream/pulumi-deltastream",
		},
	})

	return baseJS
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	goDepRoot := os.Getenv("PULUMI_GO_DEP_ROOT")
	if goDepRoot == "" {
		var err error
		goDepRoot, err = filepath.Abs("../..")
		require.NoError(t, err)
	}
	rootSdkPath, err := filepath.Abs("../sdk")
	require.NoError(t, err)

	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			fmt.Sprintf("github.com/deltastreaminc/pulumi-deltastream/sdk=%s", rootSdkPath),
		},
		Env: []string{
			fmt.Sprintf("PULUMI_GO_DEP_ROOT=%s", goDepRoot),
		},
	})

	return baseJS
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}

func getBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	binPath, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Using binPath %s\n", binPath)
	return integration.ProgramTestOptions{
		LocalProviders: []integration.LocalDependency{
			{
				Package: "deltastream",
				Path:    binPath,
			},
		},
	}
}
