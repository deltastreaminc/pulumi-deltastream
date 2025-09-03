 # Feature Specification: Database Resource

**Feature Branch**: `feat/database-resource`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Internal provider foundational scope

---
## ⚡ Quick Guidelines
Declarative management of logical databases. No renames. Focus on safe creation & idempotent deletion.

---
## 1. Summary *(mandatory)*
Provide a Pulumi resource `Database` enabling declarative creation and lifecycle management of DeltaStream logical databases with owner scoping and drift detection.

## 2. Goals *(mandatory)*
- Enable IaC-managed database provisioning.
- Expose read attributes (owner, createdAt) for auditing & dependencies.
- Ensure deletes are safe & idempotent.

## 3. Non-Goals *(mandatory)*
- Database rename / transfer flows.
- Cross-organization migrations.
- Ownership mutation beyond initial context.
 - Database labels / metadata tagging (future separate proposal).

## 4. Background & Context *(mandatory)*
Previously, operators executed ad-hoc SQL to create databases leading to drift and missing audit context. This resource encapsulates creation, read introspection, and deletion using shared provider utilities.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| Platform Engineer | Consistent database provisioning | Manual SQL scripts |
| DevOps | Automated env lifecycle | Drift & missing metadata |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As a platform engineer I want to declare a database so that my environment provisioning is reproducible.

### Acceptance Scenarios
1. Given a stack apply, when the database does not exist, then a new database is created and its owner & createdAt are retrievable.
2. Given the database exists unchanged, when rerun, then no changes are applied (idempotent plan).
3. Given the database was manually removed, when refresh/apply runs, then the resource is recreated.
4. Given the database resource is removed from code, when apply runs, then the database is dropped if present.

### Edge Cases
- Apply with invalid name → Error with clear message.  
- Delete when already absent → No failure.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST create database with specified name using correct role/org context.
- **FR-002**: System MUST read and expose `owner` and `createdAt` attributes.
- **FR-003**: System MUST treat name change as replacement.
- **FR-004**: System MUST ignore not-found errors during delete.
- **FR-005**: System MUST error on invalid / unsafe identifier input.

### Non-Functional Requirements
- **NFR-001**: Typical create/delete operations SHOULD complete <5s.
- **NFR-002**: All SQL identifiers MUST be safely quoted.
- **NFR-003**: Operations MUST be idempotent across consecutive applies.

### Key Entities *(include if data model involved)*
- **Database**: Logical namespace container (name, owner, createdAt).

## 8. High-Level Design *(optional if trivial)*
Thin wrapper around SQL `CREATE DATABASE` and catalog lookup leveraging shared connection & quoting helpers.

## 9. Detailed Notes *(optional)*
Delete path tolerates absence; read path is authoritative for diff.

## 10. Security & Privacy *(mandatory if handling sensitive data)*
No sensitive data stored; enforces quoting to prevent injection. Role scoping ensures least privilege.

## 11. Performance & Scalability *(optional)*
Single catalog query per read; constant time relative to number of databases.

## 12. Observability *(optional)*
Debug logs on create & delete; potential future metrics: creation count, failures.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes catalog exposes owner & created timestamp.
- Assumes role context valid in provider config.

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Incorrect quoting | Security exposure | Central helpers `quoteIdent` |
| Role lacks privilege | Apply failure | Clear surfaced error |

## 15. Alternatives Considered *(optional)*
- Inline SQL in user code (higher duplication).
- Custom script outside Pulumi (lacks state integration).

## 16. Open Questions *(mandatory until resolved)*
None at this time (labels explicitly designated non-goal for this iteration).

## 17. Rollout / Adoption Plan *(mandatory)*
Included in initial provider release with examples (Go, TS). Future enhancements add labels once semantics agreed.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All [NEEDS CLARIFICATION] resolved
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (idempotency, attribute exposure)

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
Related invokes: `GetDatabase`, `GetDatabases`.
