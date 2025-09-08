# Feature Specification: Pulumi Deltastream Provider SDK

**Feature Branch**: `001-initial-query-support`  
**Created**: 2025-09-07  
**Status**: Draft  
**Input**: User description: "build a pulumi sdk based pulumi plugin for deltastream. initially we will only build out the database, namespace, stores (kafka, postgres, snowflake), object and query resource and functions."

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

### Section Requirements
- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

---

## 1. Summary *(mandatory)*
The Pulumi Deltastream Provider allows users to manage Deltastream resources through infrastructure as code. This initial implementation focuses on essential Deltastream components: databases, namespaces, stores (Kafka, Postgres, Snowflake), objects, and queries, providing both resource lifecycle management and read-only data source functions.

## 2. Goals *(mandatory)*
- Enable users to provision and manage Deltastream resources via Pulumi
- Support complete lifecycle (create, read, update, delete) of core Deltastream resources
- Provide read-only data source functions to access existing resources
- Ensure proper dependency management between resources
- Generate SDKs for Node.js and Golang programming languages
- Maintain infrastructure state consistent with Deltastream platform

## 3. Non-Goals *(mandatory)*
- Supporting advanced Deltastream features beyond core resources in initial release
- Building CLI tools or direct API wrappers
- Implementing custom authentication mechanisms
- Supporting pagination for large result sets in initial release
- Automating migration of existing Deltastream resources
- In-place schema/column evolution for objects (any change requires replacement)
- Post-create SQL text modification for queries (immutability enforced)
- Parameter binding system for queries
- Scheduling or execution orchestration for queries
- Result caching or materialization for queries

## 4. Background & Context *(mandatory)*
Deltastream users need a way to manage their resources as infrastructure-as-code. Currently, they must use direct SQL statements or web interfaces, which don't integrate well with broader cloud infrastructure provisioning workflows. A Pulumi provider will enable users to define, version, and deploy Deltastream resources alongside their other cloud resources. The provider will enforce important constraints like query immutability and object structure validation to ensure consistent behavior and prevent data corruption.

## 5. Users & Personas *(mandatory)*
| Persona | Primary Need | Current Pain |
|---------|--------------|--------------|
| DevOps Engineer | Manage Deltastream resources as code | Manual SQL execution and web interface |
| Platform Engineer | Reference existing resources in new deployments | Manual configuration and duplication |
| Data Engineer | Provision data pipelines with consistent configurations | Maintaining configurations outside IaC |
| Cloud Architect | Integrate Deltastream with other cloud services | Disconnected provisioning workflows |
| Solution Engineer | Deploy consistent environments across dev/staging/prod | Manual configuration differences |
| Analyst | Stable saved queries with consistent references | Losing ad-hoc SQL and duplicating logic |

## 6. User Scenarios & Testing *(mandatory)*

### Primary User Story
As a DevOps Engineer, I want to manage Deltastream resources using Pulumi infrastructure-as-code, so that I can integrate Deltastream resources into my broader cloud infrastructure provisioning workflows and maintain everything as code.

### Acceptance Scenarios
#### Resource Management
1. **Given** a user has Pulumi CLI installed and configured, **When** they create a new Pulumi project and install the Deltastream provider, **Then** they can import the provider in their programming language of choice (Node.js or Golang).

2. **Given** a user has configured the Deltastream provider with valid credentials, **When** they create a Deltastream database resource in their Pulumi program, **Then** a database is provisioned in their Deltastream account when they run `pulumi up`.

3. **Given** a user has configured the Deltastream provider with valid credentials, **When** they create a Deltastream namespace resource within a database in their Pulumi program, **Then** a namespace is provisioned in their Deltastream account when they run `pulumi up`.

4. **Given** a user has configured the Deltastream provider with valid credentials, **When** they create a Deltastream store resource of Kafka, Postgres or Snowflake type with valid store credentials in their Pulumi program, **Then** a store is provisioned in their Deltastream account when they run `pulumi up`.

5. **Given** a user has created Deltastream resources with Pulumi, **When** they make changes to the resource configuration in their code, **Then** those changes are applied to the actual resources when they run `pulumi up`.

6. **Given** a user has created Deltastream resources with Pulumi, **When** they run `pulumi destroy`, **Then** the resources are properly cleaned up in their Deltastream account.

