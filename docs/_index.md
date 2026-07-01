---
title: DeltaStream
meta_desc: Provides an overview of the DeltaStream Pulumi provider, including installation, configuration, and examples.
layout: package
---

# DeltaStream Provider

The DeltaStream Pulumi provider enables you to manage [DeltaStream](https://www.deltastream.io) resources as infrastructure — databases, namespaces, stores, streams, changelogs, tables, queries, and applications — using any Pulumi-supported language.

## Example

{{< chooser language "typescript,python,go,csharp" >}}

{{% choosable language typescript %}}

```typescript
import * as deltastream from "@deltastream/pulumi-deltastream";

const db = new deltastream.Database("example", {
    name: "example_db",
});

const ns = new deltastream.Namespace("example", {
    name: "example_ns",
    database: db.name,
});
```

{{% /choosable %}}

{{% choosable language python %}}

```python
import pulumi_deltastream as deltastream

db = deltastream.Database("example", name="example_db")

ns = deltastream.Namespace("example",
    name="example_ns",
    database=db.name,
)
```

{{% /choosable %}}

{{% choosable language go %}}

```go
import (
    "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        db, err := deltastream.NewDatabase(ctx, "example", &deltastream.DatabaseArgs{
            Name: pulumi.String("example_db"),
        })
        if err != nil {
            return err
        }
        _, err = deltastream.NewNamespace(ctx, "example", &deltastream.NamespaceArgs{
            Name:     pulumi.String("example_ns"),
            Database: db.Name,
        })
        return err
    })
}
```

{{% /choosable %}}

{{% choosable language csharp %}}

```csharp
using Pulumi;
using Pulumi.DeltaStream;

return await Deployment.RunAsync(() =>
{
    var db = new Database("example", new DatabaseArgs { Name = "example_db" });
    var ns = new Namespace("example", new NamespaceArgs
    {
        Name = "example_ns",
        Database = db.Name,
    });
});
```

{{% /choosable %}}

{{< /chooser >}}

## Resources

- **Database** — A DeltaStream database (top-level namespace container)
- **Namespace** — A namespace within a database
- **Store** — An external data source or sink (Kafka, Snowflake, PostgreSQL)
- **DeltaStreamObject** — A stream, changelog, or table DDL object
- **Query** — A continuous INSERT INTO streaming query
- **Application** — A multi-sink streaming application

## Functions

- `getDatabase` / `getDatabases` — look up databases by name or list all
- `getNamespace` / `getNamespaces` — look up namespaces or list all
- `getStore` / `getStores` — look up stores or list all
- `getObject` / `getObjects` — look up DeltaStream objects or list all
