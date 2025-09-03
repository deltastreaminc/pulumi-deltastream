---
mode: 'agent'
model: GPT-5 (Preview)
tools: ['edit', 'notebooks', 'search', 'new', 'runCommands', 'runTasks', 'usages', 'vscodeAPI', 'think', 'problems', 'changes', 'testFailure', 'openSimpleBrowser', 'fetch', 'findTestFiles', 'githubRepo', 'extensions', 'todos', 'runTests', 'godoc', 'github-mcp']
description: 'Add a new store types'
---

Your goal is to add support for a new ${input:type:store type} store based on the patterns established in the existing `store.go` and `store_kafka.go` files. Follow the detailed design patterns and steps outlined below to ensure consistency, maintainability, and correctness.

## Summary
- Provider Language: Go (Pulumi Infer framework).
- Readiness: After Create/Update, implementation polls `deltastream.sys."stores"` until `status` is `ready` or `errored` (with `status_message`). Timeout currently set to 1 minute (adjustable if needed).
- Delete: Issues `DROP STORE "<name>";` and performs best-effort disappearance verification.
- Update: Implemented via `UPDATE STORE ... WITH (...)` using diff-based parameter map.
- Schema & SDK Generation: Automated via `make schema` and `make generate_go|_nodejs`.
- Store code organization:
  - `provider/store.go`: Generic resource orchestration (Check/Create/Read/Update/Delete/Diff/WireDependencies) delegating to type-specific helpers.
  - `provider/store_<TypeName>.go`: Store-specific validation, create, update, diff logic.
  - Naming Convention Fix: Inputs struct renamed from `StoreInputs` to `StoreArgs` to ensure correct schema emission.
- Example Updated: `examples/<TypeName>-store-go/stepNN/main.go`
    * step 1 - create a store
    * step 2 - update the store
- Create store documentation: https://docs.deltastream.io/reference/sql-syntax/ddl/create-store
- Update store documentation: https://docs.deltastream.io/reference/sql-syntax/ddl/update-store
- Reference Terraform provider with support for some stores: github.com/deltastreaminc/terraform-provider-deltastream (public repo, fetch using github mcp)

## Key Design Patterns to Reuse
1. Separation of Concerns:
   - Keep `store.go` focused on lifecycle plumbing and shared logic (readiness polling, lookup, delete verification).
   - Add a new file per store type: `store_<lowercase>.go` (e.g., `store_kinesis.go`).
2. Type-Specific Structures:
   - Define `<TypeName>Inputs` struct with Pulumi tags for its properties.
   - Extend `StoreArgs` by adding an optional pointer field for the new store type (e.g., `Kinesis *KinesisInputs `pulumi:"kinesis,optional"``).
3. Validation:
   - Implement `<storeType>Check` logic appended into the main `Check` dispatch (currently only calls Kafka). Refactor `Store.Check` to detect which block is set and route accordingly.
   - Enforce mutual exclusivity: only one store subtype block should be set per resource instance.
4. Create Logic:
   - Follow pattern in `storeKafkaCreate`: build a parameter map, escape single quotes, embed file contents if needed, and execute `CREATE STORE`.
5. Update Logic:
   - Mirror `storeKafkaUpdate` diff approach: compute changed keys, translate deletions to `NULL`, avoid unnecessary statements.
   - Only issue UPDATE if changes map non-empty.
6. Diff Logic:
   - Add subtype-aware diff (`store<Type>Diff`) returning `p.PropertyDiff` entries for mutated fields.
   - Wire into `Store.Diff` dispatch (similar to Check/Update routing refactor).
7. Readiness & Error Handling:
   - Reuse `waitForStoreReady` unchanged (works generically off system table columns).
8. Schema Exposure:
   - Pulumi Infer auto-generates schema: ensure exported struct & fields with correct `pulumi:""` tags.
   - Maintain naming consistency: `StoreArgs`, `<TypeName>Inputs`.
9. Backward Compatibility:
   - Only additive changes to `StoreArgs` and schema.
10. Tests / Examples:
   - Add or extend examples conditionally: skip real external dependencies if env vars absent.
   - Use environment-driven gating for integration examples (mirrors Kafka approach).
   - test store credentials will be loaded from credentials.yaml file (maintained by user, do not check into git). Update credentials.yaml.example for example that is checked into git.

## Step Template for Adding a New Store Type
For a store type (example: Kinesis):
1. Planning / Commit 1: Introduce inputs struct & augment `StoreArgs`.
   - File: `provider/store_kinesis.go` with `KinesisInputs` and stubbed helper functions.
   - Update `store.go` Check to detect and call `storeKinesisCheck`.
