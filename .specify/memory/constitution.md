<!--
SYNC IMPACT REPORT
==================
Version change: 2.3.0 → 2.3.1
Type: PATCH (clarification — added MCP tooling guidance under existing Principle II)

Modified principles:
  - II. Deltastream SQL as Source of Truth
    Added sub-section: "Deltastream Documentation MCP"
    Mandates use of the `deltastream-docs` MCP server for all DeltaStream
    documentation lookups (SQL syntax, sys tables, core concepts).

Added sections: none
Removed sections: none

Templates requiring updates:
  ✅ plan-template.md — Constitution Check gates verified (no structural change needed)
  ✅ spec-template.md — No structural change needed
  ✅ tasks-template.md — No structural change needed

Follow-up TODOs: none
-->

# Pulumi Deltastream Provider Constitution



## Core Principles

### I. Idiomatic Go Provider
All provider, resource, and function code MUST be written in idiomatic Go, following the latest Go best practices. Provider logic is organized under a `provider/` directory, with resources and functions as separate files. Use the latest supported version of Go.

### II. Deltastream SQL as Source of Truth
All interactions with Deltastream MUST use SQL syntax as defined in the [Deltastream SQL Reference](https://docs.deltastream.io/reference/sql-syntax) and [Core Concepts](https://docs.deltastream.io/overview/core-concepts). No custom or undocumented SQL is permitted.

#### Deltastream Documentation MCP
When documentation about DeltaStream is needed — including SQL syntax, system table schemas, core concepts, or API references — the `deltastream-docs` MCP server MUST be used. Do not rely on cached or inferred knowledge; always query the MCP for authoritative, up-to-date documentation.

#### Deltastream System Tables
To query the status of Deltastream objects, the provider MUST use the `deltastream.sys.[object]` tables. These system tables provide metadata such as `owner`, `state`, `createdAt`, and `updatedAt` for each object type. The provider MUST poll these tables to ensure that objects are ready to use after creation or modification.

> **Note:** If the fields or schema for a specific `deltastream.sys.[object]` table are not documented, the implementer MUST query the `deltastream-docs` MCP for the available fields and their meanings before proceeding.

### III. Schema and SDK Generation
The Pulumi schema.json and SDKs for Node.js and Python are generated from Go source. Direct edits to generated code or schema.json are strictly forbidden. All changes MUST be made in the Go provider source files. Node.js and Python SDKs are build artifacts only and MUST NOT be edited directly.

### IV. Integration Test-First
All features and changes MUST be covered by integration tests. Unit tests are optional. No observability or metrics support is required unless explicitly required by a feature.
If an error occurs, the error message MUST provide enough information for the user to understand and fix the issue.

### V. Versioning and Migration
- All changes MUST follow semantic versioning (MAJOR.MINOR.PATCH).
- Breaking changes require a migration plan and MUST be documented in the constitution history and templates.
- The constitution version MUST be updated with each amendment, and all templates must reference the current version.

### VI. Simplicity and Complexity Justification
- Simplicity is required: avoid unnecessary abstractions, patterns, or projects.
- Any complexity or deviation from the simplest approach MUST be justified and documented in the plan and PR.
- The number of projects MUST be minimized (prefer a single project structure unless justified).

### VII. Documentation and Code Review
All code and features MUST be documented. Code review is mandatory for all changes. Adherence to this constitution is required for approval.


## Additional Constraints

- Use Makefile and scripts to automate build, test, and codegen steps.
- Provider configuration MUST be handled via Pulumi config blocks, not hardcoded values.
- All Deltastream core concepts MUST be implemented as resources or functions, following [Deltastream Core Concepts](https://docs.deltastream.io/overview/core-concepts).
- All DeltaStream documentation lookups MUST go through the `deltastream-docs` MCP server.


## Development Workflow

- All development MUST occur in Go source files under the provider directory.
- Generated files (schema.json, Node.js/Python SDKs) MUST NOT be edited directly.
- All changes require integration test coverage.
- Code review and documentation are required for all pull requests.
- When implementing or researching DeltaStream features, ALWAYS use the `deltastream-docs` MCP to look up SQL syntax, sys table schemas, and core concepts.
- All amendments to the constitution MUST follow the [Constitution Update Checklist](../memory/constitution_update_checklist.md) and ensure all templates are updated and versioned.


## Governance

- This constitution supersedes all other development practices for the Pulumi Deltastream provider.
- Amendments require documentation, approval, and a migration plan.
- All PRs and reviews MUST verify compliance with this constitution.
- Complexity MUST be justified and documented.
- All templates and documentation MUST be kept in sync with the constitution version and requirements.

**Version**: 2.3.1 | **Ratified**: 2025-09-07 | **Last Amended**: 2026-06-29
