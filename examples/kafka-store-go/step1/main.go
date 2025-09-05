package main

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	ds "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		apiKey := os.Getenv("DELTASTREAM_API_KEY")
		server := os.Getenv("DELTASTREAM_SERVER")
		bootstrap := os.Getenv("KAFKA_SASL_URIS")
		saslUser := os.Getenv("KAFKA_SASL_USERNAME")
		saslPass := os.Getenv("KAFKA_SASL_PASSWORD")
		if apiKey == "" || server == "" || bootstrap == "" || saslUser == "" || saslPass == "" {
			return nil // test harness will skip validation if outputs absent
		}
		prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{Server: pulumi.String(server), ApiKey: pulumi.StringPtr(apiKey)})
		if err != nil {
			return err
		}
		store, err := ds.NewStore(ctx, "kafka-store", &ds.StoreArgs{
			Name: pulumi.String("pulumi_kafka_store"),
			Kafka: (ds.KafkaInputsArgs{
				Uris:             pulumi.String(bootstrap),
				SaslHashFunction: pulumi.String("SHA512"),
				SaslUsername:     pulumi.StringPtr(saslUser),
				SaslPassword:     pulumi.StringPtr(saslPass),
			}).ToKafkaInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ctx.Export("store_auth_mode", pulumi.String("SHA512"))
		ctx.Export("store_state", store.State)
		return nil
	})
}
