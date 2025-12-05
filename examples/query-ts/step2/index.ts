import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";
import * as crypto from "crypto";

// Validate required environment variables
if (!process.env.DELTASTREAM_API_KEY) {
    throw new Error("DELTASTREAM_API_KEY is required");
}
if (!process.env.DELTASTREAM_SERVER) {
    throw new Error("DELTASTREAM_SERVER is required");
}
if (!process.env.KAFKA_MSK_IAM_URIS) {
    throw new Error("KAFKA_MSK_IAM_URIS is required");
}
if (!process.env.KAFKA_MSK_IAM_ROLE_ARN) {
    throw new Error("KAFKA_MSK_IAM_ROLE_ARN is required");
}
if (!process.env.KAFKA_MSK_AWS_REGION) {
    throw new Error("KAFKA_MSK_AWS_REGION is required");
}

const provider = new deltastream.Provider("deltastream", {
    server: process.env.DELTASTREAM_SERVER,
    apiKey: process.env.DELTASTREAM_API_KEY,
});

// Derive a short suffix from stack and project to avoid global name collisions across test runs
const stackID = `${pulumi.getStack()}-${pulumi.getProject()}`;
const hash = crypto.createHash('sha1').update(stackID).digest('hex');
const suffix = hash.substring(0, 6);

const dbName = `pulumi_query_db_${suffix}`;
const kafkaStoreName = `pulumi_query_kafka_${suffix}`;
// Reuse existing Kafka topic 'pageviews' (pre-provisioned in test environment)
const pageviewsTopic = "pageviews";
const pageviews6Topic = "pulumi_pageviews_6";

// Create database and use public namespace
const db = new deltastream.Database("queryDb", {
    name: dbName,
}, { provider });

// Use public namespace directly (auto-created with database)
const nsName = "public";

// Kafka store with AWS MSK IAM authentication
const kafkaStore = new deltastream.Store("kafkaStore", {
    name: kafkaStoreName,
    kafka: {
        uris: process.env.KAFKA_MSK_IAM_URIS,
        mskIamRoleArn: process.env.KAFKA_MSK_IAM_ROLE_ARN,
        mskAwsRegion: process.env.KAFKA_MSK_AWS_REGION,
        saslHashFunction: "AWS_MSK_IAM",
    },
}, { provider });

// Pageviews stream (source) - MODIFIED: changed WITH clause order
const pageviews = new deltastream.DeltaStreamObject("pageviewsStream", {
    database: db.name,
    namespace: nsName,
    store: kafkaStore.name,
    sql: pulumi.interpolate`CREATE STREAM PAGEVIEWS (viewtime BIGINT, userid VARCHAR, pageid VARCHAR) WITH ('value.format'='json', 'topic'='${pageviewsTopic}');`,
}, { provider });

// Sink stream (PAGEVIEWS_6) for the query
const pageviews6 = new deltastream.DeltaStreamObject("pageviews6Stream", {
    database: db.name,
    namespace: nsName,
    store: kafkaStore.name,
    sql: pulumi.interpolate`CREATE STREAM PAGEVIEWS_6 (viewtime BIGINT, userid VARCHAR, pageid VARCHAR) WITH ('topic'='${pageviews6Topic}','value.format'='json','topic.partitions'=1,'topic.replicas'=2);`,
}, { provider });

// Query: copy from pageviews stream into sink stream
const querySQL = pulumi.all([pageviews6.fqn, pageviews.fqn]).apply(([sinkFqn, srcFqn]) => 
    `INSERT INTO ${sinkFqn} SELECT * FROM ${srcFqn};`
);

const q = new deltastream.Query("pageviewsToPg", {
    sourceRelationFqns: [pageviews.fqn],
    sinkRelationFqn: pageviews6.fqn,
    sql: querySQL,
}, { provider });

// Exports for test validation
export const pageviews_fqn = pageviews.fqn;
export const pageviews_created_at = pageviews.createdAt;
export const pageviews_6_fqn = pageviews6.fqn;
export const query_id = q.queryId;
export const query_state = q.state;
export const query_sql = q.sql;
export const query_owner = q.owner;
