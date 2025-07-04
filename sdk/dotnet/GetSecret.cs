// *** WARNING: this file was generated by pulumi-language-dotnet. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace DeltaStream.Pulumi
{
    public static class GetSecret
    {
        /// <summary>
        /// Secret resource
        /// 
        /// ## Example Usage
        /// 
        /// ```csharp
        /// using System.Collections.Generic;
        /// using System.Linq;
        /// using Pulumi;
        /// using Pulumi = Pulumi.Pulumi;
        /// 
        /// return await Deployment.RunAsync(() =&gt; 
        /// {
        ///     var example = Pulumi.GetSecret.Invoke(new()
        ///     {
        ///         Name = "example_secret",
        ///     });
        /// 
        /// });
        /// ```
        /// </summary>
        public static Task<GetSecretResult> InvokeAsync(GetSecretArgs args, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.InvokeAsync<GetSecretResult>("deltastream:index/getSecret:getSecret", args ?? new GetSecretArgs(), options.WithDefaults());

        /// <summary>
        /// Secret resource
        /// 
        /// ## Example Usage
        /// 
        /// ```csharp
        /// using System.Collections.Generic;
        /// using System.Linq;
        /// using Pulumi;
        /// using Pulumi = Pulumi.Pulumi;
        /// 
        /// return await Deployment.RunAsync(() =&gt; 
        /// {
        ///     var example = Pulumi.GetSecret.Invoke(new()
        ///     {
        ///         Name = "example_secret",
        ///     });
        /// 
        /// });
        /// ```
        /// </summary>
        public static Output<GetSecretResult> Invoke(GetSecretInvokeArgs args, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.Invoke<GetSecretResult>("deltastream:index/getSecret:getSecret", args ?? new GetSecretInvokeArgs(), options.WithDefaults());

        /// <summary>
        /// Secret resource
        /// 
        /// ## Example Usage
        /// 
        /// ```csharp
        /// using System.Collections.Generic;
        /// using System.Linq;
        /// using Pulumi;
        /// using Pulumi = Pulumi.Pulumi;
        /// 
        /// return await Deployment.RunAsync(() =&gt; 
        /// {
        ///     var example = Pulumi.GetSecret.Invoke(new()
        ///     {
        ///         Name = "example_secret",
        ///     });
        /// 
        /// });
        /// ```
        /// </summary>
        public static Output<GetSecretResult> Invoke(GetSecretInvokeArgs args, InvokeOutputOptions options)
            => global::Pulumi.Deployment.Instance.Invoke<GetSecretResult>("deltastream:index/getSecret:getSecret", args ?? new GetSecretInvokeArgs(), options.WithDefaults());
    }


    public sealed class GetSecretArgs : global::Pulumi.InvokeArgs
    {
        /// <summary>
        /// Name of the Secret
        /// </summary>
        [Input("name", required: true)]
        public string Name { get; set; } = null!;

        public GetSecretArgs()
        {
        }
        public static new GetSecretArgs Empty => new GetSecretArgs();
    }

    public sealed class GetSecretInvokeArgs : global::Pulumi.InvokeArgs
    {
        /// <summary>
        /// Name of the Secret
        /// </summary>
        [Input("name", required: true)]
        public Input<string> Name { get; set; } = null!;

        public GetSecretInvokeArgs()
        {
        }
        public static new GetSecretInvokeArgs Empty => new GetSecretInvokeArgs();
    }


    [OutputType]
    public sealed class GetSecretResult
    {
        /// <summary>
        /// Creation date of the Secret
        /// </summary>
        public readonly string CreatedAt;
        /// <summary>
        /// Description of the Secret
        /// </summary>
        public readonly string Description;
        /// <summary>
        /// The provider-assigned unique ID for this managed resource.
        /// </summary>
        public readonly string Id;
        /// <summary>
        /// Name of the Secret
        /// </summary>
        public readonly string Name;
        /// <summary>
        /// Owning role of the Secret
        /// </summary>
        public readonly string Owner;
        /// <summary>
        /// Status of the Secret
        /// </summary>
        public readonly string Status;
        /// <summary>
        /// Secret type. (Valid values: generic_string)
        /// </summary>
        public readonly string Type;
        /// <summary>
        /// Last update date of the Secret
        /// </summary>
        public readonly string UpdatedAt;

        [OutputConstructor]
        private GetSecretResult(
            string createdAt,

            string description,

            string id,

            string name,

            string owner,

            string status,

            string type,

            string updatedAt)
        {
            CreatedAt = createdAt;
            Description = description;
            Id = id;
            Name = name;
            Owner = owner;
            Status = status;
            Type = type;
            UpdatedAt = updatedAt;
        }
    }
}
