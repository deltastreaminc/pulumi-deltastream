import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";

const config = new pulumi.Config();
const apiKey = "***";
const orgID = "***";
const serverUri = "***";
const kafkaUri  = "***";
const kafkaUser = "***";
const kafkaPassword = "***";
const postgresUri = "***";
const postgresUser = "***";
const postgresPassword = "***";
const snowflakeUri = "***";
const snowflakeAccountId = "***";
const snowflakeCloudRegion = "***";
const snowflakeWarehouseName = "***";
const snowflakeRoleName = "***";
const snowflakeUser = "***";
const snowflakeClientKeyFile = "***";
const snowflakeClientKeyPassphrase = "***";

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

    // Stores
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

    // Raw CDC stream for both tables
    const multiCdcSource = new deltastream.DeltaStreamObject("cdc_raw_source0708", {
        database: db.name,
        store: psqlStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE STREAM cdc_raw_source0708 (
            op VARCHAR,
            ts_ms BIGINT,
            "before" BYTES,
            "after" BYTES,
            "source" STRUCT<"schema" VARCHAR, "table" VARCHAR>
        ) WITH (
            'store'='${psqlStore.name}',
            'value.format'='json',
            'postgresql.db.name'='gradient',
            'postgresql.cdc.table.list'='public.customers,test.users'
        );
        `
    }, { provider });
    
    // Kafka sink for raw CDC events
    const cdcRawSink = new deltastream.DeltaStreamObject("cdc_raw_sink0708", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
          CREATE STREAM cdc_raw_sink0708 (
            op VARCHAR,
            src_schema VARCHAR,
            src_table VARCHAR,
            ts_ms BIGINT,
            "before" BYTES,
            "after" BYTES
          ) WITH (
            'store'='${kafkaStore.name}',
            'value.format'='json',
            'topic.partitions' = 1,
            'topic.replicas' = 3
          );
        `
      }, { provider });

    // Query to populate the sink stream
    const cdcRawSinkQuery = new deltastream.Query("cdc_raw_sink_query", {
        sourceRelationFqns: [multiCdcSource.fqn],
        sinkRelationFqn: cdcRawSink.fqn,
        sql: pulumi.interpolate`
          INSERT INTO ${cdcRawSink.fqn}
          SELECT
            op,
            source->"schema" AS src_schema,
            source->"table" AS src_table,
            ts_ms,
            "before",
            "after"
          FROM ${multiCdcSource.fqn} WITH ('postgresql.slot.name' = 'cdc0708')
          WHERE "after" IS NOT NULL;
        `
      }, { provider });

    // Filter customers
    const cdcRawCustomers = new deltastream.DeltaStreamObject("cdc_raw_customers0708", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE STREAM cdc_raw_customers0708 (
            op VARCHAR,
            src_schema VARCHAR,
            src_table VARCHAR,
            ts_ms BIGINT,
            "before" BYTES,
            "after" BYTES
        ) WITH (
            'store'='${kafkaStore.name}',
            'value.format'='json',
            'topic.partitions' = 1,
            'topic.replicas' = 3
        );
        `
    }, { provider });

    // Query to filter customers
    const cdcRawCustomersQuery = new deltastream.Query("cdc_raw_customers_query", {
        sourceRelationFqns: [cdcRawSink.fqn],
        sinkRelationFqn: cdcRawCustomers.fqn,
        sql: pulumi.interpolate`
        INSERT INTO ${cdcRawCustomers.fqn}
        SELECT * FROM ${cdcRawSink.fqn} WITH ('starting.position'='earliest')
        WHERE src_schema = 'public' AND src_table = 'customers';
        `
    }, { provider });
    
    // Filter users
    const cdcRawUsers = new deltastream.DeltaStreamObject("cdc_raw_users0708", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE STREAM cdc_raw_users0708 (
            op VARCHAR,
            src_schema VARCHAR,
            src_table VARCHAR,
            ts_ms BIGINT,
            "before" BYTES,
            "after" BYTES
        ) WITH (
            'store'='${kafkaStore.name}',
            'value.format'='json',
            'topic.partitions' = 1,
            'topic.replicas' = 3
        );
        `
    }, { provider });

    // Query to filter users
    const cdcRawUsersQuery = new deltastream.Query("cdc_raw_users_query", {
        sourceRelationFqns: [cdcRawSink.fqn],
        sinkRelationFqn: cdcRawUsers.fqn,
        sql: pulumi.interpolate`
        INSERT INTO ${cdcRawUsers.fqn}
        SELECT * FROM ${cdcRawSink.fqn} WITH ('starting.position'='earliest')
        WHERE src_schema = 'test' AND src_table = 'users';
        `
    }, { provider });
    
    // Structured stream: customers
    const customersStructured = new deltastream.DeltaStreamObject("customers_cdc0708", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE STREAM customers_cdc0708 (
            op STRING,
            ts_ms BIGINT,
            "before" STRUCT<
                \`id\` BIGINT,
                first_name STRING,
                last_name STRING,
                \`email\` STRING,
                biography STRING>,
            "after" STRUCT<
                \`id\` BIGINT,
                first_name STRING,
                last_name STRING,
                \`email\` STRING,
                biography STRING>
        ) WITH (
            'value.format' = 'json',
            'store' = '${kafkaStore.name}',
            'topic' = 'cdc_raw_customers0708'
        );
        `
    }, { provider, ignoreChanges: ["sql"] });
    
    // Structured stream: users
    const usersStructured = new deltastream.DeltaStreamObject("users_cdc0708", {
        database: db.name,
        store: kafkaStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE STREAM users_cdc0708 (
            op STRING,
            ts_ms BIGINT,
            "before" STRUCT<
                uid VARCHAR,
                \`name\` VARCHAR,
                city VARCHAR,
                balance BIGINT>,
            "after" STRUCT<
                uid VARCHAR,
                \`name\` VARCHAR,
                city VARCHAR,
                balance BIGINT>
        ) WITH (
            'value.format' = 'json',
            'store' = '${kafkaStore.name}',
            'topic' = 'cdc_raw_users0708'
        );
        `
    }, { provider, ignoreChanges: ["sql"] });


    // Snowflake sink: customers - create table first
    const customersSnowflakeTable = new deltastream.DeltaStreamObject("customers_cdc_sflk0708", {
        database: db.name,
        store: sfStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE TABLE customers_cdc_sflk0708 (
            \`id\` BIGINT,
            first_name STRING,
            last_name STRING,
            \`email\` STRING,
            biography STRING,
            event_write_time BIGINT,
            op STRING
        ) WITH (
            'store' ='${sfStore.name}',
            'snowflake.db.name'='DEMO_DB',
            'snowflake.schema.name'='PUBLIC'
        );
        `
    }, { provider, ignoreChanges: ["sql"] });
    
    // Query to populate customers table
    new deltastream.Query("customersToSnowflakeQuery", {
        sourceRelationFqns: [customersStructured.fqn],
        sinkRelationFqn: customersSnowflakeTable.fqn,
        sql: pulumi.interpolate`
        INSERT INTO ${customersSnowflakeTable.fqn}
        SELECT
            after->\`id\` AS \`id\`,
            after->first_name AS first_name,
            after->last_name AS last_name,
            after->\`email\` AS \`email\`,
            after->biography AS biography,
            ts_ms AS event_write_time,
            op
        FROM ${customersStructured.fqn} WITH ('starting.position'='earliest');
        `
    }, { provider, ignoreChanges: ["sql"] });
    
    // Snowflake sink: users - create table first
    const usersSnowflakeTable = new deltastream.DeltaStreamObject("users_cdc_sflk0708", {
        database: db.name,
        store: sfStore.name,
        namespace: "public",
        sql: pulumi.interpolate`
        CREATE TABLE users_cdc_sflk0708 (
            uid VARCHAR,
            \`name\` VARCHAR,
            city VARCHAR,
            balance BIGINT,
            event_write_time BIGINT,
            op STRING
        ) WITH (
            'store' ='${sfStore.name}',
            'snowflake.db.name'='DEMO_DB',
            'snowflake.schema.name'='PUBLIC'
        );
        `
    }, { provider, ignoreChanges: ["sql"] });
    
    new deltastream.Query("usersToSnowflakeQuery", {
        sourceRelationFqns: [usersStructured.fqn],
        sinkRelationFqn: usersSnowflakeTable.fqn,
        sql: pulumi.interpolate`
        INSERT INTO ${usersSnowflakeTable.fqn}
        SELECT
            after->uid AS uid,
            after->\`name\` AS \`name\`,
            after->city AS city,
            after->balance AS balance,
            ts_ms AS event_write_time,
            op
        FROM ${usersStructured.fqn} WITH ('starting.position'='earliest');
        `
    }, { provider, ignoreChanges: ["sql"] });
    
};
