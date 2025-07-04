// *** WARNING: this file was generated by pulumi-language-nodejs. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";

export interface EntityKafkaProperties {
    /**
     * All topic configurations including any server set configurations
     */
    allConfigs?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
    /**
     * Additional topic configurations
     */
    configs?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
    /**
     * Protobuf descriptor for key
     */
    keyDescriptor?: pulumi.Input<string>;
    /**
     * Number of partitions
     */
    topicPartitions?: pulumi.Input<number>;
    /**
     * Number of replicas
     */
    topicReplicas?: pulumi.Input<number>;
    /**
     * Protobuf descriptor for value
     */
    valueDescriptor?: pulumi.Input<string>;
}

export interface EntityKinesisProperties {
    /**
     * Protobuf descriptor for the value
     */
    descriptor?: pulumi.Input<string>;
    /**
     * Number of shards
     */
    kinesisShards?: pulumi.Input<number>;
}

export interface EntityPostgresProperties {
    details?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
}

export interface EntitySnowflakeProperties {
    details?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
}

export interface GetStoreConfluentKafka {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: string;
}

export interface GetStoreConfluentKafkaArgs {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: pulumi.Input<string>;
}

export interface GetStoreKafka {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: string;
    /**
     * Specifies if the store should be accessed over TLS
     */
    tlsDisabled?: boolean;
    /**
     * Specifies if the server CNAME should be validated against the certificate
     */
    tlsVerifyServerHostname?: boolean;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: string;
}

export interface GetStoreKafkaArgs {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: pulumi.Input<string>;
    /**
     * Specifies if the store should be accessed over TLS
     */
    tlsDisabled?: pulumi.Input<boolean>;
    /**
     * Specifies if the server CNAME should be validated against the certificate
     */
    tlsVerifyServerHostname?: pulumi.Input<boolean>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: pulumi.Input<string>;
}

export interface GetStoreKinesis {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: string;
}

export interface GetStoreKinesisArgs {
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: pulumi.Input<string>;
}

export interface GetStorePostgres {
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: string;
}

export interface GetStorePostgresArgs {
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: pulumi.Input<string>;
}

export interface GetStoreSnowflake {
    /**
     * Snowflake account ID
     */
    accountId?: string;
    /**
     * Access control role to use for the Store operations after connecting to Snowflake
     */
    roleName?: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: string;
    /**
     * Warehouse name to use for queries and other store operations that require compute resource
     */
    warehouseName?: string;
}

export interface GetStoreSnowflakeArgs {
    /**
     * Snowflake account ID
     */
    accountId?: pulumi.Input<string>;
    /**
     * Access control role to use for the Store operations after connecting to Snowflake
     */
    roleName?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris?: pulumi.Input<string>;
    /**
     * Warehouse name to use for queries and other store operations that require compute resource
     */
    warehouseName?: pulumi.Input<string>;
}

export interface SchemaRegistryConfluent {
    /**
     * Password to use when authenticating with confluent schema registry
     */
    password?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the schema registry
     */
    uris: pulumi.Input<string>;
    /**
     * Username to use when authenticating with confluent schema registry
     */
    username?: pulumi.Input<string>;
}

export interface SchemaRegistryConfluentCloud {
    /**
     * Key to use when authenticating with confluent cloud schema registry
     */
    key?: pulumi.Input<string>;
    /**
     * Secret to use when authenticating with confluent cloud schema registry
     */
    secret?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the schema registry
     */
    uris: pulumi.Input<string>;
}

export interface StoreConfluentKafka {
    /**
     * SASL hash function to use when authenticating with Confluent Kafka brokers
     */
    saslHashFunction: pulumi.Input<string>;
    /**
     * Password to use when authenticating with Apache Kafka brokers
     */
    saslPassword: pulumi.Input<string>;
    /**
     * Username to use when authenticating with Apache Kafka brokers
     */
    saslUsername: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: pulumi.Input<string>;
}

export interface StoreKafka {
    /**
     * AWS region where the Amazon MSK cluster is located
     */
    mskAwsRegion?: pulumi.Input<string>;
    /**
     * IAM role ARN to use when authenticating with Amazon MSK
     */
    mskIamRoleArn?: pulumi.Input<string>;
    /**
     * SASL hash function to use when authenticating with Apache Kafka brokers
     */
    saslHashFunction: pulumi.Input<string>;
    /**
     * Password to use when authenticating with Apache Kafka brokers
     */
    saslPassword?: pulumi.Input<string>;
    /**
     * Username to use when authenticating with Apache Kafka brokers
     */
    saslUsername?: pulumi.Input<string>;
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: pulumi.Input<string>;
    /**
     * CA certificate in PEM format
     */
    tlsCaCertFile?: pulumi.Input<string>;
    /**
     * Specifies if the store should be accessed over TLS
     */
    tlsDisabled?: pulumi.Input<boolean>;
    /**
     * Specifies if the server CNAME should be validated against the certificate
     */
    tlsVerifyServerHostname?: pulumi.Input<boolean>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: pulumi.Input<string>;
}

export interface StoreKinesis {
    /**
     * AWS IAM access key to use when authenticating with an Amazon Kinesis service
     */
    accessKeyId?: pulumi.Input<string>;
    /**
     * AWS account ID to use when authenticating with an Amazon Kinesis service
     */
    awsAccountId: pulumi.Input<string>;
    /**
     * AWS IAM secret access key to use when authenticating with an Amazon Kinesis service
     */
    secretAccessKey?: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: pulumi.Input<string>;
}

export interface StorePostgres {
    /**
     * Password to use when authenticating with a Postgres database
     */
    password: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: pulumi.Input<string>;
    /**
     * Username to use when authenticating with a Postgres database
     */
    username: pulumi.Input<string>;
}

export interface StoreSnowflake {
    /**
     * Snowflake account ID
     */
    accountId: pulumi.Input<string>;
    /**
     * Snowflake account's private key in PEM format
     */
    clientKeyFile: pulumi.Input<string>;
    /**
     * Passphrase for decrypting the Snowflake account's private key
     */
    clientKeyPassphrase: pulumi.Input<string>;
    /**
     * Snowflake cloud region name, where the account resources operate in
     */
    cloudRegion: pulumi.Input<string>;
    /**
     * Access control role to use for the Store operations after connecting to Snowflake
     */
    roleName: pulumi.Input<string>;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: pulumi.Input<string>;
    /**
     * User login name for the Snowflake account
     */
    username: pulumi.Input<string>;
    /**
     * Warehouse name to use for queries and other store operations that require compute resource
     */
    warehouseName: pulumi.Input<string>;
}