7. **Given** a user has defined Deltastream stores (Kafka, Postgres, Snowflake), **When** they reference those stores in other resources, **Then** the proper dependency relationships are maintained.

8. **Given** a namespace exists, **When** a Data Engineer declares an object with columns and store reference in their Pulumi program, **Then** the object is created with the fully-qualified name and metadata when they run `pulumi up`.

9. **Given** an existing object with no changes, **When** the user runs `pulumi up`, **Then** no diff occurs and no changes are made.

10. **Given** an attempt to change an object's immutable column set, **When** the user runs `pulumi up`, **Then** a replacement operation is planned (not an in-place update).

11. **Given** a namespace exists, **When** an Analyst declares a query with name and SQL body in their Pulumi program, **Then** the query is created with the body stored when they run `pulumi up`.

12. **Given** a query exists and the user attempts to change its body, **When** they run `pulumi preview`, **Then** planning fails with an immutability violation message instructing creation of a new query instead.

#### Invoke Functions (Read-only Operations)
13. **Given** a database exists in Deltastream, **When** the user calls `getDatabase` with the database name, **Then** the matching database attributes are returned without creating a managed resource.

14. **Given** multiple namespaces exist in a database, **When** the user calls `getNamespaces` with a database filter, **Then** only the matching namespaces are returned in a deterministic order.

15. **Given** a user needs to reference an existing store, **When** they call `getStore` with the store name and type, **Then** they receive the store's attributes to use in other resources.

16. **Given** a user wants to discover existing objects, **When** they call `getObjects` with namespace filter, **Then** they receive a list of all objects in that namespace.

17. **Given** no matching object exists, **When** the user calls `getObject` with a non-existent name, **Then** a clear not-found error is returned.

18. **Given** a user wants to retrieve a saved query, **When** they call `getQuery` with the query name, **Then** they receive the query's body, creation timestamp, and other metadata.

### Edge Cases
- **Deltastream service errors during resource creation**: Report an error back to the user indicating the error message and error code provided by the Deltastream backend.
- **Permission or credential issues**: If there is a permission issue, the Deltastream backend will return an error with detail, which will be surfaced to the user with clear guidance on resolution.
- **Resource already exists with the same name**: The Deltastream backend will throw an error indicating the resource already exists. This will be surfaced to the user with appropriate context.
- **Conflicts between local state and remote state**: Follow standard Pulumi behavior of either updating the remote resource or causing the resource to be recreated depending on the nature of the conflict.
- **Timeouts during long-running operations**: Surface the timeout to the user. Pulumi programs are designed to be reentrant and will pick up from where they left off when re-executed.
- **Plural invoke function with no filters returning large result set**: Limit the result set to 200 records. If more exist, print a warning for the user indicating a partial result.
- **Singular invoke function matching multiple resources**: Objects are named uniquely in Deltastream, so this case should not happen as the backend would prevent it.
- **Filter yields >1 result for singular invoke function**: Return an explicit error with a meaningful message.
- **Empty filter set for plural invoke functions**: Return all results ordered deterministically, internally capped at 200 with a warning if truncated.
- **Invalid identifier characters in object names**: Return a validation error before sending to Deltastream.
- **Query with empty body**: Return a validation error.
- **Large query body size**: Treat as normal with no special optimization required.
- **Delete of missing object**: No failure (idempotent delete).

## 7. Requirements *(mandatory)*

### Functional Requirements
#### Resource Management
- **FR-001**: The provider MUST support creating, reading, updating, and deleting Deltastream database resources.
- **FR-002**: The provider MUST support creating, reading, updating, and deleting Deltastream namespace resources.
- **FR-003**: The provider MUST support creating, reading, updating, and deleting Deltastream store resources for Kafka, Postgres, and Snowflake types.
- **FR-004**: The provider MUST support creating, reading, updating, and deleting Deltastream object resources.
- **FR-005**: The provider MUST support executing Deltastream queries and managing query resources.
- **FR-006**: The provider MUST check resource status using Deltastream system tables before considering operations complete.
- **FR-007**: The provider MUST provide clear error messages when operations fail.
- **FR-008**: The provider MUST support configuration through standard Pulumi config mechanisms.
- **FR-009**: The provider MUST generate SDKs for Node.js and Golang.
- **FR-010**: The provider MUST support references between resources (e.g., linking a store to a database).
- **FR-011**: The provider MUST validate resource properties before sending requests to Deltastream.
- **FR-012**: The provider MUST gracefully handle rate limiting from the Deltastream API.

