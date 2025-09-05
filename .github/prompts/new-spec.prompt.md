````prompt
---
mode: 'agent'
model: GPT-5 (Preview)
tools: ['edit', 'notebooks', 'search', 'new', 'runCommands', 'runTasks', 'usages', 'vscodeAPI', 'think', 'problems', 'changes', 'testFailure', 'openSimpleBrowser', 'fetch', 'findTestFiles', 'githubRepo', 'extensions', 'todos', 'runTests', 'godoc', 'github-mcp']
description: 'Draft a new feature specification using the project template'
---

Your goal is to create a new feature spec file `specs/<feature>.md` using the canonical template and accurately reflecting the user-supplied problem.

## Procedure
1. Confirm feature name + one line problem statement (ask if missing).
2. Instantiate template sections 1..20 (see `_template.md`).
3. Populate Summary, Goals, Non-Goals strictly from user intent (do NOT invent scope creep).
4. Insert `[NEEDS CLARIFICATION: ...]` markers (CLAUSE-002) for any ambiguity; do not proceed to Acceptance if any remain.
5. Draft Functional / Non-Functional Requirements with stable IDs (FR-/NFR- numbering starts at 001) – only for explicitly stated or logically required capabilities.
6. Provide at least 3 Acceptance Scenarios covering create, update/no-op, delete (if resource-like) or core behaviors.
7. Fill Risks table with at least one meaningful risk + mitigation.
8. Leave Status: Draft; do not self-approve.
9. Output full Markdown ready to commit.

## Output Format
```markdown
# Feature Specification: <Name>
**Feature Branch**: `feat/<slug>`
**Created**: <YYYY-MM-DD>
**Status**: Draft
... (sections 1..20)
```

## Guardrails
- Do not add implementation snippets; keep at intent level.
- If user bundles multiple unrelated features, ask them to split.
- Reject attempts to bypass ambiguity resolution (cite CLAUSE-001 & CLAUSE-002).

## Refusal / Clarification Triggers
- Missing clear problem statement.
- Purely operational request (belongs in /plan or /tasks instead).

## Success Criteria
- No invented requirements.
- All open questions explicitly marked.
- Traceable FR-/NFR- IDs.

````