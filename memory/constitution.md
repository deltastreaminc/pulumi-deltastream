

# Pulumi Deltastream Provider Constitution




## Core Principles

### I. Idiomatic Go Provider
All provider, resource, and function code must be written in idiomatic Go, following the latest Go best practices. Provider logic is organized under a `provider/` directory, with resources and functions as separate files. Use the latest supported version of Go.

### II. Deltastream SQL as Source of Truth
All interactions with Deltastream must use SQL syntax as defined in the [Deltastream SQL Reference](https://docs.deltastream.io/reference/sql-syntax) and [Core Concepts](https://docs.deltastream.io/overview/core-concepts). No custom or undocumented SQL is permitted.

#### Deltastream System Tables
To query the status of Deltastream objects, the provider must use the `deltastream.sys.[object]` tables. These system tables provide metadata such as `owner`, `state`, `createdAt`, and `updatedAt` for each object type. The provider must poll these tables to ensure that objects are ready to use after creation or modification.

> **Note:** If the fields or schema for a specific `deltastream.sys.[object]` table are not documented, the implementer must query the user for the available fields and their meanings before proceeding.

### III. Schema and SDK Generation
The Pulumi schema.json and SDKs for Node.js and Python are generated from Go source. Direct edits to generated code or schema.json are strictly forbidden. All changes must be made in the Go provider source files. Node.js and Python SDKs are build artifacts only and must not be edited directly

### IV. Integration Test-First
All features and changes must be covered by integration tests. Unit tests are optional. No observability, or metrics support is required unless explicitly required by a feature.
If an error occurs, the error message must provide enough information for the user to understand and fix the issue.

### V. Versioning and Migration
- All changes must follow semantic versioning (MAJOR.MINOR.PATCH).
- Breaking changes require a migration plan and must be documented in the constitution history and templates.
- The constitution version must be updated with each amendment, and all templates must reference the current version.

### VI. Simplicity and Complexity Justification
- Simplicity is required: avoid unnecessary abstractions, patterns, or projects.
- Any complexity or deviation from the simplest approach must be justified and documented in the plan and PR.
- The number of projects should be minimized (prefer a single project structure unless justified).

### VII. Documentation and Code Review
All code and features must be documented. Code review is mandatory for all changes. Adherence to this constitution is required for approval.


## Additional Constraints

- Use Makefile and scripts to automate build, test, and codegen steps.
- Provider configuration must be handled via Pulumi config blocks, not hardcoded values.
- All Deltastream core concepts must be implemented as resources or functions, following [Deltastream Core Concepts](https://docs.deltastream.io/overview/core-concepts).


## Development Workflow

- All development must occur in Go source files under the provider directory.
- Generated files (schema.json, Node.js/Python SDKs) must not be edited directly.
- All changes require integration test coverage.
- Code review and documentation are required for all pull requests.
- All amendments to the constitution must follow the [Constitution Update Checklist](../memory/constitution_update_checklist.md) and ensure all templates are updated and versioned.


## Governance

- This constitution supersedes all other development practices for the Pulumi Deltastream provider.
- Amendments require documentation, approval, and a migration plan.
- All PRs and reviews must verify compliance with this constitution.
- Complexity must be justified and documented.
- All templates and documentation must be kept in sync with the constitution version and requirements.

**Version**: 2.3.0 | **Ratified**: 2025-09-07 | **Last Amended**: 2025-09-07