#### Object Resource
- **FR-013**: The provider MUST create objects with fully-qualified name (database.namespace.object).
- **FR-014**: The provider MUST read and expose object type, columns (if available), and store link.
- **FR-015**: The provider MUST perform idempotent delete for objects (ignore not-found errors).
- **FR-016**: The provider SHOULD support import of existing objects by name.
- **FR-017**: The provider MUST treat structural changes to objects as replacement operations.
- **FR-018**: Object reads MUST yield deterministic ordering of columns for stable diffs.

#### Query Resource
- **FR-019**: The provider MUST create queries with name and SQL body.
- **FR-020**: The provider MUST reject any attempt to change query body post-creation (plan phase error) to preserve immutability.
- **FR-021**: The provider MUST read and expose query body, createdAt timestamp, and updatedAt timestamp.
- **FR-022**: The provider MUST ignore not-found errors when deleting queries.
- **FR-023**: The provider MUST treat identical desired query body as no-op (no update operation emitted).
- **FR-024**: The provider MUST emit clear, actionable error messaging on query immutability violation, including diff summary.

#### Invoke Functions (Read-only Operations)
- **FR-025**: The provider MUST implement `getDatabase(s)` functions for looking up existing databases by name or filter.
- **FR-026**: The provider MUST implement `getNamespace(s)` functions for looking up existing namespaces filtered by database and/or name.
- **FR-027**: The provider MUST implement `getStore(s)` functions for looking up existing stores filtered by name and/or type.
- **FR-028**: The provider MUST implement `getObject(s)` functions for looking up existing objects filtered by namespace and/or name.
- **FR-029**: The provider SHOULD implement `getQuery/Queries` functions for looking up existing queries.
- **FR-030**: Singular lookup functions (e.g., `getDatabase`) MUST return a consistent not-found error when no matching resource exists.
- **FR-031**: Singular lookup functions MUST validate result cardinality and return an error if multiple matches are found.
- **FR-032**: Plural lookup functions (e.g., `getDatabases`) MUST return results in a deterministic order.
- **FR-033**: Plural lookup functions MUST apply an internal row cap (initially 200) and emit a warning if results are truncated.
- **FR-034**: All invoke functions MUST execute at most one SQL SELECT operation.
- **FR-035**: All invoke functions MUST properly quote identifiers and values to prevent SQL injection.

### Non-Functional Requirements
- **NFR-001**: Resource operations MUST be idempotent.
- **NFR-002**: The provider MUST be compatible with Pulumi CLI version 3.0 or higher.
- **NFR-003**: Error messages MUST be clear and actionable, providing users with guidance on how to resolve issues.
- **NFR-004**: The provider MUST include proper documentation for all resources and functions.
- **NFR-005**: Resource creation, update, and deletion operations MUST be atomic.
- **NFR-006**: The provider MUST gracefully handle disconnections and network issues during operations.
- **NFR-007**: The provider MUST use SQL as the primary interface for all Deltastream operations.

### Key Entities *(include if feature involves data)*
- **Database**: A Deltastream database that contains namespaces, stores, objects, and queries.
- **Namespace**: A logical grouping mechanism within a database for organizing Deltastream resources.
- **Store**: An external data source (Kafka, Postgres, Snowflake) that can be used by Deltastream.
- **Object**: A Deltastream object representing tables, views, or other data structures.
- **Query**: A Deltastream SQL query that can be executed or managed as a resource.
- **SystemTable**: Internal Deltastream tables that provide metadata about resources.
- **LookupResult**: Structured output from invoke functions containing resource attributes.

## 8. High-Level Design *(optional if trivial)*
The provider will implement resources and invoke functions as described in the requirements section. Resources will maintain CRUD operations via SQL statements, while invoke functions will use read-only SQL queries for data retrieval. All interactions will follow Deltastream SQL syntax and use system tables to track resource status.

## 9. Detailed Notes *(optional)*
Future versions of the provider may include:
- Pagination support for invoke functions to handle large result sets
- Support for additional Deltastream resource types
- Advanced filtering options for invoke functions
- Performance optimizations for large-scale deployments
- Resource import capabilities for existing Deltastream resources

