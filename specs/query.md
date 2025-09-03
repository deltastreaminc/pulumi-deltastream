# Feature Specification: Query Resource

**Feature Branch**: `feat/query-resource`  
**Created**: 2025-09-04  
**Status**: Accepted  
**Input / Source**: Reusable analytics definition requirement + immutability clarification (2025-09-04)

---
## ⚡ Quick Guidelines
Saved query definitions are IMMUTABLE after creation (no body edits); no scheduling, caching, or parameterization. Tagging deferred.

---
## 1. Summary *(mandatory)*
`Query` resource manages persisted, immutable SQL text within a namespace for reuse, governance, and dependency wiring.

## 2. Goals *(mandatory)*
- Persist reusable query text.
- Enforce immutability of query body after creation.
- Surface creation timestamp (and updated timestamp if storage layer sets it) for auditing.

## 3. Non-Goals *(mandatory)*
- Scheduling / execution orchestration.
- Result caching / materialization.
- Parameter binding system (explicitly out of scope; no future plans).
- Post-create SQL text modification.
- Hash-based optimization (not needed).
- Tagging (future backlog item, not designed here).

## 4. Background & Context *(mandatory)*
Ad-hoc unmanaged queries create duplication and shadow logic. Centralizing definitions improves discoverability and governance.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| Analyst | Stable saved queries | Losing ad-hoc SQL |
| Platform Engineer | Governed query assets | Shadow unmanaged queries |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As an analyst I want to declare a reusable query so that I (and others) can reference it consistently and trust it will not silently change.

### Acceptance Scenarios
1. Given a new query is declared, when apply runs, then the query row is created with body stored.
2. Given body text in code differs from stored body, when plan runs, then planning MUST fail with an immutability violation message instructing creation of a new query (rename) instead of mutation.
3. Given no changes, when apply runs, then plan is empty (no update).
4. Given the query is removed from code, when apply runs, then the row is deleted if present.

### Edge Cases
- Large body size → treated as normal (no optimization required).  
- Empty body → validation error.

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST create query with name + body.
- **FR-002**: System MUST reject any attempt to change body post-creation (plan phase error) to preserve immutability.
- **FR-003**: System MUST read and expose body + createdAt (+ updatedAt if storage emits it; updatedAt SHOULD equal createdAt due to immutability).
- **FR-004**: System MUST ignore not-found errors on delete.
- **FR-005**: System MUST treat identical desired body as no-op (no update operation emitted).

### Non-Functional Requirements
- **NFR-001**: String quoting MUST prevent injection.
- **NFR-002**: Provider MUST emit clear, actionable error messaging on immutability violation (includes stored vs desired body diff summary truncated to first N characters).

### Key Entities *(include if data model involved)*
- **Query**: (namespace, name, body, createdAt, updatedAt).

## 8. High-Level Design *(optional if trivial)*
Persist row in catalog with body text. During plan, fetch existing; if exists and body differs, raise error (no update or replacement) enforcing immutability.

## 9. Detailed Notes *(optional)*
Body comparison is direct string equality; no hashing or optimization required.

## 10. Security & Privacy *(mandatory if handling sensitive data)*
Body stored as plain text; no execution at creation; quoting prevents injection.

## 11. Dependencies & Assumptions *(mandatory)*
- Assumes namespace exists prior to query creation.
- Assumes catalog exposes created/updated timestamps.

## 12. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Attempted mutation | Inconsistent analytics | Immutability enforcement with explicit error |
| Injection attempts | Corruption | Strict quoting / no inline execution |

## 13. Open Questions *(mandatory until resolved)*
None (all prior questions resolved: parameter binding out of scope; immutability confirmed; hashing/tagging deferred).

## 14. Rollout / Adoption Plan *(mandatory)*
Initial provider release; future enhancement backlog (not specified here): tagging, ownership metadata.

## 15. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [x] No implementation details
- [x] Goals clear & measurable
- [x] All mandatory sections complete

### Requirements Completeness
- [x] All [NEEDS CLARIFICATION] resolved
- [x] Requirements testable & unambiguous
- [x] Success metrics defined (update detection, timestamps)

### Risk & Dependency Readiness
- [x] Critical risks have mitigations
- [x] Dependencies owned

## 16. Execution Status *(update as progresses)*
- [x] Drafted
- [x] Scenarios authored
- [x] Requirements finalized
- [x] Review completed
- [x] Accepted

## 17. Appendix *(optional)*
Helper: `lookupQuery`.
