# Pulumi DeltaStream Provider

A Pulumi provider for managing DeltaStream resources.

## Overview

The DeltaStream provider for Pulumi allows you to manage DeltaStream resources using infrastructure as code. DeltaStream is a streaming data platform that enables real-time analytics and processing.

This provider supports:
- **Databases** - Logical containers for schemas and relations
- **Namespaces** - Schema namespaces within databases
- **Stores** - External data stores (Kafka, Kinesis, etc.)
- **Objects** - Relations (STREAM/CHANGELOG/TABLE)
- **Query** - Continuous INSERT INTO queries (single sink)
- **Application** - Multi-sink streaming applications with virtual relations

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

### Streaming Queries

DeltaStream supports two types of streaming query resources:

#### 1. Query Resource (INSERT INTO)

The `Query` resource is for simple INSERT INTO queries with a single sink:

**TypeScript:**
```typescript
const query = new deltastream.Query("myQuery", {
    sourceRelationFqns: [source.fqn],
    sinkRelationFqn: sink.fqn,  // Single sink
    sql: pulumi.interpolate`INSERT INTO ${sink.fqn} SELECT * FROM ${source.fqn};`,
}, { provider });
```

**Go:**
```go
q, err := ds.NewQuery(ctx, "myQuery", &ds.QueryArgs{
    SourceRelationFqns: pulumi.StringArray{ source.Fqn },
    SinkRelationFqn:    sink.Fqn,  // Single sink
    Sql: pulumi.Sprintf("INSERT INTO %s SELECT * FROM %s;", sink.Fqn, source.Fqn),
}, pulumi.Provider(prov))
```

#### 2. Application Resource (Multi-Sink)

The `Application` resource is for complex streaming applications with:
- **Multiple INSERT INTO statements** targeting different sinks
- **Virtual intermediate relations** (CREATE VIRTUAL STREAM/CHANGELOG)
- **Complex processing logic** with joins, windows, and aggregations

**Key Features:**
- Virtual relations are internal to the APPLICATION and don't create Kafka topics
- Only physical sources and sinks need to be declared as dependencies
- Virtual relations are automatically excluded from dependency tracking

**TypeScript:**
```typescript
// 1. Create physical source and sink relations OUTSIDE the application
const pageviews = new deltastream.DeltaStreamObject("pageviews", {
    database: db.name,
    namespace: "public",
    store: kafkaStore.name,
    sql: pulumi.interpolate`CREATE STREAM pageviews (...) WITH ('topic'='pageviews');`,
}, { provider });

const visitFreq = new deltastream.DeltaStreamObject("visitFreq", {
    database: db.name,
    namespace: "public",
    store: kafkaStore.name,
    sql: pulumi.interpolate`CREATE CHANGELOG visit_freq (...) WITH ('topic'='visit_freq');`,
}, { provider });

// 2. Create APPLICATION with virtual relations and INSERT INTO
const app = new deltastream.Application("myApp", {
    sourceRelationFqns: [pageviews.fqn],      // Physical sources only
    sinkRelationFqns: [visitFreq.fqn],        // Physical sinks only
    sql: pulumi.interpolate`
        BEGIN APPLICATION my_app
            -- Virtual relation (no Kafka topic, internal only)
            CREATE VIRTUAL STREAM virtual.public.filtered AS
                SELECT * FROM ${pageviews.fqn}
                WHERE userid IS NOT NULL;
            
            -- Insert into physical sink
            INSERT INTO ${visitFreq.fqn}
                SELECT window_start, window_end, userid, count(*) as cnt
                FROM TUMBLE(virtual.public.filtered, SIZE 30 SECONDS)
                GROUP BY window_start, window_end, userid;
        END APPLICATION;
    `,
}, { provider, dependsOn: [pageviews, visitFreq] });
```

