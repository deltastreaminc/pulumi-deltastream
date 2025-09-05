 # Feature Specification: Store Resource (Kafka, Postgres, Snowflake)

**Feature Branch**: `feat/store-resource`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: External system integration requirement

---
## ⚡ Quick Guidelines
Unified abstraction for external/managed storage endpoints with uniform readiness polling (system table), secret-marked credentials, minimal required args, additive evolution only.

---
## 1. Summary *(mandatory)*
Pulumi `Store` resource manages creation and lifecycle of heterogeneous storage backends (Kafka, Postgres, Snowflake) with subtype validation and readiness tracking.

## 2. Goals *(mandatory)*
- Provide single resource interface across store technologies.
- Expose readiness state & timestamps for orchestration.
- Allow limited in-place updates of mutable metadata.

## 3. Non-Goals *(mandatory)*
- Cross-store data migration flows.
- Secret rotation management.
- Automatic scaling policy management.

## 4. Background & Context *(mandatory)*
Multiple backend types require uniform provisioning semantics to support objects & queries. Manual scripts are inconsistent and error-prone.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| Data Platform Engineer | Consistent onboarding of external stores | Divergent scripts |
| Security Auditor | Centralized visibility into store config | Opaque manual steps |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As a platform engineer I want to declare an external store so that downstream objects can reliably connect once it is ready.

### Acceptance Scenarios
1. Given a new store declaration, when apply runs, then the store is created and readiness polling completes before success.
2. Given a store is still provisioning, when timeout reached, then apply fails with surfaced status.
3. Given a supported mutable field (description) changes, when apply runs, then it is updated in place without replacement.
4. Given the store was deleted externally, when refresh/apply runs, then it is recreated and polled to readiness.

### Edge Cases
- Unsupported subtype → validation error.
- Readiness never reaches terminal state → timeout error.
- Invalid credentials → create/update fails with surfaced backend error message.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST create store with subtype-specific required arguments (minimum configuration surface).
- **FR-002**: System MUST read and expose type, owner, state, createdAt, updatedAt uniformly across subtypes.
- **FR-003**: System MUST poll `deltastream.sys."stores"` until state = `ready` (or terminal error) or timeout.
- **FR-004**: System MUST ignore not-found on delete.
- **FR-005**: System SHOULD update mutable metadata in place (e.g., description) without replacement.
- **FR-006**: System MUST validate subtype-specific argument presence and semantics.
- **FR-007**: All credential / secret fields MUST be tagged as Pulumi secrets in schema output.
- **FR-008**: Inline credential passing in CREATE / UPDATE is permitted initially; provider MUST surface clear error on invalid credentials.
- **FR-009**: Provider MUST surface backend/store errors (including status message) directly to the user in diagnostics.
- **FR-010**: Integration tests MUST cover create, update, delete, and invalid credential (configuration error) paths per subtype.
- **FR-011**: Readiness polling MUST emit actionable timeout error including last observed state if timeout occurs.

### Non-Functional Requirements
- **NFR-001**: Polling MUST have bounded timeout (uniform across subtypes unless future decision revises).
- **NFR-002**: Identifier & value quoting MUST prevent injection.
- **NFR-003**: Polling SHOULD not exceed one query per interval per store.
- **NFR-004**: Evolution MUST be additive-only; breaking changes require deprecation window & migration notes.
- **NFR-005**: Future secure credential transmission upgrade (properties.file + attachment) MUST avoid forcing replacements for unchanged logical values.
- **NFR-006**: Uniform output field set (FR-002) MUST remain stable across new subtypes.

### Key Entities *(include if data model involved)*
- **Store**: External integration endpoint (name, subtype, state, createdAt, updatedAt, metadata).

## 8. High-Level Design *(optional if trivial)*
Dispatch on subtype to build creation SQL; readiness poll executes catalog query for status until terminal.

## 9. Detailed Notes *(optional)*
Potential extension: backoff strategy; retries on transient errors.

## 10. Security & Privacy *(mandatory if handling sensitive data)*
Credentials are marked Pulumi secret fields (FR-007). Inline credential submission currently used; roadmap to enhanced secure file/attachment transport (NFR-005). No secret rotation logic included presently. Quoting utilities mitigate injection.

