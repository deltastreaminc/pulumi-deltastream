# Implementation Plan: Build CI and Release GitHub Actions

**Branch**: `003-build-a-ci` | **Date**: September 7, 2025 | **Spec**: [/specs/003-build-a-ci/spec.md](/home/kraman/deltastream/pulumi-deltastream-3/specs/003-build-a-ci/spec.md)
**Input**: Feature specification from `/specs/003-build-a-ci/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
4. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
5. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, or `GEMINI.md` for Gemini CLI).
6. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
7. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
8. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
This feature implements GitHub Actions workflows for CI (Continuous Integration) and Release processes for the Pulumi Deltastream provider. The CI workflow will run builds for all pull requests and tests for trusted pull requests, ensuring code quality without exposing secrets to forked repositories. The Release workflow will automatically create and publish releases when triggered, generating artifacts for supported platforms (linux x86_64, linux arm64, darwin arm64) and publishing to appropriate channels (npm package for Node.js, git tag for Go).

## Technical Context
**Language/Version**: YAML (GitHub Actions workflow syntax), Go 1.24.x, Node.js 20.x  
**Primary Dependencies**: GitHub Actions, Pulumi SDK build tools (>=3.182.0), yarn (>=1.22.22)  
**Storage**: N/A (GitHub Actions uses repository for configuration)  
**Testing**: GitHub Actions workflow testing, integration testing with existing test suite  
**Target Platform**: GitHub Actions CI/CD platform  
**Project Type**: Infrastructure automation (GitHub Actions workflows)  
**Performance Goals**: CI workflow completes within 15 minutes, Release workflow within 30 minutes  
**Constraints**: Must not expose secrets to forked repositories, must follow GitHub Actions security best practices, macOS builds require code signing  
**Scale/Scope**: Two GitHub Actions workflows (CI and Release) with appropriate configuration

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Simplicity**:
- Projects: 0 (GitHub Actions workflows are configuration, not code projects)
- Using framework directly? Yes (GitHub Actions workflow syntax)
- Single data model? Yes (GitHub Actions workflow schema)
- Avoiding patterns? Yes (using standard GitHub Actions patterns)

**Architecture**:
- EVERY feature as library? N/A (feature is workflow configuration, not library code)
- Libraries listed: N/A
- CLI per library: N/A
- Library docs: N/A

**Testing (NON-NEGOTIABLE)**:
- RED-GREEN-Refactor cycle enforced? Yes (initial workflow will fail, then implement until passing)
- Git commits show tests before implementation? Yes (workflow configuration before secrets setup)
- Order: Contract→Integration→E2E→Unit strictly followed? Yes
- Real dependencies used? Yes (actual GitHub Actions environment)
- Integration tests for: new libraries, contract changes, shared schemas? N/A
- Deltastream system tables (`deltastream.sys.[object]`) used for status checks? N/A
- FORBIDDEN: Implementation before test, skipping RED phase - Will comply

**Observability**:
- Structured logging included? Yes (GitHub Actions provides built-in logging)
- Frontend logs → backend? N/A
- Error context sufficient? Yes (GitHub Actions provides detailed error logs)

**Versioning**:
- Version number assigned? N/A (workflows do not have version numbers)
- BUILD increments on every change? N/A
- Breaking changes handled? Yes (workflows will be backward compatible)

## Project Structure

### Documentation (this feature)
```
specs/003-build-a-ci/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
.github/
└── workflows/
    ├── ci.yml           # CI workflow for testing PRs
    └── release.yml      # Release workflow for creating and publishing releases
```

**Structure Decision**: GitHub Actions workflow files in the standard .github/workflows directory

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - GitHub Actions security best practices for forked repositories
   - Multi-platform artifact generation strategies
   - Node.js package publishing automation
   - Go release process requirements
   - Workflow trigger configuration options

2. **Generate and dispatch research agents**:
   ```
   Task: "Research GitHub Actions security best practices for forked repositories"
   Task: "Research artifact generation strategies for multiple platforms in GitHub Actions"
   Task: "Research automated npm package publishing in GitHub Actions"
   Task: "Research Go release process best practices"
   Task: "Research GitHub Actions workflow trigger configuration options"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all topics researched and documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Workflow entities (CI and Release)
   - Secret entities (GitHub Secrets)
   - GitHub Events (pull_request, push, workflow_dispatch)
   - Job structures and dependencies

2. **Generate API contracts** from functional requirements:
   - CI workflow contract with event triggers and job structure
   - Release workflow contract with event triggers and job structure
   - Security considerations for both workflows

3. **Generate contract tests** from contracts:
   - Test scenarios for CI workflow
   - Test scenarios for Release workflow
   - Security validation tests

4. **Extract test scenarios** from user stories:
   - Testing PR from main repository
   - Testing PR from forked repository
   - Creating a release with git tag
   - Manually triggering workflows

5. **Create quickstart guide** for user reference:
   - How to use CI workflow for PRs
   - How to create releases
   - Troubleshooting common issues

**Output**: 
- data-model.md: Documents the workflow entities and their properties
- contracts/ci-workflow.md: Contract for the CI workflow
- contracts/release-workflow.md: Contract for the Release workflow
- quickstart.md: Guide for using the workflows

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Tasks will include:
  - Setup GitHub Actions workflow directory structure
  - Create CI workflow file with required jobs and security measures
  - Create Release workflow file with platform matrix and publishing steps
  - Set up required GitHub secrets for npm publishing
  - Create test scripts for validating workflows
  - Add documentation for workflow usage

**Ordering Strategy**:
- Create directory structure first
- Set up basic workflows with job skeletons
- Implement security measures for forked repositories
- Add build and test functionality
- Set up release process and npm publishing
- Add documentation and quickstart guide

**Estimated Output**: 15-20 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

No constitutional violations identified. The implementation follows standard GitHub Actions patterns and does not introduce unnecessary complexity.

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

---
*Based on Constitution v2.3.0 - See `/memory/constitution.md`*