**Go:**
```go
// 1. Create physical source and sink relations OUTSIDE the application
pageviews, err := ds.NewDeltaStreamObject(ctx, "pageviews", &ds.DeltaStreamObjectArgs{
    Database:  db.Name,
    Namespace: pulumi.String("public"),
    Store:     kafkaStore.Name,
    Sql: pulumi.Sprintf("CREATE STREAM pageviews (...) WITH ('topic'='pageviews');"),
}, pulumi.Provider(prov))

visitFreq, err := ds.NewDeltaStreamObject(ctx, "visitFreq", &ds.DeltaStreamObjectArgs{
    Database:  db.Name,
    Namespace: pulumi.String("public"),
    Store:     kafkaStore.Name,
    Sql: pulumi.Sprintf("CREATE CHANGELOG visit_freq (...) WITH ('topic'='visit_freq');"),
}, pulumi.Provider(prov))

// 2. Create APPLICATION with virtual relations and INSERT INTO
app, err := ds.NewApplication(ctx, "myApp", &ds.ApplicationArgs{
    SourceRelationFqns: pulumi.StringArray{ pageviews.Fqn },  // Physical sources only
    SinkRelationFqns:   pulumi.StringArray{ visitFreq.Fqn },  // Physical sinks only
    Sql: pulumi.Sprintf(`
        BEGIN APPLICATION my_app
            -- Virtual relation (no Kafka topic, internal only)
            CREATE VIRTUAL STREAM virtual.public.filtered AS
                SELECT * FROM %s
                WHERE userid IS NOT NULL;
            
            -- Insert into physical sink
            INSERT INTO %s
                SELECT window_start, window_end, userid, count(*) as cnt
                FROM TUMBLE(virtual.public.filtered, SIZE 30 SECONDS)
                GROUP BY window_start, window_end, userid;
        END APPLICATION;
    `, pageviews.Fqn, visitFreq.Fqn),
}, pulumi.Provider(prov))
```

**Important Notes:**
- Virtual relations (CREATE VIRTUAL) must NOT be included in `sourceRelationFqns` or `sinkRelationFqns`
- Only physical relations (with actual Kafka topics) should be declared as dependencies
- The provider validates that virtual relations are not incorrectly declared as dependencies

For complete examples, see:
- [examples/application-go](examples/application-go/) - Full Go example with multiple sinks
- [examples/application-ts](examples/application-ts/) - Full TypeScript example with multiple sinks

## Development

### Resources

The provider includes the following resources:

| Resource | Description | Use Case |
|----------|-------------|----------|
| `Database` | Logical database container | Group related schemas and relations |
| `Namespace` | Schema namespace within a database | Organize relations |
| `Store` | External data store connection (Kafka, Kinesis, etc.) | Connect to data sources |
| `DeltaStreamObject` | Physical relation (STREAM/CHANGELOG/TABLE) | Create physical data structures with Kafka topics |
| `Query` | Continuous INSERT INTO query | Simple single-sink streaming transformations |
| `Application` | Multi-sink streaming application | Complex applications with multiple sinks and virtual relations |

### Query vs Application: When to Use Each

**Use `Query` when:**
- You have a simple INSERT INTO ... SELECT query
- Single source and single sink
- No virtual intermediate relations needed
- Straightforward transformations

**Use `Application` when:**
- Multiple INSERT INTO statements targeting different sinks
- Need virtual intermediate relations (CREATE VIRTUAL STREAM/CHANGELOG)
- Complex processing with joins, windows, and aggregations
- Want to organize related queries into a single logical unit

### Query Resource Field Notes

The Query resource supports both legacy and current field names for backward compatibility:

- **`sinkRelationFqn`** (string, **deprecated**): For backward compatibility with single-sink INSERT INTO queries
- **`sinkRelationFqns`** (string[], **deprecated**): For legacy APPLICATION queries

**Recommendation:** For new code:
- Use `Query` resource for simple INSERT INTO queries
- Use `Application` resource for multi-sink streaming applications

**Why the change?** The new `Application` resource provides:
- Type-safe APPLICATION-specific validation
- Clear separation between Query and Application concerns  
- Better developer experience with dedicated fields
- Automatic validation of virtual relation dependencies

**Migration:** Existing Query resources continue to work. See [CHANGELOG.md](CHANGELOG.md) for migration guide.

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
│   ├── application-go/               # Go APPLICATION example (multi-sink)
│   ├── application-ts/               # TypeScript APPLICATION example (multi-sink)
│   ├── query-go/                     # Go Query example (single sink)
│   └── query-ts/                     # TypeScript Query example (single sink)
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