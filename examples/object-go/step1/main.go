package main

import (
	"os"

	deltastream "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		apiKey := os.Getenv("DELTASTREAM_API_KEY")
		server := os.Getenv("DELTASTREAM_SERVER")
		iamUris := os.Getenv("KAFKA_MSK_IAM_URIS")
		iamRole := os.Getenv("KAFKA_MSK_IAM_ROLE_ARN")
		iamRegion := os.Getenv("KAFKA_MSK_AWS_REGION")
		if apiKey == "" || server == "" || iamUris == "" || iamRole == "" || iamRegion == "" {
			return nil
		}
		prov, err := deltastream.NewProvider(ctx, "deltastream", &deltastream.ProviderArgs{Server: pulumi.String(server), ApiKey: pulumi.StringPtr(apiKey)})
		if err != nil {
			return err
		}

		store, err := deltastream.NewStore(ctx, "kafka-iam-store", &deltastream.StoreArgs{
			Name: pulumi.String("pulumi_kafka_iam_store"),
			Kafka: (deltastream.KafkaInputsArgs{
				Uris:             pulumi.String(iamUris),
				MskIamRoleArn:    pulumi.StringPtr(iamRole),
				MskAwsRegion:     pulumi.StringPtr(iamRegion),
				SaslHashFunction: pulumi.String("NONE"),
			}).ToKafkaInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		db, err := deltastream.NewDatabase(ctx, "pageviews_db", &deltastream.DatabaseArgs{Name: pulumi.String("pulumi_pageviews_db")}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ns, err := deltastream.NewNamespace(ctx, "pageviews_ns", &deltastream.NamespaceArgs{Database: db.Name, Name: pulumi.String("pulumi_pageviews_ns")}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		sql := pulumi.String("CREATE STREAM PAGEVIEWS (viewtime BIGINT, userid VARCHAR, pageid VARCHAR) WITH ('topic'='pageviews', 'value.format'='json');")
		obj, err := deltastream.NewDeltaStreamObject(ctx, "pageviews_stream", &deltastream.DeltaStreamObjectArgs{
			Database:  db.Name,
			Namespace: ns.Name,
			Store:     store.Name,
			Sql:       sql,
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		ctx.Export("object_fqn", obj.Fqn)
		ctx.Export("object_type", obj.Type)
		ctx.Export("object_state", obj.State)
		ctx.Export("object_name", obj.Name)
		return nil
	})
}
