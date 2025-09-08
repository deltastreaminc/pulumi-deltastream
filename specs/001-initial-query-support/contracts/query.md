# Contract: Query Resource

## Inputs (Args)
- `sourceRelationFqns` (string[], required)
- `sinkRelationFqn` (string, required)
- `sql` (string, required)
- `owner` (string, optional)

## Outputs (State)
- All `QueryArgs` fields
- `queryId` (string)
- `queryName` (string, optional)
- `queryVersion` (int64, optional)
- `state` (string): Lifecycle state.
- `createdAt` (string)
- `updatedAt` (string)
- `owner` (string)

## Notes
- Query body is immutable after creation.
- Attempts to change SQL after creation result in a plan error.
