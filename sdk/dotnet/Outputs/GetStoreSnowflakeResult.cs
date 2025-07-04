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
    public sealed class GetStoreSnowflakeResult
    {
        /// <summary>
        /// Snowflake account ID
        /// </summary>
        public readonly string AccountId;
        /// <summary>
        /// Access control role to use for the Store operations after connecting to Snowflake
        /// </summary>
        public readonly string RoleName;
        /// <summary>
        /// List of host:port URIs to connect to the store
        /// </summary>
        public readonly string Uris;
        /// <summary>
        /// Warehouse name to use for queries and other store operations that require compute resource
        /// </summary>
        public readonly string WarehouseName;

        [OutputConstructor]
        private GetStoreSnowflakeResult(
            string accountId,

            string roleName,

            string uris,

            string warehouseName)
        {
            AccountId = accountId;
            RoleName = roleName;
            Uris = uris;
            WarehouseName = warehouseName;
        }
    }
}
