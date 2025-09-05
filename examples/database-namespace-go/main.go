package main

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	ds "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{
			Server:             pulumi.String(os.Getenv("DELTASTREAM_SERVER")),
			ApiKey:             pulumi.String(os.Getenv("DELTASTREAM_API_KEY")),
			Organization:       pulumi.StringPtr(os.Getenv("DELTASTREAM_ORGANIZATION")),
			Role:               pulumi.StringPtr(os.Getenv("DELTASTREAM_ROLE")),
			InsecureSkipVerify: pulumi.BoolPtr(os.Getenv("DELTASTREAM_INSECURE_SKIP_VERIFY") == "true"),
			SessionId:          pulumi.StringPtr(os.Getenv("DELTASTREAM_SESSION_ID")),
		})
		if err != nil {
			return err
		}

		db, err := ds.NewDatabase(ctx, "pulumi-db", &ds.DatabaseArgs{
			Name: pulumi.String("pulumi_db_go"),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		if !ctx.DryRun() {
			// Use an Apply to perform the lookup after the resource is fully registered.
			pulumi.All(db.Name).ApplyT(func(all []interface{}) error {
				lookup, err := ds.LookupDatabase(ctx, &ds.LookupDatabaseArgs{Name: "pulumi_db_go"}, pulumi.Provider(prov))
				if err != nil {
					return err
				}
				ctx.Export("db_owner", pulumi.String(lookup.Owner))
				return nil
			})
			dbs, err := ds.GetDatabases(ctx, &ds.GetDatabasesArgs{}, pulumi.Provider(prov))
			if err != nil {
				return err
			}
			ctx.Export("dbs_count", pulumi.Int(len(dbs.Databases)))
		}

		// Create a namespace in the database
		ns, err := ds.NewNamespace(ctx, "pulumi-ns", &ds.NamespaceArgs{
			Database: db.Name,
			Name:     pulumi.String("pulumi_namespace_go"),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		// Namespace lookups (skip during preview)
		if !ctx.DryRun() {
			pulumi.All(db.Name, ns.Name).ApplyT(func(_ []any) error {
				lookup, err := ds.LookupNamespace(ctx, &ds.LookupNamespaceArgs{Database: "pulumi_db_go", Name: "pulumi_namespace_go"}, pulumi.Provider(prov))
				if err != nil {
					return err
				}
				ctx.Export("namespace_owner", pulumi.String(lookup.Owner))
				list, err := ds.GetNamespaces(ctx, &ds.GetNamespacesArgs{Database: "pulumi_db_go"}, pulumi.Provider(prov))
				if err != nil {
					return err
				}
				ctx.Export("namespaces_count", pulumi.Int(len(list.Namespaces)))
				return nil
			})
		}

		ctx.Export("db_createdAt", db.CreatedAt)
		ctx.Export("ns_createdAt", ns.CreatedAt)
		return nil
	})
}
