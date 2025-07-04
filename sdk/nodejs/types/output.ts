// *** WARNING: this file was generated by pulumi-language-nodejs. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";

export interface EntityKafkaProperties {
    /**
     * All topic configurations including any server set configurations
     */
    allConfigs: {[key: string]: string};
    /**
     * Additional topic configurations
     */
    configs: {[key: string]: string};
    /**
     * Protobuf descriptor for key
     */
    keyDescriptor: string;
    /**
     * Number of partitions
     */
    topicPartitions: number;
    /**
     * Number of replicas
     */
    topicReplicas: number;
    /**
     * Protobuf descriptor for value
     */
    valueDescriptor: string;
}

export interface EntityKinesisProperties {
    /**
     * Protobuf descriptor for the value
     */
    descriptor: string;
    /**
     * Number of shards
     */
    kinesisShards: number;
}

export interface EntityPostgresProperties {
    details: {[key: string]: string};
}

export interface EntitySnowflakeProperties {
    details: {[key: string]: string};
}

export interface GetDatabasesItem {
    /**
     * Creation date of the Database
     */
    createdAt: string;
    /**
     * Name of the Database
     */
    name: string;
    /**
     * Owning role of the Database
     */
    owner: string;
}

export interface GetNamespacesItem {
    /**
     * Creation date of the Namespace
     */
    createdAt: string;
    /**
     * Name of the Database
     */
    database: string;
    /**
     * Name of the Namespace
     */
    name: string;
    /**
     * Owning role of the Namespace
     */
    owner: string;
}

export interface GetObjectsObject {
    /**
     * Creation date of the object
     */
    createdAt: string;
    /**
     * Name of the Database
     */
    database: string;
    /**
     * Fully qualified name of the Object
     */
    fqn: string;
    /**
     * Name of the Object
     */
    name: string;
    /**
     * Name of the Namespace
     */
    namespace: string;
    /**
     * Owning role of the object
     */
    owner: string;
    /**
     * State of the Object
     */
    state: string;
    /**
     * Type of the Object
     */
    type: string;
    /**
     * Last update date of the object
     */
    updatedAt: string;
}

export interface GetSchemaRegistriesItem {
    /**
     * Creation date of the schema registry
     */
    createdAt: string;
    /**
     * Name of the schema registry
     */
    name: string;
    /**
     * Owning role of the schema registry
     */
    owner: string;
    /**
     * State of the schema registry
     */
    state: string;
    /**
     * Type of the schema registry
     */
    type: string;
    /**
     * Last update date of the schema registry
     */
    updatedAt: string;
}

export interface GetSecretsItem {
    /**
     * Creation date of the Secret
     */
    createdAt: string;
    /**
     * Description of the Secret
     */
    description: string;
    /**
     * Name of the Secret
     */
    name: string;
    /**
     * Owning role of the Secret
     */
    owner: string;
    /**
     * Status of the Secret
     */
    status: string;
    /**
     * Secret type. (Valid values: generic_string)
     */
    type: string;
    /**
     * Last update date of the Secret
     */
    updatedAt: string;
}

export interface GetStoreConfluentKafka {
    /**
     * Name of the schema registry
     */
    schemaRegistryName: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface GetStoreKafka {
    /**
     * Name of the schema registry
     */
    schemaRegistryName: string;
    /**
     * Specifies if the store should be accessed over TLS
     */
    tlsDisabled: boolean;
    /**
     * Specifies if the server CNAME should be validated against the certificate
     */
    tlsVerifyServerHostname: boolean;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface GetStoreKinesis {
    /**
     * Name of the schema registry
     */
    schemaRegistryName: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface GetStorePostgres {
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface GetStoreSnowflake {
    /**
     * Snowflake account ID
     */
    accountId: string;
    /**
     * Access control role to use for the Store operations after connecting to Snowflake
     */
    roleName: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
    /**
     * Warehouse name to use for queries and other store operations that require compute resource
     */
    warehouseName: string;
}

export interface GetStoresItem {
    /**
     * Creation date of the Store
     */
    createdAt: string;
    /**
     * Name of the Store
     */
    name: string;
    /**
     * Owning role of the Store
     */
    owner: string;
    /**
     * State of the Store
     */
    state: string;
    /**
     * Type of the Store
     */
    type: string;
    /**
     * Last update date of the Store
     */
    updatedAt: string;
}

export interface SchemaRegistryConfluent {
    /**
     * Password to use when authenticating with confluent schema registry
     */
    password?: string;
    /**
     * List of host:port URIs to connect to the schema registry
     */
    uris: string;
    /**
     * Username to use when authenticating with confluent schema registry
     */
    username?: string;
}

export interface SchemaRegistryConfluentCloud {
    /**
     * Key to use when authenticating with confluent cloud schema registry
     */
    key?: string;
    /**
     * Secret to use when authenticating with confluent cloud schema registry
     */
    secret?: string;
    /**
     * List of host:port URIs to connect to the schema registry
     */
    uris: string;
}

export interface StoreConfluentKafka {
    /**
     * SASL hash function to use when authenticating with Confluent Kafka brokers
     */
    saslHashFunction: string;
    /**
     * Password to use when authenticating with Apache Kafka brokers
     */
    saslPassword: string;
    /**
     * Username to use when authenticating with Apache Kafka brokers
     */
    saslUsername: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface StoreKafka {
    /**
     * AWS region where the Amazon MSK cluster is located
     */
    mskAwsRegion?: string;
    /**
     * IAM role ARN to use when authenticating with Amazon MSK
     */
    mskIamRoleArn?: string;
    /**
     * SASL hash function to use when authenticating with Apache Kafka brokers
     */
    saslHashFunction: string;
    /**
     * Password to use when authenticating with Apache Kafka brokers
     */
    saslPassword?: string;
    /**
     * Username to use when authenticating with Apache Kafka brokers
     */
    saslUsername?: string;
    /**
     * Name of the schema registry
     */
    schemaRegistryName?: string;
    /**
     * CA certificate in PEM format
     */
    tlsCaCertFile?: string;
    /**
     * Specifies if the store should be accessed over TLS
     */
    tlsDisabled: boolean;
    /**
     * Specifies if the server CNAME should be validated against the certificate
     */
    tlsVerifyServerHostname: boolean;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface StoreKinesis {
    /**
     * AWS IAM access key to use when authenticating with an Amazon Kinesis service
     */
    accessKeyId?: string;
    /**
     * AWS account ID to use when authenticating with an Amazon Kinesis service
     */
    awsAccountId: string;
    /**
     * AWS IAM secret access key to use when authenticating with an Amazon Kinesis service
     */
    secretAccessKey?: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
}

export interface StorePostgres {
    /**
     * Password to use when authenticating with a Postgres database
     */
    password: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
    /**
     * Username to use when authenticating with a Postgres database
     */
    username: string;
}

export interface StoreSnowflake {
    /**
     * Snowflake account ID
     */
    accountId: string;
    /**
     * Snowflake account's private key in PEM format
     */
    clientKeyFile: string;
    /**
     * Passphrase for decrypting the Snowflake account's private key
     */
    clientKeyPassphrase: string;
    /**
     * Snowflake cloud region name, where the account resources operate in
     */
    cloudRegion: string;
    /**
     * Access control role to use for the Store operations after connecting to Snowflake
     */
    roleName: string;
    /**
     * List of host:port URIs to connect to the store
     */
    uris: string;
    /**
     * User login name for the Snowflake account
     */
    username: string;
    /**
     * Warehouse name to use for queries and other store operations that require compute resource
     */
    warehouseName: string;
}

