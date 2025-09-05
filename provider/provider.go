// Copyright 2025, DeltaStream Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package provider implements the DeltaStream Pulumi provider.
package provider

import (
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

// Name controls how this provider is referenced in package names and elsewhere.
const Name string = "deltastream"

// Provider creates a new instance of the DeltaStream provider.
func Provider() p.Provider {
	b := infer.NewProviderBuilder()
	b = b.WithDisplayName("pulumi-deltastream")
	b = b.WithDescription("A Pulumi provider for DeltaStream.")
	b = b.WithHomepage("https://www.deltastream.io")
	b = b.WithNamespace("deltastream")
	b = b.WithResources(
		infer.Resource(Database{}),
		infer.Resource(Namespace{}),
		infer.Resource(Store{}),
		infer.Resource(DeltaStreamObject{}),
		infer.Resource(Query{}),
	)
	b = b.WithFunctions(
		infer.Function(GetDatabase{}),
		infer.Function(GetDatabases{}),
		infer.Function(GetNamespace{}),
		infer.Function(GetNamespaces{}),
		infer.Function(GetStore{}),
		infer.Function(GetStores{}),
		infer.Function(GetObject{}),
		infer.Function(GetObjects{}),
	)
	b = b.WithConfig(infer.Config(&Config{}))
	b = b.WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{"provider": "index"})
	prov, err := b.Build()
	if err != nil {
		panic(fmt.Errorf("unable to build provider: %w", err))
	}
	return prov
}

// Config defines provider-level configuration
type Config struct {
	// API key for authentication (env: DELTASTREAM_API_KEY)
	APIKey *string `pulumi:"apiKey,optional"`
	// Server base URL, e.g. https://api.deltastream.io/v2 (env: DELTASTREAM_SERVER)
	Server *string `pulumi:"server"`
	// Skip TLS certificate verification (env: DELTASTREAM_INSECURE_SKIP_VERIFY)
	InsecureSkipVerify *bool `pulumi:"insecureSkipVerify,optional"`
	// Organization ID (UUID) or name (env: DELTASTREAM_ORGANIZATION)
	Organization *string `pulumi:"organization,optional"`
	// Role to execute statements as (env: DELTASTREAM_ROLE). Default: sysadmin
	Role *string `pulumi:"role,optional"`
	// Optional session ID (env: DELTASTREAM_SESSION_ID)
	SessionID *string `pulumi:"sessionId,optional"`
}
