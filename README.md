# Pulumi DeltaStream Provider

A Pulumi provider for managing DeltaStream resources.

## Overview

The DeltaStream provider for Pulumi allows you to manage DeltaStream resources using infrastructure as code. DeltaStream is a streaming data platform that enables real-time analytics and processing.
This provider supports Databases, Namespaces, Stores, Objects (STREAM/CHANGELOG/TABLE) and now Continuous Queries (INSERT INTO ... SELECT ...).

## Installation

### TypeScript/JavaScript (Node.js)

```bash
npm install @deltastream/pulumi-deltastream
```

### Go

```bash
go get github.com/deltastreaminc/pulumi-deltastream/sdk/go/deltastream
```

## Configuration

Provider configuration mirrors environment variables used by the underlying DeltaStream SQL driver.

| Pulumi Config Key | Environment Variable | Description | Required |
|-------------------|----------------------|-------------|----------|
| `server` | `DELTASTREAM_SERVER` | Base server URL (e.g. https://api.deltastream.io/v2) | Yes |
| `apiKey` | `DELTASTREAM_API_KEY` | API key/token for authentication | Yes (unless supplied via env) |
| `organization` | `DELTASTREAM_ORGANIZATION` | Organization name or UUID | No |
| `role` | `DELTASTREAM_ROLE` | Role to execute statements as (defaults server-side) | No |
| `insecureSkipVerify` | `DELTASTREAM_INSECURE_SKIP_VERIFY` | Skip TLS verification (dev/testing) | No |
| `sessionId` | `DELTASTREAM_SESSION_ID` | Custom session ID (helps correlate logs) | No |

Example (environment variables):

```bash
export DELTASTREAM_SERVER="https://api.deltastream.io/v2"
export DELTASTREAM_API_KEY="<your key>"
```

Or Pulumi config (secrets recommended):

```bash
pulumi config set deltastream:server https://api.deltastream.io/v2
pulumi config set --secret deltastream:apiKey <your key>
```

## Example Usage

### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";

const provider = new deltastream.Provider("deltastream", {
    server: process.env.DELTASTREAM_SERVER!,
    apiKey: process.env.DELTASTREAM_API_KEY!,
});

// Create a database and a namespace
const db = new deltastream.Database("example_db", { name: "example_db" }, { provider });
const ns = new deltastream.Namespace("example_ns", { database: db.name, name: "example_ns" }, { provider });

// Invoke lookups
const dbInfo = db.name.apply(n => deltastream.getDatabase({ name: n }, { provider }));
const namespaces = db.name.apply(d => deltastream.getNamespaces({ database: d }, { provider }));

export const dbCreatedAt = db.createdAt;
export const nsCreatedAt = ns.createdAt;
export const namespaceCount = namespaces.apply(r => r.namespaces.length);
```

### Go

```go
package main

import (
    ds "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "os"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{
            Server: pulumi.String(os.Getenv("DELTASTREAM_SERVER")),
            ApiKey: pulumi.String(os.Getenv("DELTASTREAM_API_KEY")),
        })
        if err != nil { return err }

        db, err := ds.NewDatabase(ctx, "db", &ds.DatabaseArgs{ Name: pulumi.String("example_db") }, pulumi.Provider(prov))
        if err != nil { return err }

        ns, err := ds.NewNamespace(ctx, "ns", &ds.NamespaceArgs{ Database: db.Name, Name: pulumi.String("example_ns") }, pulumi.Provider(prov))
        if err != nil { return err }

        ctx.Export("dbCreatedAt", db.CreatedAt)
        ctx.Export("nsCreatedAt", ns.CreatedAt)
        return nil
    })
}
```

### Streaming Query (Go)

```go
q, err := ds.NewQuery(ctx, "insertExample", &ds.QueryArgs{
    SourceRelationFqns: pulumi.StringArray{ source.Fqn },
    SinkRelationFqn:    sink.Fqn,
    Sql: pulumi.Sprintf("INSERT INTO %s SELECT * FROM %s;", sink.Fqn, source.Fqn),
}, pulumi.Provider(prov))
if err != nil { return err }
_ = q
```

## Development

### Prerequisites

- Go 1.24+
- Pulumi CLI (installed separately; CI uses a pinned version via `pulumi/actions`)
- Node.js (only if working on / validating the Node SDK)
- Optional: Yarn for faster Node builds

The Pulumi CLI is no longer auto-installed by the Makefile (for supply‑chain safety). Install it manually or via your package manager:

```bash
curl -fsSL https://get.pulumi.com | sh   # (optional quick start; prefer package managers or pinned action in CI)
# or on macOS
brew install pulumi
```

The repository pins a CLI version in CI using the environment variable `PULUMI_CLI_VERSION` (see `.github/workflows/ci.yml`). If you encounter schema generation differences, check your local `pulumi version` and align it with that value.

### Building the Provider

```bash
make build
```

### Running Tests

```bash
make test
```

### Generating SDKs

```bash
make build_sdks
```

### Available Make Targets

- `build` – Build the provider binary
- `schema` – Generate provider schema (requires Pulumi CLI present)
- `generate` – Generate all language SDKs (nodejs, go, python)
- `build_sdks` – Build all SDKs (sanity compile)
- `install_sdks` – Install/link SDKs locally for development
- `test` – Run example integration tests (requires built provider & Pulumi CLI)
- `clean` – Remove build artifacts & generated SDKs
- `help` – Show help message

Targets intentionally avoid implicitly downloading tools; ensure the Pulumi CLI is on your PATH before running schema or generate targets.

## Project Structure

```
.
├── cmd/
│   └── pulumi-resource-deltastream/    # Provider binary entry point
├── provider/                           # Provider implementation
├── sdk/                               # Generated SDKs
│   ├── go/
│   └── nodejs/
├── examples/                          # Example programs
├── schema.json                        # Provider schema
├── Makefile                          # Build automation
└── README.md                         # This file
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:

- [GitHub Issues](https://github.com/deltastreaminc/pulumi-deltastream/issues)
- [DeltaStream Documentation](https://docs.deltastream.io)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