2. Commit 2: Implement Create helper (`storeKinesisCreate`).
3. Commit 3: Implement Update helper (`storeKinesisUpdate`).
4. Commit 4: Implement Diff helper (`storeKinesisDiff`) and wire into dispatch.
5. Commit 5: Add readiness verification uses existing polling (no change, but ensure create/update paths invoke it).
6. Commit 6: Extend example(s) with gated usage (skip if required env vars for Kinesis not set).
7. Commit 7: Regenerate schema & SDKs (`make clean ; make build schema generate build_sdks install_sdks`).
8. Commit 8: Add docs / README excerpt for new store parameters.
9. Optional: Add minimal unit tests for validation logic (if feasible) else rely on integration pattern.

### Integration Test Pattern (Example: Kafka Auth Mode Update)
Use Pulumi integration `EditDirs` to simulate an update:
1. `examples/<store>-store-go/step1` – initial creation (e.g., SCRAM auth).
2. `examples/<store>-store-go/step2` – modified inputs triggering in-place UPDATE (e.g., switch to IAM auth).
3. Test (`examples/go_test.go`) sets `Dir` to step1 and an `EditDir` for step2 with `Additive: true`.
4. Export distinguishing outputs (e.g., `store_auth_mode`) in both steps and assert they change post-update.
5. Gate test with required env vars; skip if absent to avoid flaky CI.

Replicate this pattern for new store types when meaningful (e.g., changing credentials, toggling TLS, switching mode). Keep examples minimal and idempotent.

## Dispatch Refactor (Multi-Type Support)
When adding second store type, refactor these methods in `store.go` to route by which subtype pointer is non-nil:
- Check: ensure exactly one subtype block set; call subtype-specific check.
- Create: switch on non-nil subtype, call appropriate create helper.
- Update & Diff: same pattern; if subtype changed (should be disallowed) trigger replace.

Pseudo-code sketch for future multi-type dispatch:
```go
func (Store) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[StoreArgs], error) {
  // run DefaultCheck first
  // count non-nil subtype blocks
  // route to specific subtype check function (returns mutated args & failures)
}

func (Store) Create(...) { switch {
  case input.Kafka != nil: storeKafkaCreate(...)
  case input.Kinesis != nil: storeKinesisCreate(...)
  default: error
}}
```

## Read & State Model
Current Read logic only refreshes generic columns (type, state, timestamps, owner). If future store types require surfacing computed subtype fields, extend state model (e.g., embed a read-only representation) and add corresponding columns/queries if system table supports them.

## Escaping & Safety
- Always escape single quotes in dynamic SQL values: `strings.ReplaceAll(val, "'", "''")`.
- For boolean flags map directly to `TRUE`/`FALSE` (uppercase) to avoid quoting.
- For secrets (passwords, keys) treat them as strings; Pulumi secret support can be added later by marking fields with `pulumi:"...,secret"` if desired.

## Readiness Timing
- Current timeout is 1 minute (`waitForStoreReady`). If some store types provision longer, introduce a per-type or configurable timeout (e.g., via provider config) in a later enhancement. Document rationale before changing.

## Deletion Semantics
- Generic `DROP STORE` works for all types; retain disappearance loop.
- If a type needs dependency draining, implement a pre-drop wait in its helper (optional future enhancement).

## Adding Validation Nuances
Examples:
- Kinesis: Region + stream/account identifiers; mutually exclusive auth modes.
- Snowflake: Account, warehouse, role, database parameters; optional external STS assumed role.
Add explicit failure entries via `p.CheckFailure{Property: "storeType.field", Reason: "..."}`.

## Commands Reference
Rebuild provider & regenerate schema / SDKs:
```bash
make provider
make schema
make generate_go generate_nodejs
```
Run examples tests (Go):
```bash
make test
```
Clean workspace:
```bash
make clean
```

## Acceptance Checklist Before Committing a New Store Type
- [ ] Inputs struct added with clear pulumi tags & required fields.
- [ ] `StoreArgs` updated with new optional pointer field.
- [ ] Validation enforces required vs. mutually exclusive fields.
- [ ] Create logic builds accurate parameter map and escapes values.
- [ ] Update logic issues `UPDATE STORE` only when needed.
- [ ] Diff logic enumerates changed properties correctly.
- [ ] Readiness polling invoked after create/update.
- [ ] Example updated (gated by env vars) without breaking existing ones.
- [ ] Schema + SDKs regenerated; new fields visible in generated code.
- [ ] Builds succeed: `go build ./...`.
- [ ] Optional: Docs / README updated.

## Future Enhancements (Backlog)
- Secret field tagging for credentials.
- Configurable readiness timeout.
- Structured status export (e.g., expose `status_message` when errored).
- Multi-language example parity (NodeJS, Python) for store usage.
- Unit tests for diff/validation logic.
