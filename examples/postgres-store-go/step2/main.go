package main

import (
	"os"
	"strings"

	deltastream "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		apiKey := os.Getenv("DELTASTREAM_API_KEY")
		server := os.Getenv("DELTASTREAM_SERVER")
		pgUris := os.Getenv("POSTGRES_URIS")
		pgUser := os.Getenv("POSTGRES_USERNAME")
		pgPass := os.Getenv("POSTGRES_PASSWORD")
		if apiKey == "" || server == "" || pgUris == "" || pgUser == "" || pgPass == "" {
			return nil // gated
		}
		// modify URI by appending param1=test if not already there
		if !strings.Contains(pgUris, "param1=") {
			sep := "?"
			if strings.Contains(pgUris, "?") {
				sep = "&"
			}
			pgUris = pgUris + sep + "param1=test"
		}
		prov, err := deltastream.NewProvider(ctx, "deltastream", &deltastream.ProviderArgs{Server: pulumi.String(server), ApiKey: pulumi.StringPtr(apiKey)})
		if err != nil {
			return err
		}
		store, err := deltastream.NewStore(ctx, "pgstore", &deltastream.StoreArgs{
			Name: pulumi.String("pulumi_postgres_store"),
			Postgres: &deltastream.PostgresInputsArgs{
				Uris:     pulumi.String(pgUris),
				Username: pulumi.String(pgUser),
				Password: pulumi.String(pgPass),
			},
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ctx.Export("postgres_uris", store.Postgres.ApplyT(func(p *deltastream.PostgresInputs) string {
			if p == nil {
				return ""
			}
			return p.Uris
		}).(pulumi.StringOutput))
		store.Postgres.ApplyT(func(p *deltastream.PostgresInputs) string {
			if p == nil {
				return ""
			}
			return p.Uris
		}).(pulumi.StringOutput).ApplyT(func(v string) error { ctx.Export("postgres_uris_plain", pulumi.String(v)); return nil })
		return nil
	})
}
