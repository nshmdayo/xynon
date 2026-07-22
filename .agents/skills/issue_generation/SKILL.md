---
name: issue_generation
description: リポジトリの課題や問題点を洗い出し、`gh`コマンドを使用してGitHub Issuesを作成します。
---

# Issue Generation Skill

This Skill provides a workflow to analyze the codebase and documentation within the repository, identify potential bugs, technical debt, missing features, etc., and register them on GitHub as Issues either in bulk or individually.

## Prerequisites
- The `gh` (GitHub CLI) command is available and authentication is complete.
- Terminal commands can be executed using the `run_command` tool.
- The `grep_search` tool is available for searching source code.

## Execution Steps

### 1. Identify Issues (Analyze)
Investigate the specified files, directories, or the entire repository to list issues that need improvement.
Please refer to the following perspectives:
- `TODO` and `FIXME` comments in the source code (utilize the `grep_search` tool)
- Performance bottlenecks or code requiring refactoring
- Inadequate error handling or unaddressed edge cases
- Areas not covered by unit tests (`make test`) or E2E tests (`make test-e2e`)
- Incomplete or outdated descriptions in the README or architecture documents (under `docs/`)

### 2. Issue Design (Design)
Design a title and body to register as an Issue for each identified problem.
It is also effective to have the user verify the contents beforehand (prompt for review).
Make sure to include the following elements in the body:
- **Summary**: What the problem is, what features/improvements are needed
- **Details / Steps to Reproduce**: Under what conditions the problem occurs, or specific implementation requirements
- **Task List**: Steps required for resolution (`- [ ] task description`)
- **Target Files**: Related file paths (with links)

### 3. Create GitHub Issue (Create)
**IMPORTANT: You MUST require explicit user approval immediately before executing `gh issue create`.**
Use the `run_command` tool to execute the `gh issue create` command.
To safely pass a multi-line body on the terminal, use the `--body-file` option or use a heredoc in the command line as shown in the example below.

**Example Command Execution:**
```bash
gh issue create --title "Concise Issue Title" --body "$(cat << 'EOF'
## Summary
Describe the summary of the issue here.

## Details / Tasks
- [ ] Add error handling
- [ ] Add tests
EOF
)" --label "bug"
```
* Use `--label`, `--assignee`, and `--milestone` options as necessary.

### 4. Completion and Reporting (Report)
Collect the URLs of the created Issues output as the result of the command execution, report the list of created Issues to the user, and complete the process.
