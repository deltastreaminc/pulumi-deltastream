# Contract: Invoke Functions (Read-only)

## getDatabase
- **Input:**
  - `name` (string, required)
- **Output:**
  - `name` (string)
  - `owner` (string)
  - `createdAt` (string)

## getDatabases
- **Input:**
  - (none)
- **Output:**
  - `databases` (array of DatabaseResult)

## getNamespace
- **Input:**
  - `database` (string, required)
  - `name` (string, required)
- **Output:**
  - `database` (string)
  - `name` (string)
  - `owner` (string)
  - `createdAt` (string)

## getNamespaces
- **Input:**
  - `database` (string, required)
- **Output:**
  - `namespaces` (array of NamespaceResult)

## getStore
- **Input:**
  - `name` (string, required)
- **Output:**
  - `name` (string)
  - `type` (string)
  - `state` (string)
  - `owner` (string)
  - `createdAt` (string)
  - `updatedAt` (string)

## getStores
- **Input:**
  - (none)
- **Output:**
  - `stores` (array of StoreResult)

## getObject
- **Input:**
  - `database` (string, required)
  - `namespace` (string, required)
  - `name` (string, required)
- **Output:**
  - `name` (string)
  - `fqn` (string)
  - `type` (string)
  - `state` (string)
  - `owner` (string)
  - `createdAt` (string)
  - `updatedAt` (string)

## getObjects
- **Input:**
  - `database` (string, required)
  - `namespace` (string, required)
- **Output:**
  - `objects` (array of ObjectResult)

## Notes
- All invoke functions are read-only and do not create or modify resources.
- Singular lookups return a not-found error if no match, and an error if multiple matches are found.
- Plural lookups are capped at 200 results and return a warning if truncated.
