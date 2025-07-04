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
    public sealed class GetStoreKinesisResult
    {
        /// <summary>
        /// Name of the schema registry
        /// </summary>
        public readonly string SchemaRegistryName;
        /// <summary>
        /// List of host:port URIs to connect to the store
        /// </summary>
        public readonly string Uris;

        [OutputConstructor]
        private GetStoreKinesisResult(
            string schemaRegistryName,

            string uris)
        {
            SchemaRegistryName = schemaRegistryName;
            Uris = uris;
        }
    }
}
