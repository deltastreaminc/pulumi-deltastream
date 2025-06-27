// Copyright 2016-2024, Pulumi Corporation.
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

package deltastream

import (
	"context"
	"path"

	// Allow embedding bridge-metadata.json in the provider.
	_ "embed"

	pfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"

	"github.com/deltastreaminc/pulumi-deltastream/provider/pkg/version"
	deltastream "github.com/deltastreaminc/terraform-provider-deltastream/provider" // Import the upstream provider
)

// all of the token components used below.
const (
	// This variable controls the default name of the package in the package
	// registries for nodejs and python:
	mainPkg = "deltastream"
	// modules:
	mainMod = "index" // the deltastream module
)

//go:embed cmd/pulumi-resource-deltastream/bridge-metadata.json
var metadata []byte

// Provider returns additional overlaid schema and metadata associated with the provider.
func Provider() tfbridge.ProviderInfo {
	// Create a Pulumi provider mapping
	prov := tfbridge.ProviderInfo{
		P: pfbridge.ShimProvider(deltastream.New(version.Version)()),

		Name:              "deltastream",
		Version:           version.Version,
		DisplayName:       "DeltaStream", // The display name is used in the Pulumi Registry and CLI.
		Publisher:         "DeltaStream Inc.",
		LogoURL:           "http://deltastream-static-assets.s3-website-us-west-2.amazonaws.com/logo-single-purple.svg",
		PluginDownloadURL: "https://github.com/deltastreaminc/pulumi-deltastream/releases/download/v${VERSION}/",
		Description:       "A Pulumi package for creating and managing DeltaStream cloud resources.",
		Keywords:          []string{"deltastream", "category/infrastructure"},
		License:           "Apache-2.0",
		Homepage:          "https://www.deltastream.io",
		Repository:        "https://github.com/deltastreaminc/pulumi-deltastream",
		GitHubOrg:         "deltastreaminc",
		MetadataInfo:      tfbridge.NewProviderMetadata(metadata),
		Resources: map[string]*tfbridge.ResourceInfo{
			"deltastream_database": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["name"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_namespace": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["database"].String() + "." + state["name"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_object": {
				Tok: "deltastream:index/object:DeltaStreamObject",
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["fqn"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_query": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["query_id"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_store": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["name"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_schema_registry": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["name"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_secret": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["name"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
			"deltastream_entity": {
				ComputeID: func(_ context.Context, state resource.PropertyMap) (resource.ID, error) {
					return resource.ID(state["store"].String() + "::" + state["entity_path"].String()), nil
				},
				DeleteBeforeReplace: true,
			},
		},
		JavaScript: &tfbridge.JavaScriptInfo{
			RespectSchemaVersion: true,
			PackageName:          "@deltastream/pulumi-deltastream",
			Resolutions:          map[string]string{},
		},
		Python: &tfbridge.PythonInfo{
			// RespectSchemaVersion ensures the SDK is generated linking to the correct version of the provider.
			RespectSchemaVersion: true,
			// Enable modern PyProject support in the generated Python SDK.
			PyProject:   struct{ Enabled bool }{true},
			PackageName: "deltastream-pulumi",
		},
		Golang: &tfbridge.GolangInfo{
			// Set where the SDK is going to be published to.
			ImportBasePath: path.Join(
				"github.com/deltastreaminc/pulumi-deltastream/sdk/",
				tfbridge.GetModuleMajorVersion(version.Version),
				"go",
				mainPkg,
			),
			// Opt in to all available code generation features.
			GenerateResourceContainerTypes: true,
			GenerateExtraInputTypes:        true,
			// RespectSchemaVersion ensures the SDK is generated linking to the correct version of the provider.
			RespectSchemaVersion: true,
		},
		CSharp: &tfbridge.CSharpInfo{
			// RespectSchemaVersion ensures the SDK is generated linking to the correct version of the provider.
			RespectSchemaVersion: true,
			// Use a wildcard import so NuGet will prefer the latest possible version.
			PackageReferences: map[string]string{
				"Pulumi": "3.*",
			},
			Namespaces: map[string]string{
				"deltastream": "Pulumi",
			},
			RootNamespace: "DeltaStream",
		},
	}

	// MustComputeTokens maps all resources and datasources from the upstream provider into Pulumi.
	//
	// tokens.SingleModule puts every upstream item into your provider's main module.
	//
	// You shouldn't need to override anything, but if you do, use the [tfbridge.ProviderInfo.Resources]
	// and [tfbridge.ProviderInfo.DataSources].
	prov.MustComputeTokens(tokens.SingleModule("deltastream_", mainMod,
		tokens.MakeStandard(mainPkg)))

	prov.MustApplyAutoAliases()
	prov.SetAutonaming(255, "-")

	return prov
}