## 11. Performance & Scalability *(optional)*
Polling frequency balanced for responsiveness vs load; constant overhead per store.

## 12. Observability *(optional)*
Logs: create start/end, each poll attempt (debug). Future metrics: readiness_duration_seconds.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes catalog exposes store state field.
- Assumes subtypes share readiness concept.

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Long provisioning | Apply delays | Timeout & user feedback |
| Partial update failure | Drift | Re-query post-update & fail fast |

## 15. Alternatives Considered *(optional)*
- Separate resource per subtype → higher boilerplate; less reuse.

## 16. Open Questions *(mandatory until resolved)*
Resolved (documented as Decisions below):
- Subtype readiness semantics uniform via polling `deltastream.sys."stores"` (FR-003/011) — resolved.
- Credentials secret tagging & initial inline handling acceptable, future secure channel planned (FR-007/008, NFR-005) — resolved.
- Minimum Configuration Surface clarified (Decision D-003 below) — resolved.
- Validation rules subtype-specific & must be defined per addition using official CREATE STORE docs — resolved.
- Integration tests coverage requirements (FR-010) — resolved.
- Uniform output field set (FR-002) — resolved.
- Additive-only evolution (NFR-004) — reaffirmed.
- Quotas not applicable; removed from subtype checklist — resolved.
- Error surfacing requirement (FR-009) — resolved.

Outstanding:
*(none — all clarified in Decisions)*

### Decisions
| ID | Decision | Notes |
|----|----------|-------|
| D-001 | Uniform readiness polling on system stores table | Applies to all current & future subtypes |
| D-002 | Credentials marked secret; inline allowed short-term | Future secure transport will be additive |
| D-003 | Minimum Configuration Surface = name + essential connectivity/auth fields only | Prevent premature optional expansion |
| D-004 | Output fields standardized: type,state,createdAt,updatedAt,owner | Stability guarantee (NFR-006) |
| D-005 | Integration tests mandate create/update/delete/invalid creds | Gate for new subtype PRs |
| D-006 | No quota considerations (de-scoped) | Checklist item removed |
| D-007 | Error mapping surfaced directly with backend status_message | Actionable diagnostics |
| D-008 | Maintain single union-style `Store` resource with subtype discriminator (beta) | Matches backend & Terraform; revisit if DX/confusion or divergent lifecycles emerge; planned new subtypes: s3, kinesis, clickhouse, iceberg-rest, iceberg-glue |
| D-009 | No dev/ephemeral mode concept introduced | Single consistent lifecycle & readiness semantics; future ephemeral needs to be handled via separate tooling or distinct resource proposal if demand emerges |
| D-010 | Decline introducing generic catch-all subtype | Enforce enumerated vetted subtypes for safety, validation fidelity, and predictable readiness semantics; reconsider only if >3 external requests per quarter for unsupported backends |
| D-011 | Stick with inline credential parameters only (no credentialRef/authRef) | Simplicity & alignment with Pulumi secret marking; defer references unless rotation/indirection demand materializes |
| D-012 | Global fixed 10m readiness timeout (create/update/delete), 5s fixed poll interval | Simplicity & predictable UX; all lifecycle operations share limit; failure message includes subtype, name, elapsed, lastState, lastStatusMessage |

## 17. Rollout / Adoption Plan *(mandatory)*
Initial provider release with core subtypes; add new subtypes via minor versions.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All [NEEDS CLARIFICATION] resolved
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (readiness time, update behavior)

### Risk & Dependency Readiness
- [x] Critical risks have mitigations
- [x] Dependencies owned

## 19. Execution Status *(update as progresses)*
- [x] Drafted
- [x] Scenarios authored
- [x] Requirements finalized
- [x] Review completed (all open questions resolved; decisions D-001..D-012)
- [x] Accepted (Approved on 2025-09-05)

## 20. Appendix *(optional)*
Helpers: `lookupStore`, `waitForStoreReady`.
Reference Docs: https://docs.deltastream.io/reference/sql-syntax/ddl/create-store (validation source)
Future Secure Credential Path: properties.file + attachment (planned per NFR-005)
