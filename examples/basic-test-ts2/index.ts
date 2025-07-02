import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";

const config = new pulumi.Config();
const apiKey = "";
const orgID = "";
const serverUri = "";
const kafkaUri  = "";
const kafkaUser = "";
const kafkaPassword = "";
const postgresUri = "";
const postgresUser = "";
const postgresPassword = "";
const snowflakeUri = "";
const snowflakeAccountId = "";
const snowflakeCloudRegion = "";
const snowflakeWarehouseName = "";
const snowflakeRoleName = "";
const snowflakeUser = "";
const snowflakeClientKeyFile = "";
const snowflakeClientKeyPassphrase = "";

export = async () => {
    const provider = new deltastream.Provider("deltastream", {
        apiKey: apiKey,
        organization: orgID,
        server: serverUri,
        insecureSkipVerify: true
    });

    const db = new deltastream.Database("db", {
        name: `demo`,
    }, {
        provider: provider,
    })

    const kafkaStore = new deltastream.Store("kafkaStore", {
        name: `kafka_pulumi`,
        kafka: {
            uris: kafkaUri,
            saslHashFunction: "SHA512",
            saslUsername: kafkaUser,
            saslPassword: kafkaPassword,
            tlsDisabled: false,
        },
    }, {
        provider: provider,
    });

    const psqlStore = new deltastream.Store("psqlStore", {
        name: `postgresql`,
        postgres: {
            uris: postgresUri,
            username: postgresUser,
            password: postgresPassword
        },
    }, {
        provider: provider,
    });

    const pgTable = new deltastream.DeltaStreamObject("pgTableStream", {
        database: db.name,
        store: psqlStore.name,
        namespace: "public",
        sql: pulumi.interpolate `
            CREATE STREAM sourcetable_cdc(
                op VARCHAR,
                ts_ms BIGINT,
                "before" STRUCT<viewtime BIGINT, userid VARCHAR, pageid VARCHAR>, 
                "after"  STRUCT<viewtime BIGINT, userid VARCHAR, pageid VARCHAR>, 
                "source" STRUCT<db VARCHAR, "schema" VARCHAR, "table" VARCHAR, "lsn" BIGINT>
            ) WITH (
                'store'='${psqlStore.name}', 
                'value.format'='json',
                'postgresql.db.name'='gradient',
                'postgresql.schema.name'='public',
                'postgresql.table.name'='source_table'
            );`
    }, { provider: provider });

    const kafkaTable = new deltastream.DeltaStreamObject("kafkaTableStream", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
            CREATE STREAM sourcetable_stream(
                op VARCHAR,
                ts_ms BIGINT,
                "before" STRUCT<viewtime BIGINT, userid VARCHAR, pageid VARCHAR>,
                "after"  STRUCT<viewtime BIGINT, userid VARCHAR, pageid VARCHAR>,
                "source" STRUCT<db VARCHAR, "schema" VARCHAR, "table" VARCHAR, "lsn" BIGINT>
            ) WITH (
                'store'='${kafkaStore.name}',
                'value.format'='json',
                'topic.partitions' = 1,
                'topic.replicas' = 1
            );`
    }, { provider: provider });

    const cdcToKafkaQuery = new deltastream.Query("cdcToKafkaQuery", {
        sourceRelationFqns: [pgTable.fqn],
        sinkRelationFqn: kafkaTable.fqn,
        sql: pulumi.interpolate `INSERT INTO ${kafkaTable.name} SELECT * FROM ${pgTable.name} WITH ('postgresql.slot.name'='sourcetable_slot');`
    }, {
        provider: provider,
    });

    const sfStore = new deltastream.Store("sfStore", {
        name: `snowflake`,
        snowflake: {
            uris: snowflakeUri,
            accountId: snowflakeAccountId,
            cloudRegion: snowflakeCloudRegion,
            warehouseName: snowflakeWarehouseName,
            roleName: snowflakeRoleName,
            username: snowflakeUser,
            clientKeyFile: snowflakeClientKeyFile,
            clientKeyPassphrase: snowflakeClientKeyPassphrase
        }
    }, {
        provider: provider,
    });

    const snowflakeTable = new deltastream.DeltaStreamObject("snowflakeTable", {
        database: db.name,
        store: sfStore.name,
        namespace: "public",
        sql: pulumi.interpolate `
            CREATE TABLE desttable (
                viewtime BIGINT,
                userid VARCHAR,
                pageid VARCHAR
            ) WITH (
                'store'='${sfStore.name}',
                'snowflake.db.name'='DEMO_DB',
                'snowflake.schema.name'='PUBLIC'
            );`
    }, { provider: provider });

    const kafkaToSnowflakeQuery = new deltastream.Query("kafkaToSnowflakeQuery", {
        sourceRelationFqns: [kafkaTable.fqn],
        sinkRelationFqn: snowflakeTable.fqn,
        sql: pulumi.interpolate `INSERT INTO ${snowflakeTable.name} SELECT after->viewtime as viewtime, after->userid as userid, after->pageid as pageid FROM ${kafkaTable.name};`
    }, {
        provider: provider,
    });
};
