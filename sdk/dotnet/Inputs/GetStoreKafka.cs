// *** WARNING: this file was generated by pulumi-language-dotnet. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace DeltaStream.Pulumi.Inputs
{

    public sealed class GetStoreKafkaArgs : global::Pulumi.InvokeArgs
    {
        /// <summary>
        /// Name of the schema registry
        /// </summary>
        [Input("schemaRegistryName", required: true)]
        public string SchemaRegistryName { get; set; } = null!;

        /// <summary>
        /// Specifies if the store should be accessed over TLS
        /// </summary>
        [Input("tlsDisabled", required: true)]
        public bool TlsDisabled { get; set; }

        /// <summary>
        /// Specifies if the server CNAME should be validated against the certificate
        /// </summary>
        [Input("tlsVerifyServerHostname", required: true)]
        public bool TlsVerifyServerHostname { get; set; }

        /// <summary>
        /// List of host:port URIs to connect to the store
        /// </summary>
        [Input("uris", required: true)]
        public string Uris { get; set; } = null!;

        public GetStoreKafkaArgs()
        {
        }
        public static new GetStoreKafkaArgs Empty => new GetStoreKafkaArgs();
    }
}
