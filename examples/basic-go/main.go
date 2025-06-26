package main

import (
	"github.com/deltastreaminc/pulumi-deltastream/sdk/go/deltastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		provider, err := deltastream.NewProvider(ctx, "deltastream", &deltastream.ProviderArgs{
			ApiKey:       pulumi.String("api-token"),
			Organization: pulumi.String("org-id"),
			Server:       pulumi.String("server-uri"),
		})
		if err != nil {
			return err
		}

		_, err = deltastream.NewDatabase(ctx, "pulumi-database", &deltastream.DatabaseArgs{
			Name: pulumi.String("pulumi-database"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}
		return nil
	})
}
