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
    public sealed class GetStoreKafkaResult
    {
        /// <summary>
        /// Name of the schema registry
        /// </summary>
        public readonly string SchemaRegistryName;
        /// <summary>
        /// Specifies if the store should be accessed over TLS
        /// </summary>
        public readonly bool TlsDisabled;
        /// <summary>
        /// Specifies if the server CNAME should be validated against the certificate
        /// </summary>
        public readonly bool TlsVerifyServerHostname;
        /// <summary>
        /// List of host:port URIs to connect to the store
        /// </summary>
        public readonly string Uris;

        [OutputConstructor]
        private GetStoreKafkaResult(
            string schemaRegistryName,

            bool tlsDisabled,

            bool tlsVerifyServerHostname,

            string uris)
        {
            SchemaRegistryName = schemaRegistryName;
            TlsDisabled = tlsDisabled;
            TlsVerifyServerHostname = tlsVerifyServerHostname;
            Uris = uris;
        }
    }
}
