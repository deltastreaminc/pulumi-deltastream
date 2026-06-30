<!--
SYNC IMPACT REPORT
==================
Version change: 2.3.1 → 2.3.2
Type: PATCH (broken link fix, governance section expansion, sys tables link added,
       AGENTS.md referenced as runtime guidance file)

Modified principles:
  - II. Deltastream SQL as Source of Truth
    System Tables sub-section: added explicit link to
    https://docs.deltastream.io/reference/sql-syntax/systables

Added sections: none
Removed sections: none

Fixed:
  - Broken internal link ../memory/constitution_update_checklist.md
    → replaced with ../templates/checklist-template.md (file confirmed present)
  - Governance section: added versioning policy, amendment procedure detail,
    and runtime guidance file reference (AGENTS.md)

Templates requiring updates:
  ✅ plan-template.md — Constitution Check gates updated with project-specific rules
  ✅ spec-template.md — No structural change needed
  ✅ tasks-template.md — No structural change needed

Follow-up TODOs: none
Deferred items: none
-->

# Pulumi Deltastream Provider Constitution

## Core Principles

### I. Idiomatic Go Provider
All provider, resource, and function code MUST be written in idiomatic Go, following
the latest Go best practices. Provider logic is organized under a `provider/` directory,
with resources and functions as separate files. Use the latest supported version of Go.

### II. Deltastream SQL as Source of Truth
All interactions with Deltastream MUST use SQL syntax as defined in the
[Deltastream SQL Reference](https://docs.deltastream.io/reference/sql-syntax) and
[Core Concepts](https://docs.deltastream.io/overview/core-concepts).
No custom or undocumented SQL is permitted.

#### Deltastream Documentation MCP
When documentation about DeltaStream is needed — including SQL syntax, system table
schemas, core concepts, or API references — the `deltastream-docs` MCP server MUST be
used. Do not rely on cached or inferred knowledge; always query the MCP for
authoritative, up-to-date documentation.

#### Deltastream System Tables
To query the status of Deltastream objects, the provider MUST use the
`deltastream.sys.[object]` tables as documented in the
[System Tables Reference](https://docs.deltastream.io/reference/sql-syntax/systables).
These system tables provide metadata such as `owner`, `state`, `createdAt`, and
`updatedAt` for each object type. The provider MUST poll these tables to confirm that
objects are ready to use after creation or modification.

> **Note:** If the fields or schema for a specific `deltastream.sys.[object]` table are
> not known, the implementer MUST query the `deltastream-docs` MCP for the available
> fields and their meanings before proceeding.

### III. Schema and SDK Generation
The Pulumi `schema.json` and SDKs for Node.js and Python are generated from Go source.
Direct edits to generated code or `schema.json` are strictly forbidden. All changes
MUST be made in the Go provider source files. Node.js and Python SDKs are build
artifacts only and MUST NOT be edited directly.

### IV. Integration Test-First
All features and changes MUST be covered by integration tests. Unit tests are optional.
No observability or metrics support is required unless explicitly required by a feature.
If an error occurs, the error message MUST provide enough information for the user to
understand and fix the issue.

### V. Versioning and Migration
- All changes MUST follow semantic versioning (MAJOR.MINOR.PATCH).
- Breaking changes require a migration plan and MUST be documented in the constitution
  history and templates.
- The constitution version MUST be updated with each amendment, and all templates MUST
  reference the current version.

### VI. Simplicity and Complexity Justification
- Simplicity is required: avoid unnecessary abstractions, patterns, or projects.
- Any complexity or deviation from the simplest approach MUST be justified and
  documented in the plan and PR.
- The number of projects MUST be minimized (prefer a single project structure unless
  justified).

### VII. Documentation and Code Review
All code and features MUST be documented. Code review is mandatory for all changes.
Adherence to this constitution is required for PR approval.

## Additional Constraints

- Use Makefile and scripts to automate build, test, and codegen steps.
- Provider configuration MUST be handled via Pulumi config blocks, not hardcoded values.
- All Deltastream core concepts MUST be implemented as resources or functions, following
  [Deltastream Core Concepts](https://docs.deltastream.io/overview/core-concepts).
- All DeltaStream documentation lookups MUST go through the `deltastream-docs` MCP
  server.

## Development Workflow

- All development MUST occur in Go source files under the `provider/` directory.
- Generated files (`schema.json`, Node.js/Python SDKs) MUST NOT be edited directly.
- All changes require integration test coverage before merging.
- Code review and documentation are required for all pull requests.
- When implementing or researching DeltaStream features, ALWAYS use the
  `deltastream-docs` MCP to look up SQL syntax, sys table schemas, and core concepts.
- All amendments to the constitution MUST use the checklist in
  [checklist-template.md](../templates/checklist-template.md) to ensure all dependent
  documents and templates are updated and versioned consistently.
- For runtime development guidance, see [AGENTS.md](../../AGENTS.md).

## Governance

- This constitution supersedes all other development practices for the Pulumi
  Deltastream provider.
- **Amendment procedure**: Amendments require (1) a written rationale, (2) approval
  via PR review, (3) a migration plan for breaking changes, and (4) propagation of
  changes to all dependent templates using the checklist in
  [checklist-template.md](../templates/checklist-template.md).
- **Versioning policy**: Constitution versions follow semantic versioning:
  MAJOR for backward-incompatible principle removals or redefinitions; MINOR for new
  principles or materially expanded sections; PATCH for clarifications, link fixes, or
  wording improvements.
- **Compliance review**: All PRs and reviews MUST verify compliance with this
  constitution. Violations require justification documented in the PR description.
- Complexity MUST be justified and documented in the plan and PR.
- All templates and documentation MUST be kept in sync with the current constitution
  version.

**Version**: 2.3.2 | **Ratified**: 2025-09-07 | **Last Amended**: 2026-06-29
