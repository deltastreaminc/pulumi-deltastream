# Store Type Contracts

This document describes the contract for each supported store type in the Deltastream Pulumi provider. Each store type has specific required and optional properties, behaviors, and edge cases. This contract is derived from the codebase and integration tests.

---

## 1. Kafka Store

**Resource Type:** `deltastream:Store:Kafka`

### Inputs
- `name` (string, required): The name of the store.
- `type` (string, required): Must be `kafka`.
- `bootstrapServers` (string, required): Comma-separated list of Kafka brokers.
- `topic` (string, required): Kafka topic name.
- `properties` (map[string]string, optional): Additional Kafka client properties.

### Outputs
- `id` (string): Unique store identifier.
- `name` (string): Store name.
- `type` (string): Store type (`kafka`).
- `bootstrapServers` (string): Kafka brokers.
- `topic` (string): Topic name.
- `properties` (map[string]string): Additional properties.

### Notes
- The provider validates connectivity to the Kafka cluster on creation.
- If the topic does not exist, it is not created automatically; user must ensure it exists.
- Sensitive properties (e.g., credentials) should be passed via Pulumi secrets.

---

## 2. Postgres Store

**Resource Type:** `deltastream:Store:Postgres`

### Inputs
- `name` (string, required): The name of the store.
- `type` (string, required): Must be `postgres`.
- `host` (string, required): Hostname or IP of the Postgres server.
- `port` (int, optional, default: 5432): Port number.
- `database` (string, required): Database name.
- `user` (string, required): Username.
- `password` (string, required, secret): Password.
- `sslMode` (string, optional): SSL mode (e.g., `disable`, `require`).
- `properties` (map[string]string, optional): Additional connection properties.

### Outputs
- `id` (string): Unique store identifier.
- `name` (string): Store name.
- `type` (string): Store type (`postgres`).
- `host` (string): Hostname.
- `port` (int): Port.
- `database` (string): Database name.
- `user` (string): Username.
- `sslMode` (string): SSL mode.
- `properties` (map[string]string): Additional properties.

### Notes
- The provider validates connectivity to the Postgres server on creation.
- Passwords are always handled as Pulumi secrets.
- If the database does not exist, creation fails.

---

## 3. Snowflake Store

**Resource Type:** `deltastream:Store:Snowflake`

### Inputs
- `name` (string, required): The name of the store.
- `type` (string, required): Must be `snowflake`.
- `account` (string, required): Snowflake account identifier.
- `user` (string, required): Username.
- `password` (string, required, secret): Password.
- `database` (string, required): Database name.
- `warehouse` (string, required): Warehouse name.
- `role` (string, optional): Role name.
- `properties` (map[string]string, optional): Additional connection properties.

### Outputs
- `id` (string): Unique store identifier.
- `name` (string): Store name.
- `type` (string): Store type (`snowflake`).
- `account` (string): Account identifier.
- `user` (string): Username.
- `database` (string): Database name.
- `warehouse` (string): Warehouse name.
- `role` (string): Role name.
- `properties` (map[string]string): Additional properties.

### Notes
- The provider validates connectivity to the Snowflake account on creation.
- Passwords are always handled as Pulumi secrets.
- If the database or warehouse does not exist, creation fails.

---

## 4. Memory Store (for testing)

**Resource Type:** `deltastream:Store:Memory`

### Inputs
- `name` (string, required): The name of the store.
- `type` (string, required): Must be `memory`.

### Outputs
- `id` (string): Unique store identifier.
- `name` (string): Store name.
- `type` (string): Store type (`memory`).

### Notes
- Only available in test/in-memory mode.
- Not for production use.

---

## General Notes
- All store types require a unique `name` within the namespace.
- The `type` field is used to select the store implementation and must match one of the supported types: `kafka`, `postgres`, `snowflake`, `memory`.
- Additional store types may be added in the future; contracts must be updated accordingly.
- All sensitive fields (e.g., passwords) must be marked as Pulumi secrets in resource definitions.

---

## References
- See `provider/store.go` for implementation details.
- See `examples/` for integration test coverage of each store type.
