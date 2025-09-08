# Contract: DeltaStreamObject Resource

## Inputs (Args)
- `database` (string, required)
- `namespace` (string, required)
- `store` (string, required)
- `sql` (string, required): SQL DDL statement.
- `owner` (string, optional)

## Outputs (State)
- All `DeltaStreamObjectArgs` fields
- `name` (string): Object name extracted from plan.
- `path` (string[]): [database, namespace, name]
- `fqn` (string): Fully qualified name.
- `type` (string): stream|changelog|table
- `state` (string): Provisioning state.
- `owner` (string)
- `createdAt` (string)
- `updatedAt` (string)

## Notes
- Object names must be unique within a namespace.
- SQL DDL is validated for correctness and immutability.
