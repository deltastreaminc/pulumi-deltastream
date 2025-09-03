module example-query

go 1.24.5

require (
	github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream v1.0.0-alpha.0+dev
	github.com/pulumi/pulumi/sdk/v3 v3.193.0
)

replace github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream => ../../sdk/go/pulumi-deltastream
