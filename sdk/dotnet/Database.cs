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
    /// <summary>
    /// Database resource
    /// 
    /// ## Example Usage
    /// 
    /// ```csharp
    /// using System.Collections.Generic;
    /// using System.Linq;
    /// using Pulumi;
    /// using Pulumi = DeltaStream.Pulumi;
    /// 
    /// return await Deployment.RunAsync(() =&gt; 
    /// {
    ///     var example = new Pulumi.Database("example", new()
    ///     {
    ///         Name = "example_database",
    ///     });
    /// 
    /// });
    /// ```
    /// </summary>
    [PulumiResourceType("deltastream:index/database:Database")]
    public partial class Database : global::Pulumi.CustomResource
    {
        /// <summary>
        /// Creation date of the Database
        /// </summary>
        [Output("createdAt")]
        public Output<string> CreatedAt { get; private set; } = null!;

        /// <summary>
        /// Name of the Database
        /// </summary>
        [Output("name")]
        public Output<string> Name { get; private set; } = null!;

        /// <summary>
        /// Owning role of the Database
        /// </summary>
        [Output("owner")]
        public Output<string> Owner { get; private set; } = null!;


        /// <summary>
        /// Create a Database resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public Database(string name, DatabaseArgs? args = null, CustomResourceOptions? options = null)
            : base("deltastream:index/database:Database", name, args ?? new DatabaseArgs(), MakeResourceOptions(options, ""))
        {
        }

        private Database(string name, Input<string> id, DatabaseState? state = null, CustomResourceOptions? options = null)
            : base("deltastream:index/database:Database", name, state, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                PluginDownloadURL = "https://github.com/deltastreaminc/pulumi-deltastream/releases/download/v${VERSION}/",
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing Database resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="state">Any extra arguments used during the lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static Database Get(string name, Input<string> id, DatabaseState? state = null, CustomResourceOptions? options = null)
        {
            return new Database(name, id, state, options);
        }
    }

    public sealed class DatabaseArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// Name of the Database
        /// </summary>
        [Input("name")]
        public Input<string>? Name { get; set; }

        /// <summary>
        /// Owning role of the Database
        /// </summary>
        [Input("owner")]
        public Input<string>? Owner { get; set; }

        public DatabaseArgs()
        {
        }
        public static new DatabaseArgs Empty => new DatabaseArgs();
    }

    public sealed class DatabaseState : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// Creation date of the Database
        /// </summary>
        [Input("createdAt")]
        public Input<string>? CreatedAt { get; set; }

        /// <summary>
        /// Name of the Database
        /// </summary>
        [Input("name")]
        public Input<string>? Name { get; set; }

        /// <summary>
        /// Owning role of the Database
        /// </summary>
        [Input("owner")]
        public Input<string>? Owner { get; set; }

        public DatabaseState()
        {
        }
        public static new DatabaseState Empty => new DatabaseState();
    }
}
