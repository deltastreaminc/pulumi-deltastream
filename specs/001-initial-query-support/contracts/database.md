# Contract: Database Resource

## Inputs (Args)
- `name` (string, required): Name of the database.
- `owner` (string, optional): Owning role; overrides provider role for creation.

## Outputs (State)
- `name` (string)
- `owner` (string)
- `createdAt` (string): Timestamp when the database was created.

## Notes
- All fields are validated for presence and type.
- Owner is optional and may be set at creation or inherited from provider config.
