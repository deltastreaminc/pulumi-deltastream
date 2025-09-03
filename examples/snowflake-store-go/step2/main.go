package main

import (
	"os"

	ds "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Step2 simulates a change: update warehouse name.
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		apiKey := os.Getenv("DELTASTREAM_API_KEY")
		server := os.Getenv("DELTASTREAM_SERVER")
		uris := os.Getenv("SNOWFLAKE_URIS")
		account := os.Getenv("SNOWFLAKE_ACCOUNT_ID")
		role := os.Getenv("SNOWFLAKE_ROLE_NAME")
		user := os.Getenv("SNOWFLAKE_USERNAME")
		wh := os.Getenv("SNOWFLAKE_WAREHOUSE_NAME_STEP2") // changed value
		cloudRegion := os.Getenv("SNOWFLAKE_CLOUD_REGION")
		clientKey := os.Getenv("SNOWFLAKE_CLIENT_KEY")
		if apiKey == "" || server == "" || uris == "" || account == "" || role == "" || user == "" || wh == "" || cloudRegion == "" || clientKey == "" {
			return nil
		}
		prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{Server: pulumi.String(server), ApiKey: pulumi.StringPtr(apiKey)})
		if err != nil {
			return err
		}
		store, err := ds.NewStore(ctx, "snowflake-store", &ds.StoreArgs{
			Name: pulumi.String("pulumi_snowflake_store"),
			Snowflake: (ds.SnowflakeInputsArgs{
				Uris:          pulumi.String(uris),
				AccountId:     pulumi.String(account),
				RoleName:      pulumi.String(role),
				Username:      pulumi.String(user),
				WarehouseName: pulumi.String(wh),
				CloudRegion:   pulumi.String(cloudRegion),
				ClientKey:     pulumi.String(clientKey),
			}).ToSnowflakeInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ctx.Export("store_type", store.Type)
		ctx.Export("store_state", store.State)
		ctx.Export("warehouse_name", pulumi.String(wh))
		return nil
	})
}
