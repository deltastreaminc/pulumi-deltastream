# Contract: Deltastream System Tables

## Purpose
System tables are used by the provider to poll and verify the status, existence, and metadata of Deltastream resources. All CRUD and invoke operations rely on these tables for readiness, idempotency, and accurate state reflection.

## Key System Tables Used
- `deltastream.sys.databases`
- `deltastream.sys.schemas` (namespaces)
- `deltastream.sys.stores`
- `deltastream.sys.relations` (objects)
- `deltastream.sys.queries`

## Example Schema (Fields may vary by table)

### deltastream.sys.databases
- `name` (string): Database name
- `owner` (string): Owning role
- `created_at` (timestamp): Creation time

### deltastream.sys.schemas
- `name` (string): Namespace name
- `database_name` (string): Parent database
- `owner` (string): Owning role
- `created_at` (timestamp)

### deltastream.sys.stores
- `name` (string): Store name
- `type` (string): Store type (kafka, postgres, snowflake)
- `state` (string): Provisioning state
- `owner` (string): Owning role
- `created_at` (timestamp)
- `updated_at` (timestamp)

### deltastream.sys.relations
- `name` (string): Object name
- `database_name` (string)
- `schema_name` (string)
- `fqn` (string): Fully qualified name
- `relation_type` (string): stream, changelog, table
- `state` (string): Provisioning state
- `owner` (string)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### deltastream.sys.queries
- `name` (string): Query name
- `version` (int): Query version
- `current_state` (string): Lifecycle state
- `owner` (string)
- `created_at` (timestamp)
- `updated_at` (timestamp)

## Usage Patterns
- All resource CRUD and invoke operations perform SELECTs against these tables to verify existence, readiness, and to fetch metadata.
- Status polling for readiness uses these tables to ensure resources are fully provisioned before reporting success.
- Deletes are idempotent: if a resource is not found in the system table, the operation is considered successful.
- All lookups and filters are performed using quoted identifiers to prevent SQL injection.

## Notes
- The exact schema and available fields may evolve with the Deltastream platform.
- The provider must handle missing or unexpected fields gracefully and surface backend errors to the user.
