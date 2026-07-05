# Agent Landing Page & Repository Index

Welcome, Agent. This file serves as your entry point and map for navigating this repository. To conserve context tokens and prevent errors, project requirements and technical designs have been split into modular files. 

Before executing tasks, read the specific files relevant to your current scope.

---

## Rpository Documentation Map

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

---

## Operational Instructions

1. **Context Management:** Do not read all documentation files simultaneously. Only read the specific file(s) listed above that map to your current issue or ticket.
2. **Updating Docs:** If your code changes alter an API layout or introduce a new feature, you are responsible for updating the corresponding file under `docs/` as part of your pull request. Do not append large blocks of text to this index file.
3. **Execution Context:** For framework-specific setup, linting rules, and deployment workflows, check the configuration files in `.github/agent-instructions/` if available.

