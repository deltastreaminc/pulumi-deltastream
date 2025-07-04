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
    public sealed class EntityKinesisProperties
    {
        /// <summary>
        /// Protobuf descriptor for the value
        /// </summary>
        public readonly string? Descriptor;
        /// <summary>
        /// Number of shards
        /// </summary>
        public readonly int? KinesisShards;

        [OutputConstructor]
        private EntityKinesisProperties(
            string? descriptor,

            int? kinesisShards)
        {
            Descriptor = descriptor;
            KinesisShards = kinesisShards;
        }
    }
}
