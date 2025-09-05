 # Feature Specification: DeltaStream Object Resource

**Feature Branch**: `feat/object-resource`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Core data modeling requirement

---
## ⚡ Quick Guidelines
Logical data objects (tables/streams) scoped to namespace + database; immutable structural attributes for now.

---
## 1. Summary *(mandatory)*
`DeltaStreamObject` resource declaratively manages creation and lifecycle of logical data objects inside a namespace, referencing a backing store where applicable and exposing metadata for downstream queries.

## 2. Goals *(mandatory)*
- Provide declarative object creation within namespace.
- Expose object type, columns, store linkage.
- Ensure safe idempotent deletion.

## 3. Non-Goals *(mandatory)*
- In-place schema/column evolution (NOT planned; any change requires replacement).
- Data migration between stores (achieved via queries; outside object scope).

## 4. Background & Context *(mandatory)*
Manual DDL leads to inconsistent naming and drift. Consolidating object definition in Pulumi improves reproducibility and dependency graph clarity.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| Data Engineer | Declarative object creation | Manual DDL |
| Analyst | Stable references for queries | Naming inconsistency |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As a data engineer I want to declare a namespaced object so that pipelines can reliably reference it.

### Acceptance Scenarios
1. Given a namespace exists, when an object is declared, then it is created with fully-qualified name and metadata returned.
2. Given the object exists unchanged, when apply runs, then no diff occurs.
3. Given the object is removed from code, when apply runs, then it is dropped if present.
4. Given an attempt to change an immutable column set, when apply runs, then a replace is planned (or error surfaced—TBD decision).

### Edge Cases
- Invalid identifier characters → validation error.  
- Delete of missing object → no failure.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST create object with fully-qualified name database.namespace.object.
- **FR-002**: System MUST read and expose type, columns (if available), and store link.
- **FR-003**: System MUST perform idempotent delete (ignore not-found).
- **FR-004**: System SHOULD support import by name (adopt existing object).
- **FR-005**: System MUST treat structural changes as replacement until evolution supported.

### Non-Functional Requirements
- **NFR-001**: Reads MUST yield deterministic ordering of columns for stable diffs.
- **NFR-002**: Identifiers & strings MUST be safely quoted.

### Key Entities *(include if data model involved)*
- **Object**: (database, namespace, name, type, columns[], storeRef?).

## 8. High-Level Design *(optional if trivial)*
Execute CREATE statement, then introspect system catalog for metadata snapshot; delete via DROP IF EXISTS.

## 9. Detailed Notes *(optional)*
No planned drift detection enhancements (schema immutable; replacements handle structural changes).

## 10. Security & Privacy *(mandatory if handling sensitive data)*
No sensitive data stored; quoting prevents injection; namespace/database scopes apply least privilege context.

## 11. Performance & Scalability *(optional)*
Not performance sensitive (apply-time operations only); no explicit performance targets.

## 12. Observability *(optional)*
Debug logging for create/delete; potential metric: object_create_total.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes namespace exists before object creation.
- Assumes catalog exposes object columns & type.

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Column ordering nondeterminism | Spurious diffs | Canonical sort of columns |
| Structural change partial failure | Drift | Plan replacement semantics |

## 15. Alternatives Considered *(optional)*
- Embed object ops inside store → breaks separation of concerns.

## 16. Open Questions *(mandatory until resolved)*
None at this time (schema evolution explicitly excluded; structural changes always trigger replacement).

## 17. Rollout / Adoption Plan *(mandatory)*
Initial provider release; future proposal required for column evolution & alter-in-place.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All [NEEDS CLARIFICATION] resolved
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (idempotency, deterministic diffs)

### Risk & Dependency Readiness
- [x] Critical risks have mitigations
- [x] Dependencies owned

## 19. Execution Status *(update as progresses)*
- [x] Drafted
- [x] Scenarios authored
- [x] Requirements finalized
- [x] Review completed
- [x] Accepted

## 20. Appendix *(optional)*
Helper: `lookupObject`.
