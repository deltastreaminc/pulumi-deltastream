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

		pgUris := os.Getenv("POSTGRES_URIS")
		pgUser := os.Getenv("POSTGRES_USERNAME")
		pgPass := os.Getenv("POSTGRES_PASSWORD")
		pgDatabase := os.Getenv("POSTGRES_DATABASE")
		pgCdcSlotName := os.Getenv("POSTGRES_CDC_SLOT")
		if pgUris == "" || pgUser == "" || pgPass == "" || pgCdcSlotName == "" {
			return nil
		}

		uris := os.Getenv("SNOWFLAKE_URIS")
		account := os.Getenv("SNOWFLAKE_ACCOUNT_ID")
		role := os.Getenv("SNOWFLAKE_ROLE_NAME")
		user := os.Getenv("SNOWFLAKE_USERNAME")
		wh := os.Getenv("SNOWFLAKE_WAREHOUSE_NAME")
		cloudRegion := os.Getenv("SNOWFLAKE_CLOUD_REGION")
		clientKey := os.Getenv("SNOWFLAKE_CLIENT_KEY")
		if apiKey == "" || server == "" || uris == "" || account == "" || role == "" || user == "" || wh == "" || cloudRegion == "" || clientKey == "" {
			return nil
		}

		prov, err := deltastream.NewProvider(ctx, "deltastream", &deltastream.ProviderArgs{
			Server: pulumi.String(server),
			ApiKey: pulumi.StringPtr(apiKey),
		})
		if err != nil {
			return err
		}

		// Derive a short suffix from stack and project to avoid global name collisions across test runs.
		stackID := ctx.Stack() + "-" + ctx.Project()
		h := sha1.Sum([]byte(stackID))
		suffix := fmt.Sprintf("%x", h)[:6]

		dbName := fmt.Sprintf("pulumi_app_db_%s", suffix)
		kafkaStoreName := fmt.Sprintf("pulumi_app_kafka_%s", suffix)

		userlogChangelogName := fmt.Sprintf("pulumi_userlog_%s", suffix)
		pgTableName := "pulumi_userlog"
		pgCdcStreamName := fmt.Sprintf("pulumi_pg_cdc_%s", suffix)
		sfTableName := "sf_pulumi_userlog"

		// Create database and use public namespace
		db, err := deltastream.NewDatabase(ctx, "appDb", &deltastream.DatabaseArgs{
			Name: pulumi.String(dbName),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		// Use public namespace directly (auto-created with database)
		nsName := pulumi.String("public")

		// Kafka store
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

		sfStore, err := deltastream.NewStore(ctx, "snowflake-store", &deltastream.StoreArgs{
			Name: pulumi.String("pulumi_snowflake_store"),
			Snowflake: (deltastream.SnowflakeInputsArgs{
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

		pgStore, err := deltastream.NewStore(ctx, "pgstore", &deltastream.StoreArgs{
			Name: pulumi.String("pulumi_postgres_store"),
			Postgres: &deltastream.PostgresInputsArgs{
				Uris:        pulumi.String(pgUris),
				Username:    pulumi.String(pgUser),
				Password:    pulumi.String(pgPass),
				TlsDisabled: pulumi.Bool(false),
			},
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		userlog, err := deltastream.NewDeltaStreamObject(ctx, "usersChangelog", &deltastream.DeltaStreamObjectArgs{
			Database:  db.Name,
			Namespace: nsName,
			Store:     kafkaStore.Name,
			Sql: pulumi.Sprintf(
				`CREATE CHANGELOG %s (
					userid VARCHAR,
					gender VARCHAR,
					interests ARRAY<VARCHAR>,
					PRIMARY KEY (userid)
				) WITH (
					'topic'='users',
					'value.format'='json'
				);`,
				userlogChangelogName,
			),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		userlogTable, err := deltastream.NewDeltaStreamObject(ctx, "userlogTable", &deltastream.DeltaStreamObjectArgs{
			Database:  db.Name,
			Namespace: nsName,
			Store:     pgStore.Name,
			Sql: pulumi.Sprintf(
				`CREATE TABLE %s (
					userid VARCHAR,
					gender VARCHAR,
					PRIMARY KEY (userid)
				) WITH (
					'store'='%s',
					'postgresql.db.name'='%s',
					'postgresql.schema.name'='public',
					'postgresql.table.name'='%s'
				);`,
				pgTableName,
				pgStore.Name,
				pgDatabase,
				pgTableName,
			),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		userlogCdcStream, err := deltastream.NewDeltaStreamObject(ctx, "userlogCdcStream", &deltastream.DeltaStreamObjectArgs{
			Database:  db.Name,
			Namespace: nsName,
			Store:     pgStore.Name,
			Sql: pulumi.Sprintf(
				`CREATE STREAM %s(
					op VARCHAR,
					ts_ms BIGINT,
					"before" STRUCT<userid VARCHAR, gender VARCHAR>,
					"after"  STRUCT<userid VARCHAR, gender VARCHAR>,
					"source" STRUCT<db VARCHAR, "schema" VARCHAR, "table" VARCHAR, "lsn" BIGINT>
				) WITH (
					'store'='%s',
					'value.format'='json',
					'postgresql.db.name'='%s',
					'postgresql.schema.name'='public',
					'postgresql.table.name'='%s'
				);`,
				pgCdcStreamName,
				pgStore.Name,
				pgDatabase,
				pgTableName,
			),
		}, pulumi.Provider(prov), pulumi.DependsOn([]pulumi.Resource{userlogTable}))
		if err != nil {
			return err
		}

		sfTable, err := deltastream.NewDeltaStreamObject(ctx, "snowflakeTable", &deltastream.DeltaStreamObjectArgs{
			Database:  db.Name,
			Namespace: nsName,
			Store:     sfStore.Name,
			Sql: pulumi.Sprintf(
				`CREATE TABLE %s (
					userid VARCHAR,
					gender VARCHAR
				) WITH (
					'snowflake.db.name' = 'DEMO_DB',
					'snowflake.schema.name' = 'PUBLIC',
					'snowflake.table.name' = '%s'
				);`,
				sfTableName,
				sfTableName,
			),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		querySQL := pulumi.All(userlog.Fqn, userlogTable.Fqn).ApplyT(func(vals []interface{}) (string, error) {
			userlogChangelog := vals[0].(string)
			userlogTable := vals[1].(string)
			return fmt.Sprintf(`
				INSERT INTO %s SELECT userid, gender FROM %s; -- from Kafka userlog to Postgres
			`, userlogTable, userlogChangelog), nil
		}).(pulumi.StringOutput)
		query, err := deltastream.NewQuery(ctx, "userlogToPg", &deltastream.QueryArgs{
			SourceRelationFqns: pulumi.StringArray{userlog.Fqn},
			SinkRelationFqn:    userlogTable.Fqn,
			Sql:                querySQL,
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		appSQL := pulumi.All(userlogCdcStream.Fqn, sfTable.Fqn).ApplyT(func(vals []interface{}) (string, error) {
			userlogCdcStream := vals[0].(string)
			sfTable := vals[1].(string)
			return fmt.Sprintf(`
				BEGIN APPLICATION pulumi_test_app
					INSERT INTO %s SELECT after->userid, after->gender FROM %s WITH ( 'postgresql.slot.name' = '%s' ); -- from Postgres CDC to Snowflake
				END APPLICATION;
			`, sfTable, userlogCdcStream, pgCdcSlotName), nil
		}).(pulumi.StringOutput)

		app, err := deltastream.NewApplication(ctx, "testApplication", &deltastream.ApplicationArgs{
			SourceRelationFqns: pulumi.StringArray{
				userlogCdcStream.Fqn,
			},
			SinkRelationFqns: pulumi.StringArray{
				sfTable.Fqn,
			},
			Sql: appSQL,
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}

		ctx.Export("query_id", query.ID())

		ctx.Export("app_id", app.ApplicationId)
		ctx.Export("app_state", app.State)
		ctx.Export("app_owner", app.Owner)
		ctx.Export("app_sql", app.Sql)
		ctx.Export("app_created_at", app.CreatedAt)

		// Export physical relations
		ctx.Export("pageviews_fqn", userlog.Fqn)
		ctx.Export("pg_table_fqn", userlogTable.Fqn)
		ctx.Export("pg_cdc_fqn", userlogCdcStream.Fqn)
		ctx.Export("sf_table_fqn", sfTable.Fqn)

		return nil
	})
}
