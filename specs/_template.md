 # Feature Specification: <FEATURE NAME>

**Feature Branch**: `<issue-or-short-slug>`  
**Created**: <DATE>  
**Status**: Draft | Review | Accepted | Deferred  
**Input / Source**: <link to issue / discussion / prompt>

---
## ⚡ Quick Guidelines
✅ Focus on WHAT & WHY (user value)  
❌ Avoid HOW (implementation details, code, libraries)  
🔍 Mark uncertainties with `[NEEDS CLARIFICATION: question]`  
🧪 Every functional requirement must be testable  
✂ Remove optional sections if not relevant (do not leave "N/A")

---
## 1. Summary *(mandatory)*
One paragraph: problem framing + desired outcome + success signal.

## 2. Goals *(mandatory)*
Bullet list of measurable, user-centered goals.

## 3. Non-Goals *(mandatory)*
Clear exclusions to prevent scope creep.

## 4. Background & Context *(mandatory)*
Current state, constraints, related work, why now.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| <role> | | |

## 6. User Scenarios & Testing *(mandatory)*
### Primary User Story
As a <persona>, I want <capability> so that <value>.

### Acceptance Scenarios
1. **Given** <initial state>, **When** <action>, **Then** <expected result>
2. ...

### Edge Cases
- What happens when <boundary>?  
- How do we respond to <error case>?

## 7. Requirements *(mandatory)*
### Functional Requirements
- **FR-001**: System MUST ...
- **FR-002**: System MUST ...
- **FR-003**: System MUST ...
- **FR-004**: System MUST ...
- **FR-005**: System MUST ...
- **FR-006**: System MUST ... `[NEEDS CLARIFICATION: ...]`

### Non-Functional Requirements
- **NFR-001**: Performance / latency target ...
- **NFR-002**: Reliability / availability target ...
- **NFR-003**: Security requirement ...
- **NFR-004**: Observability requirement ...

### Key Entities *(include if data model involved)*
- **<EntityName>**: Purpose, key attributes (not implementation).

## 8. High-Level Design *(optional if trivial)*
Narrative of approach, boundaries, major flows (no code specifics).

## 9. Detailed Notes *(optional)*
Deeper clarifications still at the WHAT/behavior level (avoid algorithms / DB schema DDL).

## 10. Security & Privacy *(mandatory if handling sensitive data)*
Threat considerations, data access constraints, authZ boundaries.

## 11. Performance & Scalability *(optional)*
Expected scale, limits, constraints.

## 12. Observability *(optional)*
Logging, metrics, tracing strategy aligned to requirements.

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes ...
- Depends on ...

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| | | |

## 15. Alternatives Considered *(optional)*
- Option A – Pros / Cons
- Option B – Pros / Cons

## 16. Open Questions *(mandatory until resolved)*
- `[NEEDS CLARIFICATION: ...]`

## 17. Rollout / Adoption Plan *(mandatory)*
Phases, gating criteria, success metrics, migration or deprecation steps.

## 18. Review & Acceptance Checklist *(auto or manual)*
### Content Quality
- [ ] No implementation details
- [ ] Goals clear & measurable
- [ ] All mandatory sections complete

### Requirements Completeness
- [ ] All [NEEDS CLARIFICATION] resolved
- [ ] Requirements testable & unambiguous
- [ ] Success metrics defined

### Risk & Dependency Readiness
- [ ] Critical risks have mitigations
- [ ] External dependencies scheduled/owned

## 19. Execution Status *(update as progresses)*
- [ ] Drafted
- [ ] Scenarios authored
- [ ] Requirements finalized
- [ ] Review completed
- [ ] Accepted

## 20. Appendix *(optional)*
Links, diagrams, supporting analyses.
