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
		bootstrap := os.Getenv("KAFKA_MSK_AIM_URIS")
		roleArn := os.Getenv("KAFKA_MSK_IAM_ROLE_ARN")
		region := os.Getenv("KAFKA_MSK_AWS_REGION")
		if apiKey == "" || server == "" || bootstrap == "" || roleArn == "" || region == "" {
			return nil
		}
		prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{
			Server: pulumi.String(server),
			ApiKey: pulumi.StringPtr(apiKey),
		})
		if err != nil {
			return err
		}
		store, err := ds.NewStore(ctx, "kafka-store", &ds.StoreArgs{
			Name: pulumi.String("pulumi_kafka_store"),
			Kafka: (ds.KafkaInputsArgs{
				Uris:             pulumi.String(bootstrap),
				SaslHashFunction: pulumi.String("AWS_MSK_IAM"),
				MskIamRoleArn:    pulumi.StringPtr(roleArn),
				MskAwsRegion:     pulumi.StringPtr(region),
			}).ToKafkaInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ctx.Export("store_auth_mode", pulumi.String("AWS_MSK_IAM"))
		ctx.Export("store_state", store.State)
		return nil
	})
}
