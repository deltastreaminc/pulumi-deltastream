// *** WARNING: this file was generated by pulumi-language-dotnet. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace DeltaStream.Pulumi.Outputs
{

    [OutputType]
    public sealed class SchemaRegistryConfluentCloud
    {
        /// <summary>
        /// Key to use when authenticating with confluent cloud schema registry
        /// </summary>
        public readonly string? Key;
        /// <summary>
        /// Secret to use when authenticating with confluent cloud schema registry
        /// </summary>
        public readonly string? Secret;
        /// <summary>
        /// List of host:port URIs to connect to the schema registry
        /// </summary>
        public readonly string Uris;

        [OutputConstructor]
        private SchemaRegistryConfluentCloud(
            string? key,

            string? secret,

            string uris)
        {
            Key = key;
            Secret = secret;
            Uris = uris;
        }
    }
}
