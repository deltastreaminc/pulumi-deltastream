# Contract: Namespace Resource

## Inputs (Args)
- `database` (string, required): Name of the database containing the namespace.
- `name` (string, required): Name of the namespace.
- `owner` (string, optional): Owning role.

## Outputs (State)
- `database` (string)
- `name` (string)
- `owner` (string)
- `createdAt` (string): Creation timestamp.

## Notes
- Namespace must be unique within a database.
- Owner is optional and may be set at creation or inherited from provider config.
