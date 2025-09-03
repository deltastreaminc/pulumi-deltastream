````prompt
---
mode: 'agent'
model: GPT-5 (Preview)
tools: ['todos', 'think', 'search']
description: 'Generate a structured task list mapped to spec & governance IDs'
---

Generate or update a TODO list mapped to spec requirement and/or governance clause IDs.

## Procedure
1. Determine scope (ask if unclear): repo-wide | spec:<name> | area:<domain>.
2. Enumerate tasks grouped by: Blockers, Ready, Needs Clarification, Deferred.
3. Each task: imperative verb + concise object (<120 chars) + mapping `(FR-XXX)` / `(NFR-XXX)` / `(CLAUSE-XXX)`.
4. Prefix blockers with `BLOCKER:`; must relate to unmet FR/NFR or violated clause.
5. If an Accepted spec contains `[NEEDS CLARIFICATION]`, emit a blocker citing CLAUSE-002.
6. Avoid duplicates; consolidate overlapping tasks.
7. Provide final coverage list of referenced requirement IDs (optional if multi-spec scope).

## Output Format
```
Tasks (Scope: <scope>)
Blockers
- [ ] BLOCKER: <desc> (FR-00X)
Ready
- [ ] <desc> (FR-00Y)
Needs Clarification
- [ ] <desc> [NEEDS CLARIFICATION] (FR-00Z)
Deferred
- [ ] <desc> (NFR-00A)
Coverage
- FR-001, FR-002, NFR-001
```

## Guardrails
- No AND/OR inside single task; split instead.
- Do not include tasks without an ID mapping.
- Defer speculative ideas (mark Deferred) rather than mixing into Ready.

## Refusal / Clarification Triggers
- Missing or ambiguous scope after one clarification attempt.
- User asks for status unrelated to specs/governance.

## Success Criteria
- All blockers precede other tasks.
- Each requirement ID appears in at least one task (for single-spec scope).
- No unresolved ambiguity left unmarked.

````