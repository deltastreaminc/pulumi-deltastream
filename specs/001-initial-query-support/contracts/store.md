# Contract: Store Resource

## Inputs (Args)
- `name` (string, required)
- `owner` (string, optional)
- `kafka` (KafkaInputs, optional)
- `snowflake` (SnowflakeInputs, optional)
- `postgres` (PostgresInputs, optional)

## Outputs (State)
- All `StoreArgs` fields
- `type` (string): Type of the store.
- `state` (string): Provisioning state.
- `createdAt` (string)
- `updatedAt` (string)
- `owner` (string): Owner as returned from system.

## Notes
- Exactly one of `kafka`, `snowflake`, or `postgres` must be set.
- Subtype inputs are validated for required fields.
