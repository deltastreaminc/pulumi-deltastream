package main

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	deltastream "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
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

		// Derive a short suffix from stack and project to avoid global name collisions across test runs.
		stackID := ctx.Stack() + "-" + ctx.Project()
		h := sha1.Sum([]byte(stackID))
		suffix := fmt.Sprintf("%x", h)[:6]

		dbName := fmt.Sprintf("pulumi_query_db_%s", suffix)
		kafkaStoreName := fmt.Sprintf("pulumi_query_kafka_%s", suffix)
		// Reuse existing Kafka topic 'pageviews' (pre-provisioned in test environment)
		pageviewsTopic := "pageviews"
		pageviews6Topic := fmt.Sprintf("pulumi_pageviews_6")

		// Create database and use public namespace
		db, err := deltastream.NewDatabase(ctx, "queryDb", &deltastream.DatabaseArgs{
			Name: pulumi.String(dbName),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		// Use public namespace directly (auto-created with database); avoid invoke during preview
		nsName := pulumi.String("public")

		// Kafka store (simplified example)
		kafkaStore, err := deltastream.NewStore(ctx, "kafkaStore", &deltastream.StoreArgs{
			Name: pulumi.String(kafkaStoreName),
			Kafka: (deltastream.KafkaInputsArgs{
				Uris:             pulumi.String(iamUris),
				MskIamRoleArn:    pulumi.StringPtr(iamRole),
				MskAwsRegion:     pulumi.StringPtr(iamRegion),
				SaslHashFunction: pulumi.String("AWS_MSK_IAM"),
			}).ToKafkaInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		// Pageviews stream
		pageviews, err := deltastream.NewDeltaStreamObject(ctx, "pageviewsStream", &deltastream.DeltaStreamObjectArgs{
			Database: db.Name, Namespace: nsName, Store: kafkaStore.Name,
			Sql: pulumi.Sprintf("CREATE STREAM PAGEVIEWS (viewtime BIGINT, userid VARCHAR, pageid VARCHAR) WITH ('topic'='%s','value.format'='json');", pageviewsTopic),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		// Sink stream (PAGEVIEWS_6) for the query
		pageviews6, err := deltastream.NewDeltaStreamObject(ctx, "pageviews6Stream", &deltastream.DeltaStreamObjectArgs{
			Database: db.Name, Namespace: nsName, Store: kafkaStore.Name,
			Sql: pulumi.Sprintf("CREATE STREAM PAGEVIEWS_6 (viewtime BIGINT, userid VARCHAR, pageid VARCHAR) WITH ('topic'='%s','value.format'='json','topic.partitions'=1,'topic.replicas'=2);", pageviews6Topic),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		// Query: copy from pageviews stream into changelog
		querySQL := pulumi.All(pageviews6.Fqn, pageviews.Fqn).ApplyT(func(vals []interface{}) (string, error) {
			sinkF := vals[0].(string)
			srcF := vals[1].(string)
			return fmt.Sprintf("INSERT INTO %s SELECT * FROM %s;", sinkF, srcF), nil
		}).(pulumi.StringOutput)

		q, err := deltastream.NewQuery(ctx, "pageviewsToPg", &deltastream.QueryArgs{
			SourceRelationFqns: pulumi.StringArray{pageviews.Fqn},
			SinkRelationFqn:    pageviews6.Fqn,
			Sql:                querySQL,
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		ctx.Export("pageviews_fqn", pageviews.Fqn)
		ctx.Export("pageviews_6_fqn", pageviews6.Fqn)
		ctx.Export("query_id", q.QueryId)
		ctx.Export("query_state", q.State)
		ctx.Export("query_sql", q.Sql)
		ctx.Export("query_owner", q.Owner)
		return nil
	})
}
