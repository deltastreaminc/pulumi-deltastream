---
title: DeltaStream Installation & Configuration
meta_desc: Information on how to install the DeltaStream Pulumi provider and configure credentials.
layout: package
---

# DeltaStream Installation & Configuration

## Installation

The DeltaStream provider is available as a package in all Pulumi languages.

{{< chooser language "typescript,python,go,csharp" >}}

{{% choosable language typescript %}}

```bash
npm install @deltastream/pulumi-deltastream
```

{{% /choosable %}}

{{% choosable language python %}}

```bash
pip install pulumi-deltastream
```

{{% /choosable %}}

{{% choosable language go %}}

```bash
go get github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream
```

{{% /choosable %}}

{{% choosable language csharp %}}

```bash
dotnet add package Pulumi.DeltaStream
```

{{% /choosable %}}

{{< /chooser >}}

## Configuration

The DeltaStream provider requires the following configuration to connect to your DeltaStream deployment.

### Required

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `server` | `DELTASTREAM_SERVER` | The URL of your DeltaStream server (e.g. `https://api.deltastream.io/v2`) |

### Optional

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `apiKey` | `DELTASTREAM_API_KEY` | API key for authentication |
| `organization` | `DELTASTREAM_ORGANIZATION` | Default organization name |
| `role` | `DELTASTREAM_ROLE` | Default role to assume |
| `sessionId` | `DELTASTREAM_SESSION_ID` | Session ID for stateful connections |
| `insecureSkipVerify` | — | Skip TLS certificate verification (development only) |

### Setting configuration

Use `pulumi config set` to configure the provider:

```bash
pulumi config set deltastream:server https://api.deltastream.io/v2
pulumi config set --secret deltastream:apiKey <your-api-key>
pulumi config set deltastream:organization my-org
```

Or use environment variables:

```bash
export DELTASTREAM_SERVER=https://api.deltastream.io/v2
export DELTASTREAM_API_KEY=<your-api-key>
export DELTASTREAM_ORGANIZATION=my-org
```

### Provider block (explicit)

{{< chooser language "typescript,python,go,csharp" >}}

{{% choosable language typescript %}}

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";

const provider = new deltastream.Provider("deltastream", {
    server: "https://api.deltastream.io/v2",
    apiKey: process.env.DELTASTREAM_API_KEY,
    organization: "my-org",
});
```

{{% /choosable %}}

{{% choosable language python %}}

```python
import pulumi_deltastream as deltastream

provider = deltastream.Provider("deltastream",
    server="https://api.deltastream.io/v2",
    api_key=os.environ["DELTASTREAM_API_KEY"],
    organization="my-org",
)
```

{{% /choosable %}}

{{% choosable language go %}}

```go
provider, err := deltastream.NewProvider(ctx, "deltastream", &deltastream.ProviderArgs{
    Server:       pulumi.String("https://api.deltastream.io/v2"),
    ApiKey:       pulumi.String(os.Getenv("DELTASTREAM_API_KEY")),
    Organization: pulumi.String("my-org"),
})
```

{{% /choosable %}}

{{% choosable language csharp %}}

```csharp
var provider = new Pulumi.DeltaStream.Provider("deltastream", new()
{
    Server = "https://api.deltastream.io/v2",
    ApiKey = System.Environment.GetEnvironmentVariable("DELTASTREAM_API_KEY"),
    Organization = "my-org",
});
```

{{% /choosable %}}

{{< /chooser >}}
