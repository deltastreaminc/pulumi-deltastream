 # Feature Specification: Invoke Functions (Get* Lookups)

**Feature Branch**: `feat/invokes`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Read-only composition need

---
## ⚡ Quick Guidelines
Lookup (data source) style functions; no mutation; consistent error semantics.

---
## 1. Summary *(mandatory)*
Set of read-only invoke functions fetching existing resources (databases, namespaces, stores, objects, queries) by filters or name without lifecycle management.

## 2. Goals *(mandatory)*
- Provide unified read-only accessors.
- Support singular + plural patterns.
- Expose rich attributes to wire dependencies.

## 3. Non-Goals *(mandatory)*
- Pagination (initial release).
- Any create/update/delete behavior.

## 4. Background & Context *(mandatory)*
Users need to reference existing infra components; invoking avoids importing state into managed resources.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| App Engineer | Reference existing db/namespace/object | Manual duplication |
| Auditor | Introspect current state | Direct DB access |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As an application engineer I want to fetch existing resources so that I can link new resources without redefining them.

### Acceptance Scenarios
1. Given a database exists, when `getDatabase` called with name, then matching attributes are returned.
2. Given multiple namespaces, when `getNamespaces` called with db filter, then only matching rows returned.
3. Given no object matches, when `getObject` called, then a not-found error is returned.
4. Given multiple stores of subtype, when `getStores` called with subtype filter, then list returned ordered deterministically.

### Edge Cases
- Filter yields >1 result for singular → explicit error.
- Empty filter set for plural → returns all (ordered, internally capped) per Decision D-INV-001.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST provide `getDatabase(s)` lookup by name or filter.
- **FR-002**: System MUST provide `getNamespace(s)` filtered by database and/or name.
- **FR-003**: System MUST provide `getStore(s)` filtered by name and subtype.
- **FR-004**: System MUST provide `getObject(s)` filtered by namespace and/or name.
- **FR-005**: System SHOULD provide `getQuery/Queries` (if applicable) with body & timestamps.
- **FR-006**: System MUST return consistent not-found error for singular lookups.
- **FR-007**: System MUST validate singular result cardinality (0 or >1 → error).

### Non-Functional Requirements
- **NFR-001**: Each invoke MUST execute at most one SQL SELECT.
- **NFR-002**: Identifier/value quoting MUST prevent injection.
- **NFR-003**: Results SHOULD be deterministic in ordering for plurals.
- **NFR-004**: Plural invokes MUST apply an internal row cap (initially 100) and emit a warning if truncated (see D-INV-001).

### Key Entities *(include if data model involved)*
- **Lookup Result**: Struct of resource attributes (mirrors resource read fields).

## 8. High-Level Design *(optional if trivial)*
Parameterized SQL construction using quoting helpers; scanning rows into typed outputs.

## 9. Detailed Notes *(optional)*
Potential future: pagination & limit parameters to control large result sets.

## 10. Security & Privacy *(mandatory if handling sensitive data)*
Read-only operations; quoting ensures no injection vulnerabilities.

## 11. Performance & Scalability *(optional)*
Single SELECT per call keeps latency minimal; large result sets may require pagination later.

## 12. Observability *(optional)*
Debug log includes filters & row counts.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes catalog tables accessible with read permissions.
- Assumes deterministic ordering achievable via ORDER BY name (if applied).

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Ambiguous filters | User confusion | Document filter precedence |
| Large unpaginated result | Memory overhead | Future pagination support |

## 15. Alternatives Considered *(optional)*
- Forcing resource import → heavier workflow.

## 16. Decisions & Closed Questions
### D-INV-001: Plural Default Scope & Safety Cap
Adopt permissive enumeration: when no filters are supplied to plural invokes (e.g., `getDatabases`, `getStores`, `getNamespaces`) the provider returns all visible rows ordered deterministically by name. To prevent pathological large responses prior to pagination support, an internal cap of 100 rows is enforced. Implementation executes `LIMIT 101` (cap + 1 sentinel) and if more than 100 are present the result list is truncated to 100 and a provider warning is emitted indicating truncation occurred and suggesting narrower filters or filing an issue for pagination.

Rationale:
1. Matches common Pulumi data source ergonomics (easy discovery).
2. Avoids premature introduction of pagination parameters.
3. Provides bounded memory & serialization cost via cap.
4. Warning preserves transparency when truncation changes completeness.

Future Evolution: Pagination parameters (`pageSize`, `nextToken`) may supersede the cap. Until then, raising the cap requires only updating constant; behavior remains backward compatible.

Open Questions: None.

## 17. Rollout / Adoption Plan *(mandatory)*
Initial release; pagination & advanced filtering added later.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All clarifications resolved (no open questions)
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (single SELECT, deterministic errors)

### Risk & Dependency Readiness
- [x] Critical risks have mitigations
- [x] Dependencies owned

## 19. Execution Status *(update as progresses)*
- [x] Drafted
- [x] Scenarios authored
- [x] Requirements finalized
- [x] Plural scope decision recorded (D-INV-001)
- [x] Review completed
- [x] Accepted

## 20. Appendix *(optional)*
Shared helpers: `quoteIdent`, `quoteString`.
