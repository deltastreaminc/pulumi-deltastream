import * as pulumi from "@pulumi/pulumi";
// Import from the published package name.
import * as deltastream from "@deltastream/pulumi-deltastream";

// Configure the DeltaStream provider using the new keys
if (!process.env.DELTASTREAM_SERVER) {
    throw new Error("DELTASTREAM_SERVER is required");
}
if (!process.env.DELTASTREAM_API_KEY) {
    throw new Error("DELTASTREAM_API_KEY is required");
}
const provider = new deltastream.Provider("deltastream", {
    server: process.env.DELTASTREAM_SERVER,
    apiKey: process.env.DELTASTREAM_API_KEY,
    organization: process.env.DELTASTREAM_ORGANIZATION, // optional (ID or name)
    role: process.env.DELTASTREAM_ROLE, // optional
    insecureSkipVerify: process.env.DELTASTREAM_INSECURE_SKIP_VERIFY === "true", // optional
    sessionId: process.env.DELTASTREAM_SESSION_ID  // optional
});

// Create a database
const db = new deltastream.Database(
    "example_db",
    {
        name: "example_db",
    },
    { provider }
);

// Create a namespace within the database
const ns = new deltastream.Namespace(
    "example_namespace",
    {
        database: db.name,
        name: "example_namespace_ts",
    },
    { provider }
);

// Invoke: get a single database
const dbInfo = pulumi.output(db.name).apply(async (n:string) => {
  if (pulumi.runtime.isDryRun()) {
    return { name: n, owner: "", createdAt: "" }; // placeholder
  }
  return await deltastream.getDatabase({ name: n }, { provider });
});

// Invoke: list databases
const dbs = pulumi.output(
    deltastream.getDatabases({}, { provider })
);

export const createdAt = db.createdAt;
export const namespaceCreatedAt = ns.createdAt;
export const namespaceOwner = pulumi.output(ns.owner);
export const lookedUpOwner = dbInfo.owner;
export const databaseCount = dbs.apply((r: any) => (r?.databases?.length ?? r?.items?.length ?? 0));

// Namespace invokes (skip during preview)
export const namespacesCount = pulumi.output(db.name).apply(async (d: string) => {
    if (pulumi.runtime.isDryRun()) { return 0; }
    const list = await deltastream.getNamespaces({ database: d }, { provider });
    return list.namespaces.length;
});