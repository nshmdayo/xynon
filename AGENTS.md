# Agent Landing Page & Repository Index

Welcome, Agent. This file serves as your entry point and map for navigating this repository. To conserve context tokens and prevent errors, project requirements and technical designs have been split into modular files. 

Before executing tasks, read the specific files relevant to your current scope.

---

## Repository Documentation Map

### 1. Requirements & Features
If you are modifying, adding, or debugging features, refer to the individual Product Requirement Documents (PRDs):
*   **Proxy Engine:** `docs/requirements/PRD-proxy-engine.md`
*   **Hot-Reloading:** `docs/requirements/PRD-hot-reloading.md`
*   **WASM Plugins:** `docs/requirements/PRD-wasm-plugins.md`
*   **RPC Plugins:** `docs/requirements/PRD-rpc-plugins.md`
*   **CLI Tooling:** `docs/requirements/PRD-cli-tooling.md`
*   **Configuration:** `docs/requirements/PRD-configuration.md`

### 2. Architecture & Design
If you need to understand the codebase structure, system schemas, or system contracts:
*   **System Architecture:** `docs/architecture/system-design.md` — Core infrastructure layout, services, and execution models.
*   **API Specification:** `docs/architecture/api-spec.md` — Contract definitions, lifecycle actions, and ABI boundaries.
*   **Coding & Review Guidelines:** `docs/architecture/coding-guidelines.md` — Standards for code quality, security, and performance.

---

## Operational Instructions

1. **Context Management:** Do not read all documentation files simultaneously. Only read the specific file(s) listed above that map to your current issue or ticket.
2. **Updating Docs:** If your code changes alter an API layout or introduce a new feature, you are responsible for updating the corresponding file under `docs/` as part of your pull request. Do not append large blocks of text to this index file.


---

## Loop Engineering & Task Completion Conditions

When performing tasks, employ the **Loop Engineering** methodology: iteratively run tests, debug failures, and apply fixes autonomously. Do not prompt the user for help solely because a test fails.

### Definition of Done (DoD)
Before concluding any task and reporting back to the user, you MUST verify that the following conditions are met:
1. **Unit Tests Pass:** Run `make test` and ensure all Go tests pass without errors.
2. **E2E Tests Pass:** Run `make test-e2e` and ensure the end-to-end proxy and plugin integration tests pass.
3. **Documentation Updated:** Any new API endpoints, CLI flags, or architecture changes have been reflected in the corresponding `docs/` or `README.md` files.

### Agent Skills & Workflows
Before executing specific tasks, review the following instructions located in `.agents/skills/`:
*   **Feature Development:** `.agents/skills/feature/SKILL.md` — The 5-step loop methodology for adding new features.
*   **Testing:** `.agents/skills/testing/SKILL.md` — Instructions for running unit and E2E tests.
*   **Plugin Scaffold:** `.agents/skills/plugin_scaffold/SKILL.md` — Boilerplate templates for generating WASM and RPC plugins.
*   **Code Review:** `.agents/skills/code_review/SKILL.md` — Instructions for conducting codebase reviews according to project standards.