## 10. Security & Privacy *(mandatory if handling sensitive data)*
- The provider will handle sensitive connection credentials (database passwords, API keys, etc.)
- All credentials will be handled according to Pulumi's secure configuration practices
- SQL injection prevention through proper quoting of identifiers and values
- No sensitive data will be exposed in logs or error messages
- Read-only invoke functions ensure no unintended mutations

## 11. Performance & Scalability *(optional)*
- Invoke functions limited to single SELECT operations to minimize latency
- Internal result caps (200 rows) for plural invoke functions until pagination is implemented
- Resources will use Deltastream system tables to efficiently check resource status
- Appropriate retry mechanisms for transient failures
- Efficient dependency management to minimize unnecessary operations

## 12. Observability *(optional)*
- Provider will log operations at appropriate detail levels
- Detailed error messages for troubleshooting
- Warnings for potentially problematic situations (e.g., truncated result sets)
- Trace logging for SQL statements (with sensitive data redacted)
- Structured output for easier parsing and analysis

## 13. Dependencies & Assumptions *(mandatory)*
- Assumes access to Deltastream with sufficient permissions
- Assumes Deltastream system tables are accessible for status checking
- Depends on Pulumi SDK for provider framework
- Assumes SQL interface is the primary method for interacting with Deltastream
- Assumes deterministic ordering can be achieved via ORDER BY clauses

## 14. Risks & Mitigations *(mandatory)*
| Risk | Impact | Mitigation |
|------|--------|------------|
| Deltastream API changes | Breaking functionality | Version compatibility checks, clear error messages |
| Long-running operations timing out | Failed deployments | Appropriate timeouts, retry logic, polling mechanisms |
| Large result sets in invoke functions | Memory pressure, slow response | Internal row cap (200) with warnings when truncated |
| SQL injection vulnerabilities | Security breach | Proper quoting of all identifiers and values |
| Inconsistent state between Pulumi and Deltastream | Deployment failures | System table validation, clear error messages, idempotent operations |
| Ambiguous lookup filters | User confusion | Clear documentation, precedence rules, validation |

## 15. Alternatives Considered *(optional)*
- Direct REST API wrapper: Rejected due to Deltastream's SQL-first interface
- Custom CLI tool: Rejected in favor of Pulumi's ecosystem integration
- Forcing resource import: Rejected as too heavy a workflow compared to invoke functions
- Separate packages for resources vs. invoke functions: Rejected for simplicity

## 16. Decisions & Closed Questions
### D-INV-001: Plural Invoke Function Scope & Safety Cap
For plural invoke functions (e.g., `getDatabases`, `getStores`, `getNamespaces`) with no filters, the provider will return all visible rows ordered deterministically by name. To prevent overly large responses before pagination support is implemented, an internal cap of 200 rows will be enforced using `LIMIT 201`. If more than 200 results exist, the result will be truncated to 200 rows and a warning will be emitted, suggesting narrower filters or future pagination support.

Rationale:
1. Matches common Pulumi data source patterns for easy discovery
2. Avoids premature introduction of pagination parameters
3. Provides bounded memory and serialization costs
4. Warning preserves transparency when truncation occurs

Future Evolution: Pagination parameters may be added in future releases, but until then, the internal cap provides a safety mechanism while maintaining backward compatibility.

### D-OBJ-001: Object Schema Immutability
The provider will treat any structural changes to objects (columns, types) as replacement operations rather than in-place updates. This ensures consistency and prevents potential data corruption. Users must explicitly create new objects with different names if they need to change structure.

Rationale:
1. Simplifies implementation in initial release
2. Prevents accidental data loss from schema changes
3. Enforces clear migration paths between object versions
4. Aligns with best practices for data schema evolution

Future Evolution: A future release may consider supporting in-place schema evolution with explicit opt-in and safeguards.

### D-QRY-001: Query Body Immutability
Queries are immutable after creation - any attempt to change a query's SQL body will be rejected during the planning phase with a clear error message. Users must create a new query with a different name if they need different SQL.

Rationale:
1. Prevents silent changes to critical business logic
2. Ensures analytical reproducibility
3. Maintains audit trail of query evolution
4. Enforces proper versioning of analytics definitions

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified
- [x] Deltastream system tables usage documented (if applicable)

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---
*Based on Constitution v2.3.0 - See `/memory/constitution.md`*
