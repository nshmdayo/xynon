---
name: feature
description: 新機能実装のためのプロセスに従い、実装にはOpus 4.6モデルを使用し、テストはtesting、レビューはcode_review、プラグイン作成はplugin_scaffoldスキルに委譲し、ドキュメントを更新します。作業を通じて得た新しい学びや知見は、`/learn`コマンドを推奨して記録します。
---
# Feature Implementation Skill

## Model Specification
- **Implementation & Test-Fix Loop (実装・テスト修正サイクル):** MUST be executed using the **Claude 4.6 Opus (Opus 4.6)** model.
- **Code Review (コードレビュー):** MUST be delegated to a subagent using the **Gemini Pro 3.1 (pro)** model.

## Execution Steps

1. **Implementation Plan (計画を立てる)**
   Before writing any code, analyze the requirements and generate an Implementation Plan. Detail the files to be modified, the architecture changes, and the exact steps you will take.
   
2. **Implementation (実装を行う)**
   Write the necessary code based on your Implementation Plan using the **Claude 4.6 Opus (Opus 4.6)** model.
   - **Plugin Scaffolding:** If the feature involves adding a new WASM or RPC plugin, you MUST follow the instructions in **`.agents/skills/plugin_scaffold/SKILL.md`** to generate the correct boilerplate code and update the plugin configuration.
   - **Documentation Update (ドキュメント更新):** If the feature introduces new specifications, API changes, CLI flags, or configuration options, you MUST update the corresponding documentation files under `docs/` (e.g., `docs/requirements/`, `docs/architecture/`) or `README.md` so that the docs accurately reflect the new state of the project.

3. **Test & Fix (テストと修正 / Loop Engineering)**
   Execute tests and the self-correcting loop (Loop Engineering) using the **Claude 4.6 Opus (Opus 4.6)** model according to **`.agents/skills/testing/SKILL.md`**.
   If tests fail, analyze logs and fix the code using Opus 4.6 until all unit and E2E tests pass cleanly.

4. **Code Review (コードレビュー)**
   Once all tests pass, launch a dedicated subagent via `invoke_subagent` to perform a comprehensive code review by following **`.agents/skills/code_review/SKILL.md`**:
   - **Role:** `"Code Reviewer"`
   - **Model:** `"pro"` (Gemini Pro 3.1)
   - **Prompt:** Direct the subagent to evaluate the code changes against the 7 perspectives in `.agents/skills/code_review/SKILL.md`.
   If any issues are found during the review, return to Step 2 (Implementation) or Step 3 (Test & Fix) to resolve them before proceeding.

5. **Completion (完了)**
   You may only consider the task "done" and report back to the user when **ALL** tests have passed without any errors, the code review process is fully satisfied, and all necessary documentation under `docs/` has been updated.

6. **Learn (学びの記録)**
   If you have learned new patterns, resolved complex setup issues, or acquired any valuable insights during the feature implementation process, recommend the user to use the `/learn` slash command to persist this behavior or knowledge for future tasks. If the learning process dictates creating a new Skill, you MUST use the **`.agents/skills/create_skills/SKILL.md`** skill to scaffold it properly.
