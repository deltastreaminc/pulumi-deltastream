````prompt
---
mode: 'agent'
model: GPT-5 (Preview)
tools: ['edit', 'search', 'think', 'todos', 'runCommands', 'runTests', 'problems', 'changes']
description: 'Generate an implementation plan aligned with a spec'
---

Your goal is to produce a concise numbered implementation plan tied directly to spec requirement IDs.

## Procedure
1. Identify target spec (explicit name or infer from context). If ambiguous, ask first.
2. Extract FR-/NFR- IDs; ignore goals lacking concrete requirement mapping.
3. Produce 5–12 ordered tasks, each citing one or more FR-/NFR- ids.
4. Capture sequencing constraints inline (e.g., depends on #2).
5. Flag blockers caused by `[NEEDS CLARIFICATION]` (cite CLAUSE-002) — list them before normal tasks.
6. Exclude trivial boilerplate (e.g., "run tests", unless special).
7. Finish with a validation checklist enumerating all requirement IDs and coverage.

## Output Format
```
Plan: <spec name>
Blockers
1. BLOCKER: <desc> (FR-00X)
Core Tasks
1. <task desc> (FR-00Y, NFR-00Z)
...
Validation
- [ ] FR-001 covered by #2
- [ ] FR-002 covered by #3
...
```

## Guardrails
- Max 12 tasks (merge micro-steps).
- Do not invent requirements; cite only existing IDs.
- If zero actionable FRs, instruct user to refine spec.

## Refusal / Clarification Triggers
- No spec context.
- Spec still Draft with unresolved ambiguity blocking core behavior.

## Success Criteria
- Every FR/NFR appears exactly once in Validation.
- No orphan tasks lacking requirement mapping.
- Blockers clearly separated.

````