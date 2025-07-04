// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package deltastream

import (
	"context"
	"reflect"

	"github.com/deltastreaminc/pulumi-deltastream/sdk/go/deltastream/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Schema registry datasource
func LookupSchemaRegistry(ctx *pulumi.Context, args *LookupSchemaRegistryArgs, opts ...pulumi.InvokeOption) (*LookupSchemaRegistryResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv LookupSchemaRegistryResult
	err := ctx.Invoke("deltastream:index/getSchemaRegistry:getSchemaRegistry", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getSchemaRegistry.
type LookupSchemaRegistryArgs struct {
	// Name of the schema registry
	Name string `pulumi:"name"`
}

// A collection of values returned by getSchemaRegistry.
type LookupSchemaRegistryResult struct {
	// Creation date of the schema registry
	CreatedAt string `pulumi:"createdAt"`
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// Name of the schema registry
	Name string `pulumi:"name"`
	// Owning role of the schema registry
	Owner string `pulumi:"owner"`
	// State of the schema registry
	State string `pulumi:"state"`
	// Type of the schema registry
	Type string `pulumi:"type"`
	// Last update date of the schema registry
	UpdatedAt string `pulumi:"updatedAt"`
}

func LookupSchemaRegistryOutput(ctx *pulumi.Context, args LookupSchemaRegistryOutputArgs, opts ...pulumi.InvokeOption) LookupSchemaRegistryResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupSchemaRegistryResultOutput, error) {
			args := v.(LookupSchemaRegistryArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("deltastream:index/getSchemaRegistry:getSchemaRegistry", args, LookupSchemaRegistryResultOutput{}, options).(LookupSchemaRegistryResultOutput), nil
		}).(LookupSchemaRegistryResultOutput)
}

// A collection of arguments for invoking getSchemaRegistry.
type LookupSchemaRegistryOutputArgs struct {
	// Name of the schema registry
	Name pulumi.StringInput `pulumi:"name"`
}

func (LookupSchemaRegistryOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupSchemaRegistryArgs)(nil)).Elem()
}

// A collection of values returned by getSchemaRegistry.
type LookupSchemaRegistryResultOutput struct{ *pulumi.OutputState }

func (LookupSchemaRegistryResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupSchemaRegistryResult)(nil)).Elem()
}

func (o LookupSchemaRegistryResultOutput) ToLookupSchemaRegistryResultOutput() LookupSchemaRegistryResultOutput {
	return o
}

func (o LookupSchemaRegistryResultOutput) ToLookupSchemaRegistryResultOutputWithContext(ctx context.Context) LookupSchemaRegistryResultOutput {
	return o
}

// Creation date of the schema registry
func (o LookupSchemaRegistryResultOutput) CreatedAt() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.CreatedAt }).(pulumi.StringOutput)
}

// The provider-assigned unique ID for this managed resource.
func (o LookupSchemaRegistryResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.Id }).(pulumi.StringOutput)
}

// Name of the schema registry
func (o LookupSchemaRegistryResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.Name }).(pulumi.StringOutput)
}

// Owning role of the schema registry
func (o LookupSchemaRegistryResultOutput) Owner() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.Owner }).(pulumi.StringOutput)
}

// State of the schema registry
func (o LookupSchemaRegistryResultOutput) State() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.State }).(pulumi.StringOutput)
}

// Type of the schema registry
func (o LookupSchemaRegistryResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.Type }).(pulumi.StringOutput)
}

// Last update date of the schema registry
func (o LookupSchemaRegistryResultOutput) UpdatedAt() pulumi.StringOutput {
	return o.ApplyT(func(v LookupSchemaRegistryResult) string { return v.UpdatedAt }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupSchemaRegistryResultOutput{})
}
