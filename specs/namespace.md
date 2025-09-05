 # Feature Specification: Namespace Resource

**Feature Branch**: `feat/namespace-resource`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Foundational schema management

---
## ⚡ Quick Guidelines
Declarative schema (namespace) management within a database. No rename support.

---
## 1. Summary *(mandatory)*
Pulumi `Namespace` resource provides declarative creation and lifecycle management of schemas inside a Database, enabling logical grouping of objects and reducing manual SQL drift.

## 2. Goals *(mandatory)*
- Allow reproducible namespace creation across environments.
- Surface read metadata (owner, createdAt) for audit & dependency ordering.
- Provide safe, idempotent deletion path.

## 3. Non-Goals *(mandatory)*
- Schema rename / move operations.
- Bulk migration or data movement tooling.
- Ownership transfer semantics (future separate proposal).

## 4. Background & Context *(mandatory)*
Namespaces (schemas) logically partition objects. Current ad-hoc SQL causes inconsistencies. Integrating with Pulumi ensures state alignment and predictable provisioning.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| Platform Engineer | Repeatable namespace creation | Manual SQL |
| Data Engineer | Organized object placement | Ad-hoc scripts |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As a platform engineer I want to declare a namespace inside a database so that objects can be grouped predictably.

### Acceptance Scenarios
1. Given a database exists, when a new namespace is declared, then a schema is created and metadata returned.
2. Given the namespace exists unchanged, when apply runs, then the plan is empty.
3. Given the namespace was deleted externally, when refresh/apply runs, then it is recreated.
4. Given the namespace is removed from code, when apply runs, then the schema is dropped if present.

### Edge Cases
- Invalid identifier → clear validation error.  
- Delete of absent schema → no failure.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST create schema under specified database.
- **FR-002**: System MUST read and expose `owner` and `createdAt`.
- **FR-003**: System MUST treat database or name change as replacement.
- **FR-004**: System MUST ignore not-found errors on delete.
- **FR-005**: System SHOULD allow optional owner override context.

### Non-Functional Requirements
- **NFR-001**: Operations MUST be idempotent across applies.
- **NFR-002**: Identifiers MUST be safely quoted.

### Key Entities *(include if data model involved)*
- **Namespace**: Schema-scoped container (database, name, owner, createdAt).

## 8. High-Level Design *(optional if trivial)*
Thin wrapper over `CREATE SCHEMA`, catalog lookup, and `DROP SCHEMA` using shared helpers.

## 9. Detailed Notes *(optional)*
Composite ID `database/name` aids drift detection.

## 10. Security & Privacy *(mandatory if handling sensitive data)*
Identifier quoting prevents injection; role/org context restricts privileges.

## 11. Performance & Scalability *(optional)*
Constant time operations; one catalog query per read.

## 12. Observability *(optional)*
Debug logs on create/read/delete; future metric: schema_create_total.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes database exists prior to namespace creation.
- Assumes catalog exposes owner + timestamp.

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| External deletion mid-apply | Inconsistent plan | Re-lookup before finalize |
| Permission insufficiency | Apply failure | Clear surfaced error |

## 15. Alternatives Considered *(optional)*
- Implicit creation during object provisioning (less explicit control).

## 16. Open Questions *(mandatory until resolved)*
None at this time.

## 17. Rollout / Adoption Plan *(mandatory)*
Included in initial provider release; examples in Go/TS.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All [NEEDS CLARIFICATION] resolved
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (idempotency, metadata exposure)

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
Linked invokes: `GetNamespace`, `GetNamespaces`.
